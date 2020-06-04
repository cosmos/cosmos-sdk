package keeper

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ types.QueryServer = BaseKeeper{}

// Balance implements the Query/Balance gRPC method
func (q BaseKeeper) Balance(ctx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	balance := q.GetBalance(sdk.UnwrapSDKContext(ctx), req.Address, req.Denom)
	return &types.QueryBalanceResponse{Balance: &balance}, nil
}

// AllBalances implements the Query/AllBalances gRPC method
func (q BaseKeeper) AllBalances(c context.Context, req *types.QueryAllBalancesRequest) (*types.QueryAllBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	balances := q.GetAllBalances(ctx, req.Address)

	return &types.QueryAllBalancesResponse{Balances: balances}, nil
}

// TotalSupply implements the Query/TotalSupply gRPC method
func (q BaseKeeper) TotalSupply(c context.Context, request *types.QueryTotalSupplyRequest) (*types.QueryTotalSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// TODO: add pagination once it can be supported properly with supply no longer being stored as a single blob
	totalSupply := q.GetSupply(ctx).GetTotal()

	return &types.QueryTotalSupplyResponse{Supply: totalSupply}, nil
}

// SupplyOf implements the Query/SupplyOf gRPC method
func (q BaseKeeper) SupplyOf(c context.Context, request *types.QuerySupplyOfRequest) (*types.QuerySupplyOfResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	supply := q.GetSupply(ctx).GetTotal().AmountOf(request.Denom)

	return &types.QuerySupplyOfResponse{Amount: supply}, nil
}

// NewQuerier returns a new sdk.Keeper instance.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryBalance:
			return queryBalance(ctx, req, k)

		case types.QueryAllBalances:
			return queryAllBalance(ctx, req, k)

		case types.QueryTotalSupply:
			return queryTotalSupply(ctx, req, k)

		case types.QuerySupplyOf:
			return querySupplyOf(ctx, req, k)

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

func queryTotalSupply(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryTotalSupplyParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	totalSupply := k.GetSupply(ctx).GetTotal()

	start, end := client.Paginate(len(totalSupply), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		totalSupply = sdk.Coins{}
	} else {
		totalSupply = totalSupply[start:end]
	}

	res, err := totalSupply.MarshalJSON()
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func querySupplyOf(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QuerySupplyOfParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	supply := k.GetSupply(ctx).GetTotal().AmountOf(params.Denom)

	res, err := supply.MarshalJSON()
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
