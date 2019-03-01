package auth

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// query endpoints supported by the auth Querier
const (
	QueryAccount = "account"
)

// creates a querier for auth REST endpoints
func NewQuerier(keeper AccountKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryAccount:
			return queryAccount(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown auth query endpoint")
		}
	}
}

// defines the params for query: "custom/acc/account"
type QueryAccountParams struct {
	Address sdk.AccAddress
}

func NewQueryAccountParams(addr sdk.AccAddress) QueryAccountParams {
	return QueryAccountParams{
		Address: addr,
	}
}

func queryAccount(ctx sdk.Context, req abci.RequestQuery, keeper AccountKeeper) ([]byte, sdk.Error) {
	var params QueryAccountParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	account := keeper.GetAccount(ctx, params.Address)
	if account == nil {
		return nil, sdk.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", params.Address))
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, account)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
