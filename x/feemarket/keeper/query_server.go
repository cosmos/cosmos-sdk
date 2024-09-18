package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/x/feemarket/types"
)

var _ types.QueryServer = (*QueryServer)(nil)

// QueryServer defines the gRPC server for the x/feemarket module.
type QueryServer struct {
	k Keeper
}

// NewQueryServer creates a new instance of the x/feemarket QueryServer type.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return &QueryServer{k: keeper}
}

// Params defines a method that returns the current feemarket parameters.
func (q QueryServer) Params(goCtx context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := q.k.GetParams(ctx)
	return &types.ParamsResponse{Params: params}, err
}

// State defines a method that returns the current feemarket state.
func (q QueryServer) State(goCtx context.Context, _ *types.StateRequest) (*types.StateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	state, err := q.k.GetState(ctx)
	return &types.StateResponse{State: state}, err
}

// GasPrice defines a method that returns the current feemarket base gas price.
func (q QueryServer) GasPrice(goCtx context.Context, req *types.GasPriceRequest) (*types.GasPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	gasPrice, err := q.k.GetMinGasPrice(ctx, req.GetDenom())
	return &types.GasPriceResponse{Price: gasPrice}, err
}

// GasPrices defines a method that returns the current feemarket list of gas prices.
func (q QueryServer) GasPrices(goCtx context.Context, _ *types.GasPricesRequest) (*types.GasPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	gasPrices, err := q.k.GetMinGasPrices(ctx)
	return &types.GasPricesResponse{Prices: gasPrices}, err
}
