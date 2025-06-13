package internal

import (
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type ProcessRunner struct {
	cmd  *exec.Cmd
	done chan error
	wg   *sync.WaitGroup
}

func RunProcess(cmd *exec.Cmd) *ProcessRunner {
	done := make(chan error, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cmd.Run()
		done <- err
		close(done)
	}()
	return &ProcessRunner{
		cmd:  cmd,
		done: done,
		wg:   wg,
	}
}

func (pr *ProcessRunner) Done() chan error {
	return pr.done
}

func (pr *ProcessRunner) Shutdown(shutdownGrace time.Duration) error {
	proc := pr.cmd.Process
	if proc == nil {
		return nil // process already exited
	}
	// TODO make sure we don't kill a process that is not running
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	done := make(chan struct{})
	go func() {
		pr.wg.Wait()
		close(done)
	}()
	go func() {
		select {
		case <-done:
		case <-time.After(shutdownGrace):
			// TODO handle the error
			proc.Kill()
		}
	}()
	return nil
}
