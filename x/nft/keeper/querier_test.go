package keeper

import (
	"encoding/binary"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/nft/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestQueryCollectionSupply(t *testing.T) {
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

func TestQueryOwner(t *testing.T) {
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

	denom2 := "test_denom2"
	err = keeper.MintNFT(ctx, denom2, &nft)
	require.Nil(t, err)

	querier := NewQuerier(keeper)

	// query the balance using the first denom
	params := types.NewQueryBalanceParams(address, denom)
	bz, err2 := cdc.MarshalJSON(params)
	require.Nil(t, err2)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	query.Data = bz

	var res []byte

	query.Path = "/custom/nft/ownerByDenom"
	res, err = querier(ctx, []string{"ownerByDenom"}, query)
	require.Nil(t, err)
	var out types.Owner
	cdc.MustUnmarshalJSON(res, &out)

	// build the owner using only the first denom
	idCollection1 := types.NewIDCollection(denom, []string{id})
	owner := types.NewOwner(address, idCollection1)

	require.Equal(t, out.String(), owner.String())

	// query the balance using no denom so that all denoms will be returns
	params = types.NewQueryBalanceParams(address, "")
	bz, err2 = cdc.MarshalJSON(params)
	require.Nil(t, err2)

	query.Path = "/custom/nft/owner"
	res, err = querier(ctx, []string{"owner"}, query)

	require.Nil(t, err)
	cdc.MustUnmarshalJSON(res, &out)

	// build the owner using both denoms TODO: add sorting to ensure the objects are the same
	idCollection2 := types.NewIDCollection(denom2, []string{id})
	owner = types.NewOwner(address, idCollection2, idCollection1)

	require.Equal(t, out.String(), owner.String())
}

func TestQueryNFT(t *testing.T) {
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

	params := types.NewQueryNFTParams(denom, id)
	bz, err2 := cdc.MarshalJSON(params)
	require.Nil(t, err2)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}
	var res []byte
	query.Data = bz
	query.Path = "/custom/nft/nft"
	res, err = querier(ctx, []string{"nft"}, query)

	var out types.NFT
	cdc.MustUnmarshalJSON(res, &out)

	require.Equal(t, out.String(), nft.String())
}

func TestQueryDenoms(t *testing.T) {
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

	denom2 := "test_denom2"
	err = keeper.MintNFT(ctx, denom2, &nft)
	require.Nil(t, err)

	querier := NewQuerier(keeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}
	var res []byte
	query.Path = "/custom/nft/denoms"

	res, err = querier(ctx, []string{"denoms"}, query)

	denoms := []string{denom2, denom}

	var out []string
	cdc.MustUnmarshalJSON(res, &out)

	// TODO: add sorting to ensure the objects are the same
	for key, denomInQuestion := range out {
		require.Equal(t, denomInQuestion, denoms[key])
	}
}
