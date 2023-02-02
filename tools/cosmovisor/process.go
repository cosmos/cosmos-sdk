package cosmovisor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/otiai10/copy"
	"github.com/rs/zerolog"

	"cosmossdk.io/x/upgrade/plan"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

type Launcher struct {
	logger *zerolog.Logger
	cfg    *Config
	fw     *fileWatcher
}

func NewLauncher(logger *zerolog.Logger, cfg *Config) (Launcher, error) {
	fw, err := newUpgradeFileWatcher(logger, cfg.UpgradeInfoFilePath(), cfg.PollInterval)
	if err != nil {
		return Launcher{}, err
	}

	return Launcher{logger: logger, cfg: cfg, fw: fw}, nil
}

// Run launches the app in a subprocess and returns when the subprocess (app)
// exits (either when it dies, or *after* a successful upgrade.) and upgrade finished.
// Returns true if the upgrade request was detected and the upgrade process started.
func (l Launcher) Run(args []string, stdout, stderr io.Writer) (bool, error) {
	bin, err := l.cfg.CurrentBin()
	if err != nil {
		return false, fmt.Errorf("error creating symlink to genesis: %w", err)
	}

	if err := plan.EnsureBinary(bin); err != nil {
		return false, fmt.Errorf("current binary is invalid: %w", err)
	}

	l.logger.Info().Str("path", bin).Strs("args", args).Msg("running app")
	cmd := exec.Command(bin, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("launching process %s %s failed: %w", bin, strings.Join(args, " "), err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		if err := cmd.Process.Signal(sig); err != nil {
			l.logger.Fatal().Err(err).Str("bin", bin).Msg("terminated")
		}
	}()

	if needsUpdate, err := l.WaitForUpgradeOrExit(cmd); err != nil || !needsUpdate {
		return false, err
	}

	if !IsSkipUpgradeHeight(args, l.fw.currentInfo) {
		l.cfg.WaitRestartDelay()

		if err := l.doBackup(); err != nil {
			return false, err
		}

		if err := UpgradeBinary(l.logger, l.cfg, l.fw.currentInfo); err != nil {
			return false, err
		}

		if err = l.doPreUpgrade(); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

// WaitForUpgradeOrExit checks upgrade plan file created by the app.
// When it returns, the process (app) is finished.
//
// It returns (true, nil) if an upgrade should be initiated (and we killed the process)
// It returns (false, err) if the process died by itself, or there was an issue reading the upgrade-info file.
// It returns (false, nil) if the process exited normally without triggering an upgrade. This is very unlikely
// to happened with "start" but may happened with short-lived commands like `gaiad export ...`
func (l Launcher) WaitForUpgradeOrExit(cmd *exec.Cmd) (bool, error) {
	currentUpgrade, err := l.cfg.UpgradeInfo()
	if err != nil {
		l.logger.Error().Err(err)
	}

	cmdDone := make(chan error)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	select {
	case <-l.fw.MonitorUpdate(currentUpgrade):
		// upgrade - kill the process and restart
		l.logger.Info().Msg("daemon shutting down in an attempt to restart")
		_ = cmd.Process.Kill()
	case err := <-cmdDone:
		l.fw.Stop()
		// no error -> command exits normally (eg. short command like `gaiad version`)
		if err == nil {
			return false, nil
		}
		// the app x/upgrade causes a panic and the app can die before the filwatcher finds the
		// update, so we need to recheck update-info file.
		if !l.fw.CheckUpdate(currentUpgrade) {
			return false, err
		}
	}
	return true, nil
}

func (l Launcher) doBackup() error {
	// take backup if `UNSAFE_SKIP_BACKUP` is not set.
	if !l.cfg.UnsafeSkipBackup {
		// check if upgrade-info.json is not empty.
		var uInfo upgradetypes.Plan
		upgradeInfoFile, err := os.ReadFile(filepath.Join(l.cfg.Home, "data", "upgrade-info.json"))
		if err != nil {
			return fmt.Errorf("error while reading upgrade-info.json: %w", err)
		}

		if err = json.Unmarshal(upgradeInfoFile, &uInfo); err != nil {
			return err
		}

		if uInfo.Name == "" {
			return fmt.Errorf("upgrade-info.json is empty")
		}

		// a destination directory, Format YYYY-MM-DD
		st := time.Now()
		stStr := fmt.Sprintf("%d-%d-%d", st.Year(), st.Month(), st.Day())
		dst := filepath.Join(l.cfg.DataBackupPath, fmt.Sprintf("data"+"-backup-%s", stStr))

		l.logger.Info().Time("backup start time", st).Msg("starting to take backup of data directory")

		// copy the $DAEMON_HOME/data to a backup dir
		if err = copy.Copy(filepath.Join(l.cfg.Home, "data"), dst); err != nil {
			return fmt.Errorf("error while taking data backup: %w", err)
		}

		// backup is done, lets check endtime to calculate total time taken for backup process
		et := time.Now()
		l.logger.Info().Str("backup saved at", dst).Time("backup completion time", et).TimeDiff("time taken to complete backup", et, st).Msg("backup completed")
	}

	return nil
}

// doPreUpgrade runs the pre-upgrade command defined by the application and handles respective error codes.
// cfg contains the cosmovisor config from env var.
// doPreUpgrade runs the new APP binary in order to process the upgrade (post-upgrade for cosmovisor).
func (l *Launcher) doPreUpgrade() error {
	counter := 0
	for {
		if counter > l.cfg.PreupgradeMaxRetries {
			return fmt.Errorf("pre-upgrade command failed. reached max attempt of retries - %d", l.cfg.PreupgradeMaxRetries)
		}

		if err := l.executePreUpgradeCmd(); err != nil {
			counter += 1

			switch err.(*exec.ExitError).ProcessState.ExitCode() {
			case 1:
				l.logger.Info().Msg("pre-upgrade command does not exist. continuing the upgrade.")
				return nil
			case 30:
				return fmt.Errorf("pre-upgrade command failed : %w", err)
			case 31:
				l.logger.Error().Err(err).Int("attempt", counter).Msg("pre-upgrade command failed. retrying")
				continue
			}
		}

		l.logger.Info().Msg("pre-upgrade successful. continuing the upgrade.")
		return nil
	}
}

// executePreUpgradeCmd runs the pre-upgrade command defined by the application
// cfg contains the cosmovisor config from the env vars
func (l *Launcher) executePreUpgradeCmd() error {
	bin, err := l.cfg.CurrentBin()
	if err != nil {
		return fmt.Errorf("error while getting current binary path: %w", err)
	}

	result, err := exec.Command(bin, "pre-upgrade").Output()
	if err != nil {
		return err
	}

	l.logger.Info().Bytes("result", result).Msg("pre-upgrade result")
	return nil
}

// IsSkipUpgradeHeight checks if pre-upgrade script must be run.
// If the height in the upgrade plan matches any of the heights provided in --unsafe-skip-upgrades, the script is not run.
func IsSkipUpgradeHeight(args []string, upgradeInfo upgradetypes.Plan) bool {
	skipUpgradeHeights := UpgradeSkipHeights(args)
	for _, h := range skipUpgradeHeights {
		if h == int(upgradeInfo.Height) {
			return true
		}
	}
	return false
}

// UpgradeSkipHeights gets all the heights provided when
// simd start --unsafe-skip-upgrades <height1> <optional_height_2> ... <optional_height_N>
func UpgradeSkipHeights(args []string) []int {
	var heights []int
	for i, arg := range args {
		if arg == "--unsafe-skip-upgrades" {
			j := i + 1

			for j < len(args) {
				tArg := args[j]
				if strings.HasPrefix(tArg, "-") {
					break
				}
				h, err := strconv.Atoi(tArg)
				if err == nil {
					heights = append(heights, h)
				}
				j++
			}

			break
		}
	}
	return heights
}
