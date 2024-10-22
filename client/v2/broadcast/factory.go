package broadcast

import (
	"context"
	"fmt"

	"cosmossdk.io/client/v2/broadcast/types"
	"cosmossdk.io/client/v2/internal/comet"
)

var _ types.BroadcasterFactory = &Factory{}

// Factory is a factory for creating Broadcaster instances.
type Factory struct {
	engines map[string]types.NewBroadcasterFn
}

// Create creates a new Broadcaster based on the given consensus type.
func (f *Factory) Create(_ context.Context, consensus, url string, opts ...types.Option) (types.Broadcaster, error) {
	creator, ok := f.engines[consensus]
	if !ok {
		return nil, fmt.Errorf("invalid consensus type: %s", consensus)
	}
	return creator(url, opts...)
}

// Register adds a new BroadcasterCreator for a given consensus type to the factory.
func (f *Factory) Register(consensus string, creator types.NewBroadcasterFn) {
	f.engines[consensus] = creator
}

// NewFactory creates and returns a new Factory instance with a default CometBFT broadcaster.
func NewFactory() Factory {
	return Factory{
		engines: map[string]types.NewBroadcasterFn{
			"comet": func(url string, opts ...types.Option) (types.Broadcaster, error) {
				return comet.NewCometBFTBroadcaster(url, opts...)
			},
		},
	}
}
