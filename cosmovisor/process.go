package cosmovisor

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// LaunchProcess runs a subprocess and returns when the subprocess exits,
// either when it dies, or *after* a successful upgrade.
func LaunchProcess(cfg *Config, args []string, stdout, stderr io.Writer) (bool, error) {
	bin, err := cfg.CurrentBin()
	if err != nil {
		return false, fmt.Errorf("error creating symlink to genesis: %w", err)
	}

	if err := EnsureBinary(bin); err != nil {
		return false, fmt.Errorf("current binary invalid: %w", err)
	}

	cmd := exec.Command(bin, args...)
	outpipe, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}

	errpipe, err := cmd.StderrPipe()
	if err != nil {
		return false, err
	}

	scanOut := bufio.NewScanner(io.TeeReader(outpipe, stdout))
	scanErr := bufio.NewScanner(io.TeeReader(errpipe, stderr))

	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("launching process %s %s: %w", bin, strings.Join(args, " "), err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		if err := cmd.Process.Signal(sig); err != nil {
			log.Fatal(err)
		}
	}()

	// three ways to exit - command ends, find regexp in scanOut, find regexp in scanErr
	upgradeInfo, err := WaitForUpgradeOrExit(cmd, scanOut, scanErr)
	if err != nil {
		return false, err
	}

	if upgradeInfo != nil {
		return true, DoUpgrade(cfg, upgradeInfo)
	}

	return false, nil
}

// WaitResult is used to wrap feedback on cmd state with some mutex logic.
// This is needed as multiple go-routines can affect this - two read pipes that can trigger upgrade
// As well as the command, which can fail
type WaitResult struct {
	// both err and info may be updated from several go-routines
	// access is wrapped by mutex and should only be done through methods
	err   error
	info  *UpgradeInfo
	mutex sync.Mutex
}

// AsResult reads the data protected by mutex to avoid race conditions
func (u *WaitResult) AsResult() (*UpgradeInfo, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	return u.info, u.err
}

// SetError will set with the first error using a mutex
// don't set it once info is set, that means we chose to kill the process
func (u *WaitResult) SetError(myErr error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if u.info == nil && myErr != nil {
		u.err = myErr
	}
}

// SetUpgrade sets first non-nil upgrade info, ensure error is then nil
// pass in a command to shutdown on successful upgrade
func (u *WaitResult) SetUpgrade(up *UpgradeInfo) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if u.info == nil && up != nil {
		u.info = up
		u.err = nil
	}
}

// WaitForUpgradeOrExit listens to both output streams of the process, as well as the process state itself
// When it returns, the process is finished and all streams have closed.
//
// It returns (info, nil) if an upgrade should be initiated (and we killed the process)
// It returns (nil, err) if the process died by itself, or there was an issue reading the pipes
// It returns (nil, nil) if the process exited normally without triggering an upgrade. This is very unlikely
// to happened with "start" but may happened with short-lived commands like `gaiad export ...`
func WaitForUpgradeOrExit(cmd *exec.Cmd, scanOut, scanErr *bufio.Scanner) (*UpgradeInfo, error) {
	var res WaitResult

	waitScan := func(scan *bufio.Scanner) {
		upgrade, err := WaitForUpdate(scan)
		if err != nil {
			res.SetError(err)
		} else if upgrade != nil {
			res.SetUpgrade(upgrade)
			// now we need to kill the process
			_ = cmd.Process.Kill()
		}
	}

	// wait for the scanners, which can trigger upgrade and kill cmd
	go waitScan(scanOut)
	go waitScan(scanErr)

	// if the command exits normally (eg. short command like `gaiad version`), just return (nil, nil)
	// we often get broken read pipes if it runs too fast.
	// if we had upgrade info, we would have killed it, and thus got a non-nil error code
	err := cmd.Wait()
	if err == nil {
		return nil, nil
	}
	// this will set the error code if it wasn't killed due to upgrade
	res.SetError(err)
	return res.AsResult()
}
