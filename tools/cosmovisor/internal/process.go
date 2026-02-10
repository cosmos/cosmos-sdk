package internal

import (
	"os/exec"
	"syscall"
	"time"
)

type ProcessRunner struct {
	cmd  *exec.Cmd
	done chan error // closed exactly once, after Wait returns
}

func RunProcess(cmd *exec.Cmd) (*ProcessRunner, error) {
	// start the process before returning a ProcessRunner
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	done := make(chan error, 1)
	go func() {
		// wait on the process to complete in a separate go routine
		done <- cmd.Wait()
		close(done)
	}()
	return &ProcessRunner{cmd: cmd, done: done}, nil
}

// Done returns the error that the process returned when it exited.
func (pr *ProcessRunner) Done() <-chan error {
	return pr.done
}

// Shutdown attempts to gracefully shut down the process by sending a SIGTERM signal.
// If the process does not exit within the specified grace period, it will be forcefully killed.
// An error will only be returned if there was an error shutting down the process.
// To get the error that the process itself returned, use Done().
func (pr *ProcessRunner) Shutdown(grace time.Duration) error {
	// check if already finished
	select {
	case <-pr.done:
		// already finished, nothing to do
		return nil
	default:
		// not finished yet, proceed with shutdown
	}

	proc := pr.cmd.Process
	if proc == nil {
		// this should only be true if the process has already exited
		<-pr.done // make sure Wait() has returned
		return nil
	}

	// signal shutdown
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	// wait for graceful exit or force-kill after timeout
	select {
	case <-pr.done:
	case <-time.After(grace):
		_ = proc.Kill()
		<-pr.done
	}
	return nil
}
