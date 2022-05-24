package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/upgrade/types"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewQuerier creates a querier for upgrade cli and REST endpoints
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {

		case types.QueryCurrent:
			return queryCurrent(ctx, req, k, legacyQuerierCdc)

		case types.QueryApplied:
			return queryApplied(ctx, req, k, legacyQuerierCdc)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}
	}
}

func queryCurrent(ctx sdk.Context, _ abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	plan, has := k.GetUpgradePlan(ctx)
	if !has {
		return nil, nil
	}

	res, err := legacyQuerierCdc.MarshalJSON(&plan)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryApplied(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryAppliedPlanRequest

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
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
