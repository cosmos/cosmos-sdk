package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type Querier struct {
	Keeper
}

var _ types.QueryServiceServer = Querier{}

func (q Querier) QueryBalance(ctx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	balance := q.GetBalance(sdk.UnwrapSDKContext(ctx), req.Address, req.Denom)
	return &types.QueryBalanceResponse{Balance: &balance}, nil
}

func (q Querier) QueryAllBalances(ctx context.Context, req *types.QueryAllBalancesRequest) (*types.QueryAllBalancesResponse, error) {
	balances := q.GetAllBalances(sdk.UnwrapSDKContext(ctx), req.Address)
	return &types.QueryAllBalancesResponse{Balances: balances}, nil
}

// NewQuerier returns a new sdk.Keeper instance.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryBalance:
			return queryBalance(ctx, req, k)

		case types.QueryAllBalances:
			return queryAllBalance(ctx, req, k)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}
	}
}

func queryBalance(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryBalanceRequest

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	balance := k.GetBalance(ctx, params.Address, params.Denom)

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, balance)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAllBalance(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryAllBalancesRequest

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	balances := k.GetAllBalances(ctx, params.Address)

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, balances)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
