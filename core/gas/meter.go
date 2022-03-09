package gas

import "context"

type Gas = uint64

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

func GetMeter(ctx context.Context) Meter {
	panic("TODO")
}

func GetBlockMeter(ctx context.Context) Meter {
	panic("TODO")
}

func WithMeter(ctx context.Context, meter Meter) context.Context {
	panic("TODO")
}

func WithBlockMeter(ctx context.Context, meter Meter) context.Context {
	panic("TODO")
}
