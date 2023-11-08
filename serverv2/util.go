package serverv2

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"cosmossdk.io/log"
	"golang.org/x/sync/errgroup"
)

// ListenForQuitSignals listens for SIGINT and SIGTERM. When a signal is received,
// the cleanup function is called, indicating the caller can gracefully exit or
// return.
// The caller must ensure the corresponding context derived from the cancelFn is used correctly.
func ListenForQuitSignals(g *errgroup.Group, cancelFn context.CancelFunc, logger log.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	f := func() {
		sig := <-sigCh
		cancelFn()

		logger.Info("caught signal", "signal", sig.String())
	}

	g.Go(func() error {
		f()
		return nil
	})
}
