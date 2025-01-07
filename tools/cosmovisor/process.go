package cosmovisor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/otiai10/copy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/log"
	"cosmossdk.io/x/upgrade/plan"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
)

type Launcher struct {
	logger log.Logger
	cfg    *Config
	fw     *fileWatcher
}

func NewLauncher(logger log.Logger, cfg *Config) (Launcher, error) {
	fw, err := newUpgradeFileWatcher(cfg)
	if err != nil {
		return Launcher{}, err
	}

	return Launcher{logger: logger, cfg: cfg, fw: fw}, nil
}

// loadBatchUpgradeFile loads the batch upgrade file into memory, sorted by
// their upgrade heights
func loadBatchUpgradeFile(cfg *Config) ([]upgradetypes.Plan, error) {
	var uInfos []upgradetypes.Plan
	upgradeInfoFile, err := os.ReadFile(cfg.UpgradeInfoBatchFilePath())
	if os.IsNotExist(err) {
		return uInfos, nil
	} else if err != nil {
		return nil, fmt.Errorf("error while reading %s: %w", cfg.UpgradeInfoBatchFilePath(), err)
	}

	if err = json.Unmarshal(upgradeInfoFile, &uInfos); err != nil {
		return nil, err
	}
	sort.Slice(uInfos, func(i, j int) bool {
		return uInfos[i].Height < uInfos[j].Height
	})
	return uInfos, nil
}

// BatchUpgradeWatcher starts a watcher loop that swaps upgrade manifests at the correct
// height, given the batch upgrade file. It watches the current state of the chain
// via the websocket API.
func BatchUpgradeWatcher(ctx context.Context, cfg *Config, logger log.Logger) {
	// load batch file in memory
	uInfos, err := loadBatchUpgradeFile(cfg)
	if err != nil {
		logger.Warn("failed to load batch upgrade file", "error", err)
		uInfos = []upgradetypes.Plan{}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Warn("failed to init watcher", "error", err)
		return
	}
	defer watcher.Close()
	err = watcher.Add(filepath.Dir(cfg.UpgradeInfoBatchFilePath()))
	if err != nil {
		logger.Warn("watcher failed to add upgrade directory", "error", err)
		return
	}

	var conn *grpc.ClientConn
	var grpcErr error

	defer func() {
		if conn != nil {
			if err := conn.Close(); err != nil {
				logger.Warn("couldn't stop gRPC client", "error", err)
			}
		}
	}()

	// Wait for the chain process to be ready
pollLoop:
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, grpcErr = grpc.NewClient(cfg.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if grpcErr == nil {
				break pollLoop
			}
			time.Sleep(time.Second)
		}
	}

	client := cmtservice.NewServiceClient(conn)

	var prevUpgradeHeight int64 = -1

	logger.Info("starting the batch watcher loop")
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				uInfos, err = loadBatchUpgradeFile(cfg)
				if err != nil {
					logger.Warn("failed to load batch upgrade file", "error", err)
					continue
				}
			}
		case <-ctx.Done():
			return
		default:
			if len(uInfos) == 0 {
				// prevent spending extra CPU cycles
				time.Sleep(time.Second)
				continue
			}
			resp, err := client.GetLatestBlock(ctx, &cmtservice.GetLatestBlockRequest{})
			if err != nil {
				logger.Warn("error getting latest block", "error", err)
				time.Sleep(time.Second)
				continue
			}

			h := resp.SdkBlock.Header.Height
			upcomingUpgrade := uInfos[0].Height
			// replace upgrade-info and upgrade-info batch file
			if h > prevUpgradeHeight && h < upcomingUpgrade {
				jsonBytes, err := json.Marshal(uInfos[0])
				if err != nil {
					logger.Warn("error marshaling JSON for upgrade-info.json", "error", err, "upgrade", uInfos[0])
					continue
				}
				if err := os.WriteFile(cfg.UpgradeInfoFilePath(), jsonBytes, 0o600); err != nil {
					logger.Warn("error writing upgrade-info.json", "error", err)
					continue
				}
				uInfos = uInfos[1:]

				jsonBytes, err = json.Marshal(uInfos)
				if err != nil {
					logger.Warn("error marshaling JSON for upgrade-info.json.batch", "error", err, "upgrades", uInfos)
					continue
				}
				if err := os.WriteFile(cfg.UpgradeInfoBatchFilePath(), jsonBytes, 0o600); err != nil {
					logger.Warn("error writing upgrade-info.json.batch", "error", err)
					// remove the upgrade-info.json.batch file to avoid non-deterministic behavior
					err := os.Remove(cfg.UpgradeInfoBatchFilePath())
					if err != nil && !os.IsNotExist(err) {
						logger.Warn("error removing upgrade-info.json.batch", "error", err)
						return
					}
					continue
				}
				prevUpgradeHeight = upcomingUpgrade
			}

			// Add a small delay to avoid hammering the gRPC endpoint
			time.Sleep(time.Second)
		}
	}
}

