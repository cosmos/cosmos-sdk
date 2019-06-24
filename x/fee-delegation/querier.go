package fee_delegation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	QueryGetFeeAllowances = "fees"
)

// NewQuerier creates a new querier
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryGetFeeAllowances:
			return queryGetFeeAllowances(ctx, path[1:], keeper)
		default:
			return nil, sdk.ErrUnknownRequest("Unknown package delegation query endpoint")
		}
	}
}

func queryGetFeeAllowances(ctx sdk.Context, args []string, keeper Keeper) ([]byte, sdk.Error) {
	grantee := args[0]
	granteeAddr, err := sdk.AccAddressFromBech32(grantee)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("invalid address", err.Error()))
	}

	fees := keeper.GetFeeAllowances(ctx, granteeAddr)
	if fees == nil {
		fees = []FeeAllowanceGrant{}
	}
	bz, jErr := keeper.cdc.MarshalJSON(fees)
	if jErr != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", jErr.Error()))
	}
	return bz, nil
}
