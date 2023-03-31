package testnet

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/node"
)

// Nodes is a slice of comet nodes,
// with some additional convenience methods.
type Nodes []*node.Node

// Stop stops each node sequentially.
// All errors occurring during stop are returned as a joined error.
func (ns Nodes) Stop() error {
	var errs []error
	for i, n := range ns {
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
func (ns Nodes) Wait() {
	for _, n := range ns {
		n.Wait()
	}
}
