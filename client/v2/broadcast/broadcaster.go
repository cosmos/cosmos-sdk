package broadcast

import "context"

// Broadcaster defines an interface for broadcasting transactions to the consensus engine.
type Broadcaster interface {
	// Broadcast sends a transaction to the network and returns the result.
	//
	// It returns a byte slice containing the formatted result that will be
	// passed to the output writer, and an error if the broadcast failed.
	Broadcast(ctx context.Context, txBytes []byte) ([]byte, error)

	// Consensus returns the consensus engine identifier for this Broadcaster.
	Consensus() string
}
