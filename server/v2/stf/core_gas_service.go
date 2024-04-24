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
func (g gasService) GasConfig(ctx context.Context) gas.GasConfig {
	panic("unimplemented")
}

func (g gasService) GasMeter(ctx context.Context) gas.Meter {
	return ctx.(*executionContext).meter
}

func (g gasService) BlockGasMeter(ctx context.Context) gas.Meter {
	panic("stf has no block gas meter")
}

func (g gasService) WithGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	panic("unimplemented")
}

func (g gasService) WithBlockGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	panic("unimplemented")
}
