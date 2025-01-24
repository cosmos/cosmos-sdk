package runtime

import (
	"context"

	"cosmossdk.io/core/gas"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ gas.Service = GasService{}

type GasService struct{}

func (g GasService) GasMeter(ctx context.Context) gas.Meter {
	return CoreGasmeter{gm: sdk.UnwrapSDKContext(ctx).GasMeter()}
}

func (g GasService) BlockGasMeter(ctx context.Context) gas.Meter {
	return CoreGasmeter{gm: sdk.UnwrapSDKContext(ctx).BlockGasMeter()}
}

func (g GasService) GasConfig(ctx context.Context) gas.GasConfig {
	return gas.GasConfig(sdk.UnwrapSDKContext(ctx).KVGasConfig())
}

// CoreGasmeter is a wrapper around the SDK's GasMeter that implements the GasMeter interface.
type CoreGasmeter struct {
	gm storetypes.GasMeter
}

func (cgm CoreGasmeter) Consume(amount gas.Gas, descriptor string) error {
	cgm.gm.ConsumeGas(amount, descriptor)
	return nil
}

func (cgm CoreGasmeter) Consumed() gas.Gas {
	return cgm.gm.GasConsumed()
}

func (cgm CoreGasmeter) Refund(amount gas.Gas, descriptor string) error {
	cgm.gm.RefundGas(amount, descriptor)
	return nil
}

func (cgm CoreGasmeter) Remaining() gas.Gas {
	return cgm.gm.GasRemaining()
}

func (cgm CoreGasmeter) Limit() gas.Gas {
	return cgm.gm.Limit()
}

type GasConfig struct {
	gc gas.GasConfig
}

func (gc GasConfig) HasCost() gas.Gas {
	return gc.gc.HasCost
}

func (gc GasConfig) DeleteCost() gas.Gas {
	return gc.gc.DeleteCost
}

func (gc GasConfig) ReadCostFlat() gas.Gas {
	return gc.gc.ReadCostFlat
}

func (gc GasConfig) ReadCostPerByte() gas.Gas {
	return gc.gc.ReadCostPerByte
}

func (gc GasConfig) WriteCostFlat() gas.Gas {
	return gc.gc.WriteCostFlat
}

func (gc GasConfig) WriteCostPerByte() gas.Gas {
	return gc.gc.WriteCostPerByte
}

func (gc GasConfig) IterNextCostFlat() gas.Gas {
	return gc.gc.IterNextCostFlat
}
