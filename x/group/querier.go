package group

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the governance Querier
const (
	QueryGet = "get"
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryGet:
			return queryGroup(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown data query endpoint")
		}
	}
}

func queryGroup(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	idStr := path[0]

	decodedId, e := sdk.AccAddressFromBech32(idStr)

	if e != nil {
		return []byte{}, sdk.ErrUnknownRequest("could not decode group ID")
	}

	info, err := keeper.GetGroupInfo(ctx, decodedId)

	if err != nil {
		return []byte{}, err
	}

	res, jsonErr := codec.MarshalJSONIndent(keeper.cdc, info)
	if jsonErr != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", jsonErr.Error()))
	}
	return res, nil
}
