package broadcast

import (
	"context"
	"fmt"
)

const (
	// BroadcastSync defines a tx broadcasting mode where the client waits for
	// a CheckTx execution response only.
	BroadcastSync = "sync"
	// BroadcastAsync defines a tx broadcasting mode where the client returns
	// immediately.
	BroadcastAsync = "async"

	// cometBftConsensus is the identifier for the CometBFT consensus engine.
	cometBFTConsensus = "comet"
)

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
	NewBroadcasterFn func(url string, opts ...Option) (Broadcaster, error)

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

var _ BroadcasterFactory = Factory{}

// Factory is a factory for creating Broadcaster instances.
type Factory struct {
	engines map[string]NewBroadcasterFn
}

// Create creates a new Broadcaster based on the given consensus type.
func (f Factory) Create(_ context.Context, consensus, url string, opts ...Option) (Broadcaster, error) {
	creator, ok := f.engines[consensus]
	if !ok {
		return nil, fmt.Errorf("invalid consensus type: %s", consensus)
	}
	return creator(url, opts...)
}

// Register adds a new BroadcasterCreator for a given consensus type to the factory.
func (f Factory) Register(consensus string, creator NewBroadcasterFn) {
	f.engines[consensus] = creator
}

// NewFactory creates and returns a new Factory instance with a default CometBFT broadcaster.
func NewFactory() Factory {
	return Factory{
		engines: map[string]NewBroadcasterFn{
			cometBFTConsensus: func(url string, opts ...Option) (Broadcaster, error) {
				return NewCometBFTBroadcaster(url, opts...)
			},
		},
	}
}
