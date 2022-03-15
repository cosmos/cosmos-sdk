package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/tieredfee/types"
)

// Params returns the current consensus parameters
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{
		Params: &params,
	}, nil
}

// GasPrices returns the current consensus parameters
func (k Keeper) GasPrices(goCtx context.Context, req *types.QueryGasPricesRequest) (*types.QueryGasPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	gasPrices := k.GetAllGasPrice(ctx)
	return &types.QueryGasPricesResponse{
		GasPrices: types.ToProtoGasPrices(gasPrices),
	}, nil
}

// BlockGasUsed returns the current consensus parameters
func (k Keeper) BlockGasUsed(goCtx context.Context, req *types.QueryBlockGasUsedRequest) (*types.QueryBlockGasUsedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	gasUsed, found := k.GetBlockGasUsed(ctx)
	if !found {
		return nil, sdkerrors.ErrNotFound.Wrapf("failed to load the block gas used")
	}
	return &types.QueryBlockGasUsedResponse{
		GasUsed: gasUsed,
	}, nil
}
