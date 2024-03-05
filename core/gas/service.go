// Package gas provides a basic API for app modules to track gas usage.
package gas

import (
	"context"
	"errors"
	"math"
)

// ErrOutOfGas must be used by GasMeter implementers to signal
// that the state transition consumed all the allowed computational
// gas.
var ErrOutOfGas = errors.New("out of gas")

// Gas defines type alias of uint64 for gas consumption. Gas is used
// to measure computational overhead when executing state transitions,
// it might be related to storage access and not only.
type Gas = uint64

// NoGasLimit signals that no gas limit must be applied.
const NoGasLimit Gas = math.MaxUint64

// Service represents a gas service which can retrieve and set a gas meter in a context.
// gas.Service is a core API type that should be provided by the runtime module being used to
// build an app via depinject.
type Service interface {
	// GetGasMeter returns the current transaction-level gas meter. A non-nil meter
	// is always returned. When one is unavailable in the context an infinite gas meter
	// will be returned.
	GetGasMeter(context.Context) Meter

	// GetBlockGasMeter returns the current block-level gas meter. A non-nil meter
	// is always returned. When one is unavailable in the context an infinite gas meter
	// will be returned.
	GetBlockGasMeter(context.Context) Meter

	// WithGasMeter returns a new context with the provided transaction-level gas meter.
	WithGasMeter(ctx context.Context, meter Meter) context.Context

	// WithBlockGasMeter returns a new context with the provided block-level gas meter.
	WithBlockGasMeter(ctx context.Context, meter Meter) context.Context

	GetGasConfig(ctx context.Context) GasConfig
}

// Meter represents a gas meter for modules consumption
type Meter interface {
	Consume(amount Gas, descriptor string)
	Refund(amount Gas, descriptor string)
	Remaining() Gas
	Limit() Gas
}

type GasConfig struct {
	HasCost          Gas
	DeleteCost       Gas
	ReadCostFlat     Gas
	ReadCostPerByte  Gas
	WriteCostFlat    Gas
	WriteCostPerByte Gas
	IterNextCostFlat Gas
}
