package runtime

import (
	"context"

	"cosmossdk.io/core/gas"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ gas.Service = GasService{}

type GasService struct{}

func (g GasService) GetGasMeter(ctx context.Context) gas.Meter {
	return CoreGasmeter{gm: sdk.UnwrapSDKContext(ctx).GasMeter()}
}

func (g GasService) GetBlockGasMeter(ctx context.Context) gas.Meter {
	return CoreGasmeter{gm: sdk.UnwrapSDKContext(ctx).BlockGasMeter()}
}

func (g GasService) WithGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	return sdk.UnwrapSDKContext(ctx).WithGasMeter(SDKGasMeter{gm: meter})
}

func (g GasService) WithBlockGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	return sdk.UnwrapSDKContext(ctx).WithGasMeter(SDKGasMeter{gm: meter})
}

// ______________________________________________________________________________________________
// Gas Meter Wrappers
// ______________________________________________________________________________________________

type SDKGasMeter struct {
	gm gas.Meter
}

func (gm SDKGasMeter) GasConsumed() storetypes.Gas {
	return gm.gm.Remaining()
}

func (gm SDKGasMeter) GasConsumedToLimit() storetypes.Gas {
	return gm.gm.Limit() // TODO is this correct?
}

func (gm SDKGasMeter) GasRemaining() storetypes.Gas {
	return gm.gm.Remaining()
}

func (gm SDKGasMeter) Limit() storetypes.Gas {
	return gm.gm.Limit()
}

func (gm SDKGasMeter) ConsumeGas(amount storetypes.Gas, descriptor string) {
	gm.gm.Consume(amount, descriptor)
}

func (gm SDKGasMeter) RefundGas(amount storetypes.Gas, descriptor string) {
	gm.gm.Refund(amount, descriptor)
}

func (gm SDKGasMeter) IsPastLimit() bool {
	return gm.gm.Remaining() <= gm.gm.Limit()
}

func (gm SDKGasMeter) IsOutOfGas() bool {
	return gm.gm.Remaining() >= gm.gm.Limit()
}

func (gm SDKGasMeter) String() string {
	return "gas go fast" // TODO
}

type CoreGasmeter struct {
	gm storetypes.GasMeter
}

func (cgm CoreGasmeter) Consume(amount gas.Gas, descriptor string) {
	cgm.gm.ConsumeGas(amount, descriptor)
}

func (cgm CoreGasmeter) Refund(amount gas.Gas, descriptor string) {
	cgm.gm.RefundGas(amount, descriptor)
}

func (cgm CoreGasmeter) Remaining() gas.Gas {
	return cgm.gm.GasRemaining()
}

func (cgm CoreGasmeter) Limit() gas.Gas {
	return cgm.gm.Limit()
}
