package internal

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v5"

	"cosmossdk.io/log"
)

type RetryBackoffManager struct {
	lastCmd     string
	lastArgs    []string
	backoff     backoff.BackOff
	retryCount  int
	maxRestarts int
	logger      log.Logger
}

// NewRetryBackoffManager creates a new RetryBackoffManager instance.
func NewRetryBackoffManager(logger log.Logger, maxRestarts int) *RetryBackoffManager {
	backoffAlg := backoff.NewExponentialBackOff()
	return &RetryBackoffManager{
		backoff:     backoffAlg,
		maxRestarts: maxRestarts,
		logger:      logger,
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
		r.retryCount = 0
		r.lastCmd = cmd
		r.lastArgs = args
	} else {
		r.retryCount++
		if r.maxRestarts > 0 && r.retryCount >= r.maxRestarts {
			return backoff.Permanent(fmt.Errorf("maximum number of restarts reached: %d", r.maxRestarts))
		}
		// if the command and arguments are the same, we wait for the next backoff interval
		duration := r.backoff.NextBackOff()
		r.logger.Info("Applying backoff before restarting command",
			"backoff_duration", duration.String())
		time.Sleep(duration)
		r.logger.Info("Backoff time elapsed, restarting ")
	}
	return nil
}
