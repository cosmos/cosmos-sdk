package types

import "context"

type (
	// Broadcaster defines an interface for broadcasting transactions to the consensus engine.
	Broadcaster interface {
		// Broadcast sends a transaction to the network and returns the result.
		//
		// It returns a byte slice containing the formatted result that will be
		// passed to the output writer, and an error if the broadcast failed.
		Broadcast(ctx context.Context, txBytes []byte) ([]byte, error)

		// Consensus returns the consensus engine identifier for this Broadcaster.
		Consensus() string
	}

	// NewBroadcasterFn is a function type for creating Broadcaster instances.
	NewBroadcasterFn func(ctx context.Context, url string, opts ...Option) (Broadcaster, error)

	// BroadcasterFactory defines an interface for creating and registering Broadcasters.
	BroadcasterFactory interface {
		// Register adds a new BroadcasterCreator for a given consensus type to the factory.
		Register(consensus string, creator NewBroadcasterFn)
		// Create instantiates a new Broadcaster based on the given consensus type, URL, and options.
		// It returns the created Broadcaster and any error encountered during creation.
		Create(ctx context.Context, consensus, url string, opts ...Option) (Broadcaster, error)
	}

	// Option is a function that configures a Broadcaster.
	Option func(Broadcaster)
)
