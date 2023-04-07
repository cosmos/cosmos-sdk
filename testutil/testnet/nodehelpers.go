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
	const backoff = 100 * time.Millisecond
	attempts := int64(totalWait / backoff)

	curHeight := int64(-1)
	for i := int64(0); i < attempts; i++ {
		curHeight = n.ConsensusState().GetLastHeight()

		if curHeight < desiredHeight {
			time.Sleep(backoff)
			continue
		}

		// Met or exceeded target height.
		return nil
	}

	return fmt.Errorf("node did not reach desired height %d in %s; only reached height %d", desiredHeight, totalWait, curHeight)
}
