package stf

import (
	"context"

	"cosmossdk.io/core/gas"
)

func NewGasMeterService() gas.Service {
	return gasService{}
}

type gasService struct{}

// GetGasConfig implements gas.Service.
func (g gasService) GetGasConfig(ctx context.Context) gas.GasConfig {
	panic("unimplemented")
}

func (g gasService) GetGasMeter(ctx context.Context) gas.Meter {
	return ctx.(*ExecutionContext).meter
}

func (g gasService) GetBlockGasMeter(ctx context.Context) gas.Meter {
	panic("stf has no block gas meter")
}

func (g gasService) WithGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	panic("unimplemented")
}

func (g gasService) WithBlockGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	panic("unimplemented")
}
