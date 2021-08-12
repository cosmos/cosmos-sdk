package cosmovisor

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/otiai10/copy"
)

type Launcher struct {
	cfg *Config
	fw  *fileWatcher
}

func NewLauncher(cfg *Config) (Launcher, error) {
	fw, err := newUpgradeFileWatcher(cfg.UpgradeInfoFilePath(), cfg.PollInterval)
	return Launcher{cfg, fw}, err
}

// Run launches the app in a subprocess and returns when the subprocess (app)
// exits (either when it dies, or *after* a successful upgrade.) and upgrade finished.
// Returns true if the upgrade request was detected and the upgrade process started.
func (l Launcher) Run(args []string, stdout, stderr io.Writer) (bool, error) {
	bin, err := l.cfg.CurrentBin()
	if err != nil {
		return false, fmt.Errorf("error creating symlink to genesis: %w", err)
	}

	if err := EnsureBinary(bin); err != nil {
		return false, fmt.Errorf("current binary is invalid: %w", err)
	}
	fmt.Println("[cosmovisor] running ", bin, args)
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
			log.Fatal(bin, "terminated. Error:", err)
		}
	}()

	needsUpdate, err := l.WaitForUpgradeOrExit(cmd)
	if err != nil || !needsUpdate {
		return false, err
	}
	if err := doBackup(l.cfg); err != nil {
		return false, err
	}

	return true, DoUpgrade(l.cfg, l.fw.currentInfo)
}

// WaitForUpgradeOrExit checks upgrade plan file created by the app.
// When it returns, the process (app) is finished.
//
// It returns (true, nil) if an upgrade should be initiated (and we killed the process)
// It returns (false, err) if the process died by itself, or there was an issue reading the upgrade-info file.
// It returns (false, nil) if the process exited normally without triggering an upgrade. This is very unlikely
// to happened with "start" but may happened with short-lived commands like `gaiad export ...`
func (l Launcher) WaitForUpgradeOrExit(cmd *exec.Cmd) (bool, error) {
	currentUpgrade := l.cfg.UpgradeInfo()
	var cmdDone = make(chan error)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	select {
	case <-l.fw.MonitorUpdate(currentUpgrade):
		// upgrade - kill the process and restart
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

func doBackup(cfg *Config) error {
	// take backup if `UNSAFE_SKIP_BACKUP` is not set.
	if !cfg.UnsafeSkipBackup {
		// check if upgrade-info.json is not empty.
		var uInfo UpgradeInfo
		upgradeInfoFile, err := ioutil.ReadFile(filepath.Join(cfg.Home, "data", "upgrade-info.json"))
		if err != nil {
			return fmt.Errorf("error while reading upgrade-info.json: %w", err)
		}

		err = json.Unmarshal(upgradeInfoFile, &uInfo)
		if err != nil {
			return err
		}

		if uInfo.Name == "" {
			return fmt.Errorf("upgrade-info.json is empty")
		}

		// a destination directory, Format YYYY-MM-DD
		st := time.Now()
		stStr := fmt.Sprintf("%d-%d-%d", st.Year(), st.Month(), st.Day())
		dst := filepath.Join(cfg.Home, fmt.Sprintf("data"+"-backup-%s", stStr))

		fmt.Printf("starting to take backup of data directory at time %s", st)

		// copy the $DAEMON_HOME/data to a backup dir
		err = copy.Copy(filepath.Join(cfg.Home, "data"), dst)

		if err != nil {
			return fmt.Errorf("error while taking data backup: %w", err)
		}

		// backup is done, lets check endtime to calculate total time taken for backup process
		et := time.Now()
		timeTaken := et.Sub(st)
		fmt.Printf("backup saved at location: %s, completed at time: %s\n"+
			"time taken to complete the backup: %s", dst, et, timeTaken)
	}

	return nil
}
