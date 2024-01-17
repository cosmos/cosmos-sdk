package runtime

import (
	"context"

	"cosmossdk.io/core/gas"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ gas.Service = (*GasService)(nil)

type GasService struct{}

func (g GasService) GetGasMeter(ctx context.Context) gas.Meter {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.GasMeter()
}

func (g GasService) GetBlockGasMeter(ctx context.Context) gas.Meter {
	return sdk.UnwrapSDKContext(ctx).BlockGasMeter()
}

func (g GasService) WithGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	return sdk.UnwrapSDKContext(ctx).WithGasMeter(meter)
}

func (g GasService) WithBlockGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	return sdk.UnwrapSDKContext(ctx).WithBlockGasMeter(meter)
}
