package cosmovisor

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

type Launcher struct {
	cfg *Config
	fw  *fileWatcher
}

func NewLauncher(cfg *Config) (Launcher, error) {
	fw, err := newUpgradeFileWatcher(cfg.UpgradeInfoFilePath(), cfg.PoolInterval)
	return Launcher{cfg, fw}, err
}

// Run a subprocess and returns when the subprocess exits,
// either when it dies, or *after* a successful upgrade.
func (l Launcher) Run(args []string, stdout, stderr io.Writer) (bool, error) {
	bin, err := l.cfg.CurrentBin()
	if err != nil {
		return false, fmt.Errorf("error creating symlink to genesis: %w", err)
	}

	if err := EnsureBinary(bin); err != nil {
		return false, fmt.Errorf("current binary is invalid: %w", err)
	}
	fmt.Println(">>> running", bin, args)
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
	currentUpgradeName := l.cfg.UpgradeName()
	fmt.Printf(">>>> current upgrade name: %q\n", currentUpgradeName)
	var cmdDone = make(chan error)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	select {
	case <-l.fw.MonitorUpdate(currentUpgradeName):
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
		if !l.fw.CheckUpdate(currentUpgradeName) {
			return false, err
		}
	}
	return true, nil
}
