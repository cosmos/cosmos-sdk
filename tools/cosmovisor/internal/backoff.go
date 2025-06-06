package internal

import (
	"time"

	"github.com/cenkalti/backoff/v5"
)

type RetryBackoffManager struct {
	lastCmd  string
	lastArgs []string
	backoff  backoff.BackOff
}

// NewRetryBackoffManager creates a new RetryBackoffManager instance.
func NewRetryBackoffManager() *RetryBackoffManager {
	backoff := backoff.NewExponentialBackOff()
	return &RetryBackoffManager{
		backoff: backoff,
	}
}

func (r *RetryBackoffManager) BeforeRun(cmd string, args []string) error {
	reset := false
	// we reset the backoff if the command or its arguments have changed
	if r.lastCmd != cmd || len(r.lastArgs) != len(args) {
		reset = true
	} else {
		n := min(len(r.lastArgs), len(args))
		for i := 0; i < n; i++ {
			if r.lastArgs[i] != args[i] {
				reset = true
				break
			}
		}
	}
	if reset {
		// if the command or arguments have changed, we reset the backoff and store the new command and arguments
		r.backoff.Reset()
		r.lastCmd = cmd
		r.lastArgs = args
	} else {
		// if the command and arguments are the same, we wait for the next backoff interval
		time.Sleep(r.backoff.NextBackOff())
	}
	return nil
}
