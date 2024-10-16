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
	cometBftConsensus = "comet"
)

type (
	// Broadcaster defines an interface for broadcasting transactions to the consensus engine.
	Broadcaster interface {
		// Broadcast sends a transaction to the network and returns the result.
		//
		// It returns a byte slice containing the formatted result that will be
		// passed to the output writer, and an error if the broadcast failed.
		Broadcast(ctx context.Context, txBytes []byte) ([]byte, error)
	}

	// factory defines a generic interface for creating a Broadcaster.
	factory interface {
		// Create creates a new Broadcaster instance of type T.
		create(ctx context.Context, consensus, url string, opts ...Option) (Broadcaster, error)
	}

	// Option is a function that configures a Broadcaster.
	Option func(Broadcaster)
)

var _ factory = broadcasterFactory{}

// broadcasterFactory is a factory for creating Broadcaster instances.
type broadcasterFactory struct{}

// create creates a new Broadcaster based on the given consensus type.
func (f broadcasterFactory) create(_ context.Context, consensus, url string, opts ...Option) (Broadcaster, error) {
	switch consensus {
	case cometBftConsensus:
		return NewCometBftBroadcaster(url, opts...)
	default:
		return nil, fmt.Errorf("invalid consensus type: %s", consensus)
	}
}
