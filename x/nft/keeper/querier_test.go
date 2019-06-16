package keeper

import (
	"encoding/binary"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/nft/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestQuerySupply(t *testing.T) {
	ctx, keeper, cdc := Initialize()

	addresses := CreateTestAddrs(1)
	id := "test_id"
	denom := "test_denom"
	address := addresses[0]
	tokenURI := "https://google.com"
	description := "test_description"
	image := "test_image"
	name := "test_name"

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.Nil(t, err)

	querier := NewQuerier(keeper)

	queryCollectionParams := types.NewQueryCollectionParams(denom)
	bz, errRes := cdc.MarshalJSON(queryCollectionParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	query.Path = "/custom/nft/supply"
	query.Data = bz

	res, err := querier(ctx, []string{"supply"}, query)
	supplyResp := binary.LittleEndian.Uint64(res)
	require.Equal(t, 1, int(supplyResp))
}

func TestQueryCollection(t *testing.T) {
	ctx, keeper, cdc := Initialize()

	addresses := CreateTestAddrs(1)
	id := "test_id"
	denom := "test_denom"
	address := addresses[0]
	tokenURI := "https://google.com"
	description := "test_description"
	image := "test_image"
	name := "test_name"

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.Nil(t, err)

	querier := NewQuerier(keeper)

	queryCollectionParams := types.NewQueryCollectionParams(denom)
	bz, errRes := cdc.MarshalJSON(queryCollectionParams)
	require.Nil(t, errRes)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	query.Path = "/custom/nft/collection"
	query.Data = bz

	res, err := querier(ctx, []string{"collection"}, query)

	var collection types.Collection
	types.ModuleCdc.MustUnmarshalJSON(res, &collection)

	require.Equal(t, len(collection.NFTs), 1)
}
