package testnet

import (
	"fmt"
	"time"

	"github.com/cometbft/cometbft/node"
)

// WaitForNodeHeight blocks until the node's consensus state reports
// a last height equal to or greater than desiredHeight.
// If totalWait has elapsed and the desired height has not been reached,
// an error is returned.
func WaitForNodeHeight(n *node.Node, desiredHeight int64, totalWait time.Duration) error {
	const backOff = 100 * time.Millisecond
	attempts := int64(totalWait / backOff)

	// In Comet 0.37, the consensus state was exposed directly on the Node.
	// As of 0.38, the node no longer exposes consensus state,
	// but the consensus state is available as a field on the RPC environment.
	//
	// Luckily, in 0.38 the RPC environment is no longer a package-level singleton,
	// so retrieving the RPC environment for a single node should be safe.
	env, err := n.ConfigureRPC()
	if err != nil {
		return fmt.Errorf("failed to configure RPC to reach into consensus state: %w", err)
	}

	curHeight := int64(-1)
	for i := int64(0); i < attempts; i++ {
		curHeight = env.ConsensusState.GetState().LastBlockHeight

		if curHeight < desiredHeight {
			time.Sleep(backOff)
			continue
		}

		// Met or exceeded target height.
		return nil
	}

	return fmt.Errorf("node did not reach desired height %d in %s; only reached height %d", desiredHeight, totalWait, curHeight)
}
