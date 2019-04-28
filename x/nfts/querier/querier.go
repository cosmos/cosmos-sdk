package querier

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nfts/keeper"
	"github.com/cosmos/cosmos-sdk/x/nfts/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the nft Querier
const (
	QuerySupply     = "supply"
	QueryBalance    = "balance"
	QueryCollection = "collection"
	QueryNFT        = "nft"
)

// NewQuerier is the module level router for state queries
func NewQuerier(k keeper.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QuerySupply:
			return querySupply(ctx, path[1:], req, k)
		case QueryBalance:
			return queryBalance(ctx, path[1:], req, k)
		case QueryCollection:
			return queryCollection(ctx, path[1:], req, k)
		case QueryNFT:
			return querNFT(ctx, path[1:], req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown nft query endpoint")
		}
	}
}

func queryBalanceOf(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
	denom := types.Denom(path[0])
	owner := path[1]

	collection, found := k.GetCollection(ctx, denom)
	if !found {
		return []byte{}, types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("Unknown denom %s", denom))
	}

	balance := 0
	for _, v := range collection {
		if v.Owner.String() == owner {
			balance++
		}
	}

	bz, err2 := codec.MarshalJSONIndent(k.Cdc, QueryResBalance{denom, balance})
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

// QueryResBalance resolves int balance
type QueryResBalance struct {
	Denom   types.Denom `json:"denom"`
	Balance int         `json:"balance"`
}

func (p QueryResBalance) String() string {
	return fmt.Sprintf("%s %d", p.Denom, p.Balance)
}

func queryOwnerOf(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
	denom := types.Denom(path[0])
	uintID, error := strconv.ParseUint(path[1], 10, 64)
	if error != nil {
		panic("could not parse TokenID string")
	}
	id := types.TokenID(uintID)

	nft, err := k.GetNFT(ctx, denom, id)
	if err != nil {
		return
	}

	bz, err2 := codec.MarshalJSONIndent(k.Cdc, QueryResOwnerOf{nft.Owner})
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

//QueryResOwnerOf resolves sdk.AccAddress owner
type QueryResOwnerOf struct {
	Owner sdk.AccAddress `json:"owner"`
}

func (q QueryResOwnerOf) String() string {
	return q.Owner.String()
}

func queryMetadata(ctx sdk.Context, path []string, req abci.RequestQuery, k keeper.Keeper) (res []byte, err sdk.Error) {
	denom := types.Denom(path[0])
	uintID, error := strconv.ParseUint(path[1], 10, 64)
	if error != nil {
		panic("could not parse TokenID string")
	}
	id := types.TokenID(uintID)

	nft, err := k.GetNFT(ctx, denom, id)
	if err != nil {
		return
	}

	bz, err2 := codec.MarshalJSONIndent(k.Cdc, nft)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}
