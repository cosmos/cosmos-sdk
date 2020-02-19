package keeper

import (
	"encoding/binary"
	types2 "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewQuerier creates a querier for upgrade cli and REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {

		case types2.QueryCurrent:
			return queryCurrent(ctx, req, k)

		case types2.QueryApplied:
			return queryApplied(ctx, req, k)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types2.ModuleName, path[0])
		}
	}
}

func queryCurrent(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	plan, has := k.GetUpgradePlan(ctx)
	if !has {
		return nil, nil
	}

	res, err := k.cdc.MarshalJSON(&plan)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryApplied(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types2.QueryAppliedParams

	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	applied := k.GetDoneHeight(ctx, params.Name)
	if applied == 0 {
		return nil, nil
	}

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(applied))

	return bz, nil
}
