package keeper

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/internal/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier creates a querier for upgrade cli and REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {

		case types.QueryCurrent:
			return queryCurrent(ctx, req, k)

		case types.QueryApplied:
			return queryApplied(ctx, req, k)

		default:
			return nil, sdk.ErrUnknownRequest("unknown supply query endpoint")
		}
	}
}

func queryCurrent(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	plan, has := k.GetUpgradePlan(ctx)
	if !has {
		// empty data - client can respond Not Found
		return nil, nil
	}
	res, err := k.cdc.MarshalJSON(&plan)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}
	return res, nil
}

func queryApplied(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryAppliedParams

	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	applied := k.GetDoneHeight(ctx, params.Name)
	if applied == 0 {
		// empty data - client can respond Not Found
		return nil, nil
	}
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(applied))
	return bz, nil
}
