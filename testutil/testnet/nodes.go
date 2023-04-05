package testnet

import (
	"errors"
	"fmt"

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
	return err
}