// Run launches the app in a subprocess and returns when the subprocess (app)
// exits (either when it dies, or *after* a successful upgrade.) and upgrade finished.
// Returns true if the upgrade request was detected and the upgrade process started.
func (l Launcher) Run(args []string, stdin io.Reader, stdout, stderr io.Writer) (bool, error) {
	bin, err := l.cfg.CurrentBin()
	if err != nil {
		return false, fmt.Errorf("error creating symlink to genesis: %w", err)
	}

	if err := plan.EnsureBinary(bin); err != nil {
		return false, fmt.Errorf("current binary is invalid: %w", err)
	}

	l.logger.Info("running app", "path", bin, "args", args)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("launching process %s %s failed: %w", bin, strings.Join(args, " "), err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		BatchUpgradeWatcher(ctx, l.cfg, l.logger)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		cancel()
		wg.Wait()
		if err := cmd.Process.Signal(sig); err != nil {
			l.logger.Error("terminated", "error", err, "bin", bin)
			os.Exit(1)
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

		if err := l.doCustomPreUpgrade(); err != nil {
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

	cancel()
	wg.Wait()

	return false, nil
}

// WaitForUpgradeOrExit checks upgrade plan file created by the app.
// When it returns, the process (app) is finished.
//
// It returns (true, nil) if an upgrade should be initiated (and we killed the process)
// It returns (false, err) if the process died by itself
// It returns (false, nil) if the process exited normally without triggering an upgrade. This is very unlikely
// to happen with "start" but may happen with short-lived commands like `simd genesis export ...`
func (l Launcher) WaitForUpgradeOrExit(cmd *exec.Cmd) (bool, error) {
	currentUpgrade, err := l.cfg.UpgradeInfo()
	if err != nil {
		// upgrade info not found do nothing
		currentUpgrade = upgradetypes.Plan{}
	}

	cmdDone := make(chan error)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	select {
	case <-l.fw.MonitorUpdate(currentUpgrade):
		// upgrade - kill the process and restart
		l.logger.Info("daemon shutting down in an attempt to restart")

		if l.cfg.ShutdownGrace > 0 {
			// Interrupt signal
			l.logger.Info("sent interrupt to app, waiting for exit")
			_ = cmd.Process.Signal(os.Interrupt)

			// Wait app exit
			psChan := make(chan *os.ProcessState)
			go func() {
				pstate, _ := cmd.Process.Wait()
				psChan <- pstate
			}()

			// Timeout and kill
			select {
			case <-psChan:
				// Normal Exit
				l.logger.Info("app exited normally")
			case <-time.After(l.cfg.ShutdownGrace):
				l.logger.Info("DAEMON_SHUTDOWN_GRACE exceeded, killing app")
				// Kill after grace period
				_ = cmd.Process.Kill()
			}
		} else {
			// Default: Immediate app kill
			_ = cmd.Process.Kill()
		}
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
		upgradeInfoFile, err := os.ReadFile(l.cfg.UpgradeInfoFilePath())
		if err != nil {
			return fmt.Errorf("error while reading upgrade-info.json: %w", err)
		}

		if err = json.Unmarshal(upgradeInfoFile, &uInfo); err != nil {
			return err
		}

		if uInfo.Name == "" {
			return errors.New("upgrade-info.json is empty")
		}

		// a destination directory, Format YYYY-MM-DD
		st := time.Now()
		ymd := fmt.Sprintf("%d-%d-%d", st.Year(), st.Month(), st.Day())
		dst := filepath.Join(l.cfg.DataBackupPath, fmt.Sprintf("data"+"-backup-%s", ymd))

		l.logger.Info("starting to take backup of data directory", "backup start time", st)

		// copy the $DAEMON_HOME/data to a backup dir
		if err = copy.Copy(filepath.Join(l.cfg.Home, "data"), dst); err != nil {
			return fmt.Errorf("error while taking data backup: %w", err)
		}

		// backup is done, lets check endtime to calculate total time taken for backup process
		et := time.Now()
		l.logger.Info("backup completed", "backup saved at", dst, "backup completion time", et, "time taken to complete backup", et.Sub(st))
	}

	return nil
}

// doCustomPreUpgrade executes the custom preupgrade script if provided.
func (l Launcher) doCustomPreUpgrade() error {
	if l.cfg.CustomPreUpgrade == "" {
		return nil
	}

	// check if upgrade-info.json is not empty.
	var upgradePlan upgradetypes.Plan
	upgradeInfoFile, err := os.ReadFile(l.cfg.UpgradeInfoFilePath())
	if err != nil {
		return fmt.Errorf("error while reading upgrade-info.json: %w", err)
	}

	if err = json.Unmarshal(upgradeInfoFile, &upgradePlan); err != nil {
		return err
	}

	if err = upgradePlan.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid upgrade plan: %w", err)
	}

	// check if preupgradeFile is executable file
	preupgradeFile := filepath.Join(l.cfg.Home, "cosmovisor", l.cfg.CustomPreUpgrade)
	l.logger.Info("looking for COSMOVISOR_CUSTOM_PREUPGRADE file", "file", preupgradeFile)
	info, err := os.Stat(preupgradeFile)
	if err != nil {
		l.logger.Error("COSMOVISOR_CUSTOM_PREUPGRADE file missing", "file", preupgradeFile)
		return err
	}
	if !info.Mode().IsRegular() {
		_, f := filepath.Split(preupgradeFile)
		return fmt.Errorf("COSMOVISOR_CUSTOM_PREUPGRADE: %s is not a regular file", f)
	}

	// Set the execute bit for only the current user
	// Given:  Current user - Group - Everyone
	//       0o     RWX     - RWX   - RWX
	oldMode := info.Mode().Perm()
	newMode := oldMode | 0o100
	if oldMode != newMode {
		if err := os.Chmod(preupgradeFile, newMode); err != nil {
			l.logger.Info("COSMOVISOR_CUSTOM_PREUPGRADE could not add execute permission")
			return errors.New("COSMOVISOR_CUSTOM_PREUPGRADE could not add execute permission")
		}
	}

	// Run preupgradeFile
	cmd := exec.Command(preupgradeFile, upgradePlan.Name, fmt.Sprintf("%d", upgradePlan.Height))
	cmd.Dir = l.cfg.Home
	result, err := cmd.Output()
	if err != nil {
		return err
	}

	l.logger.Info("COSMOVISOR_CUSTOM_PREUPGRADE result", "command", preupgradeFile, "argv1", upgradePlan.Name, "argv2", fmt.Sprintf("%d", upgradePlan.Height), "result", result)

	return nil
}

// doPreUpgrade runs the pre-upgrade command defined by the application and handles respective error codes.
// cfg contains the cosmovisor config from env var.
// doPreUpgrade runs the new APP binary in order to process the upgrade (post-upgrade for cosmovisor).
func (l *Launcher) doPreUpgrade() error {
	counter := 0
	for {
		if counter > l.cfg.PreUpgradeMaxRetries {
			return fmt.Errorf("pre-upgrade command failed. reached max attempt of retries - %d", l.cfg.PreUpgradeMaxRetries)
		}

		if err := l.executePreUpgradeCmd(); err != nil {
			counter++

			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				switch exitErr.ProcessState.ExitCode() {
				case 1:
					l.logger.Info("pre-upgrade command does not exist. continuing the upgrade.")
					return nil
				case 30:
					return fmt.Errorf("pre-upgrade command failed : %w", err)
				case 31:
					l.logger.Error("pre-upgrade command failed. retrying", "error", err, "attempt", counter)
					continue
				}
			}
		}

		l.logger.Info("pre-upgrade successful. continuing the upgrade.")
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

	l.logger.Info("pre-upgrade result", "result", result)
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
		if arg == fmt.Sprintf("--%s", FlagSkipUpgradeHeight) {
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
