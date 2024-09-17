package stf

import (
	"context"

	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
	stfgas "cosmossdk.io/server/v2/stf/gas"
)

type (
	// makeGasMeterFn is a function type that takes a gas limit as input and returns a gas.Meter.
	// It is used to measure and limit the amount of gas consumed during the execution of a function.
	makeGasMeterFn func(gasLimit uint64) gas.Meter

	// makeGasMeteredStateFn is a function type that wraps a gas meter and a store writer map.
	makeGasMeteredStateFn func(meter gas.Meter, store store.WriterMap) store.WriterMap
)

// NewGasMeterService creates a new instance of the gas meter service.
func NewGasMeterService() gas.Service {
	return gasService{}
}

type gasService struct{}

// GasConfig implements gas.Service.
func (g gasService) GasConfig(ctx context.Context) gas.GasConfig {
	return stfgas.DefaultConfig
}

func (g gasService) GasMeter(ctx context.Context) gas.Meter {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		panic(err)
	}

	return exCtx.meter
}
