package coretesting

import (
	"context"
	"cosmossdk.io/core/gas"
)

var _ gas.Service = &MemGasService{}

type MemGasService struct{}

func (m MemGasService) GasMeter(ctx context.Context) gas.Meter {
	dummy := unwrap(ctx)

	return dummy.gasMeter
}

func (m MemGasService) GasConfig(ctx context.Context) gas.GasConfig {
	dummy := unwrap(ctx)

	return dummy.gasConfig
}
