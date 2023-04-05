package testnet

import (
	"errors"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/node"
)

// Nodes is a slice of comet nodes,
// with some additional convenience methods.
//
// Nodes may contain nil elements,
// so that a partially failed call to NewNetwork
// can still be properly cleaned up.
type Nodes []*node.Node

// Stop stops each node sequentially.
// All errors occurring during stop are returned as a joined error.
//
// Nil elements in ns are skipped.
func (ns Nodes) Stop() error {
	var errs []error
	for i, n := range ns {
		if n == nil {
			continue
		}
		if err := n.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop node %d: %w", i, err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// Wait blocks until every node in ns has completely stopped.
//
// Nil elements in ns are skipped.
func (ns Nodes) Wait() {
	for _, n := range ns {
		if n == nil {
			continue
		}
		n.Wait()
	}
}

// StopAndWait is shorthand for calling both Stop() and Wait(),
// useful as a deferred call in tests.
func (ns Nodes) StopAndWait() error {
	err := ns.Stop()
	ns.Wait()

	// TODO(mr): remove this sleep call after we are using a version of Comet
	// that includes a fix for https://github.com/cometbft/cometbft/issues/646.
	//
	// On my machine, this sleep appears to completely eliminate the late file write.
	// It also almost always works around https://github.com/cometbft/cometbft/issues/652.
	//
	// Presumably the fix for those two issues will be included in a v0.37.1 release.
	// If not, I assume they will be part of the first v0.38 series release.
	time.Sleep(250 * time.Millisecond)

	return err
}
