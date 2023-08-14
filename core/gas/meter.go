// Package gas provides a basic API for app modules to track gas usage.
package gas

import "context"

type Gas = uint64

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
}

// Meter represents a gas meter.
type Meter interface {
	GasConsumed() Gas
	GasConsumedToLimit() Gas
	GasRemaining() Gas
	Limit() Gas
	ConsumeGas(amount Gas, descriptor string)
	RefundGas(amount Gas, descriptor string)
	IsPastLimit() bool
	IsOutOfGas() bool
	String() string
}
