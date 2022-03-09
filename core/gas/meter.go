package gas

import "context"

type Gas = uint64

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

// GetMeter returns the current transaction-level gas meter. A non-nil meter
// is always returned. When one is unavailable in the context a dummy instance
// will be returned.
func GetMeter(ctx context.Context) Meter {
	panic("TODO")
}

// GetBlockMeter returns the current block-level gas meter. A non-nil meter
// is always returned. When one is unavailable in the context a dummy instance
// will be returned.
func GetBlockMeter(ctx context.Context) Meter {
	panic("TODO")
}

// WithMeter returns a new context with the provided transaction-level gas meter.
func WithMeter(ctx context.Context, meter Meter) context.Context {
	panic("TODO")
}

// WithMeter returns a new context with the provided block-level gas meter.
func WithBlockMeter(ctx context.Context, meter Meter) context.Context {
	panic("TODO")
}
