package runtime

import (
	"context"

	"cosmossdk.io/core/gas"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ gas.Service = GasService{}

type GasService struct {
	storetypes.GasMeter
}

func (g GasService) GetGasMeter(ctx context.Context) gas.Meter {
	return sdk.UnwrapSDKContext(ctx).GasMeter()
}

func (g GasService) GetBlockGasMeter(ctx context.Context) gas.Meter {
	return sdk.UnwrapSDKContext(ctx).BlockGasMeter()
}

func (g GasService) WithGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	return sdk.UnwrapSDKContext(ctx).WithGasMeter(meter)
}

func (g GasService) WithBlockGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	return sdk.UnwrapSDKContext(ctx).WithGasMeter()
}

// ______________________________________________________________________________________________
// Gas Meter Wrappers
// ______________________________________________________________________________________________

type SDKGasMeter struct {
	gas.Meter
}

type CoreGasmeter struct {
	storetypes.GasMeter
}
