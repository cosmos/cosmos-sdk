// Package host provides a telemetry instrument wrapper for host-level metrics.
package host

import (
	"go.opentelemetry.io/contrib/instrumentation/host"

	"github.com/cosmos/cosmos-sdk/telemetry/registry"
)

// Name is the instrument name used in configuration.
const Name = "host"

func init() {
	registry.Register(instrument{})
}

type instrument struct{}

func (instrument) Name() string { return Name }

func (instrument) Start(_ map[string]any) error {
	return host.Start()
}
