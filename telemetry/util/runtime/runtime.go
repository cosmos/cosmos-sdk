// Package runtime provides a telemetry instrument wrapper for Go runtime metrics.
package runtime

import (
	"go.opentelemetry.io/contrib/instrumentation/runtime"

	"github.com/cosmos/cosmos-sdk/telemetry/registry"
)

// Name is the instrument name used in configuration.
const Name = "runtime"

func init() {
	registry.Register(instrument{})
}

type instrument struct{}

func (instrument) Name() string { return Name }

func (instrument) Start(_ map[string]any) error {
	return runtime.Start()
}
