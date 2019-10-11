package keeper_test

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/x/nft/exported"
	keep "github.com/cosmos/cosmos-sdk/x/nft/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

func TestNewQuerier(t *testing.T) {
	app, ctx := createTestApp(false)
	querier := keep.NewQuerier(app.NFTKeeper)
	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}
	_, err := querier(ctx, []string{"foo", "bar"}, query)
	require.Error(t, err)
}

func TestQuerySupply(t *testing.T) {
	app, ctx := createTestApp(false)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	querier := keep.NewQuerier(app.NFTKeeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	query.Path = "/custom/nft/supply"
	query.Data = []byte("?")

	res, err := querier(ctx, []string{"supply"}, query)
	require.Error(t, err)
	require.Nil(t, res)

	queryCollectionParams := types.NewQueryCollectionParams(denom2)
	bz, errRes := app.Codec().MarshalJSON(queryCollectionParams)
	require.Nil(t, errRes)
	query.Data = bz
	res, err = querier(ctx, []string{"supply"}, query)
	require.Error(t, err)
	require.Nil(t, res)

	queryCollectionParams = types.NewQueryCollectionParams(denom)
	bz, errRes = app.Codec().MarshalJSON(queryCollectionParams)
	require.Nil(t, errRes)
	query.Data = bz

	res, err = querier(ctx, []string{"supply"}, query)
	require.NoError(t, err)
	require.NotNil(t, res)

	supplyResp := binary.LittleEndian.Uint64(res)
	require.Equal(t, 1, int(supplyResp))
}

func TestQueryCollection(t *testing.T) {
	app, ctx := createTestApp(false)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	querier := keep.NewQuerier(app.NFTKeeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	query.Path = "/custom/nft/collection"

	query.Data = []byte("?")
	res, err := querier(ctx, []string{"collection"}, query)
	require.Error(t, err)
	require.Nil(t, res)

	queryCollectionParams := types.NewQueryCollectionParams(denom2)
	bz, errRes := app.Codec().MarshalJSON(queryCollectionParams)
	require.Nil(t, errRes)

	query.Data = bz
	res, err = querier(ctx, []string{"collection"}, query)
	require.Error(t, err)
	require.Nil(t, res)

	queryCollectionParams = types.NewQueryCollectionParams(denom)
	bz, errRes = app.Codec().MarshalJSON(queryCollectionParams)
	require.Nil(t, errRes)

	query.Data = bz
	res, err = querier(ctx, []string{"collection"}, query)
	require.NoError(t, err)
	require.NotNil(t, res)

	var collections types.Collections
	types.ModuleCdc.MustUnmarshalJSON(res, &collections)
	require.Len(t, collections, 1)
	require.Len(t, collections[0].NFTs, 1)
}

func TestQueryOwner(t *testing.T) {
	app, ctx := createTestApp(false)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	denom2 := "test_denom2"
	err = app.NFTKeeper.MintNFT(ctx, denom2, &nft)
	require.NoError(t, err)

	querier := keep.NewQuerier(app.NFTKeeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}
	query.Path = "/custom/nft/ownerByDenom"

	query.Data = []byte("?")
	res, err := querier(ctx, []string{"ownerByDenom"}, query)
	require.Error(t, err)
	require.Nil(t, res)

	// query the balance using the first denom
	params := types.NewQueryBalanceParams(address, denom)
	bz, err2 := app.Codec().MarshalJSON(params)
	require.Nil(t, err2)
	query.Data = bz

	res, err = querier(ctx, []string{"ownerByDenom"}, query)
	require.NoError(t, err)
	require.NotNil(t, res)

	var out types.Owner
	app.Codec().MustUnmarshalJSON(res, &out)

	// build the owner using only the first denom
	idCollection1 := types.NewIDCollection(denom, []string{id})
	owner := types.NewOwner(address, idCollection1)

	require.Equal(t, out.String(), owner.String())

	// query the balance using no denom so that all denoms will be returns
	params = types.NewQueryBalanceParams(address, "")
	bz, err2 = app.Codec().MarshalJSON(params)
	require.Nil(t, err2)

	query.Path = "/custom/nft/owner"
	query.Data = []byte("?")
	_, err = querier(ctx, []string{"owner"}, query)
	require.Error(t, err)

	query.Data = bz
	res, err = querier(ctx, []string{"owner"}, query)
	require.NoError(t, err)
	require.NotNil(t, res)

	app.Codec().MustUnmarshalJSON(res, &out)

	// build the owner using both denoms TODO: add sorting to ensure the objects are the same
	idCollection2 := types.NewIDCollection(denom2, []string{id})
	owner = types.NewOwner(address, idCollection2, idCollection1)

	require.Equal(t, out.String(), owner.String())
}

func TestQueryNFT(t *testing.T) {
	app, ctx := createTestApp(false)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	querier := keep.NewQuerier(app.NFTKeeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}
	query.Path = "/custom/nft/nft"
	var res []byte

	query.Data = []byte("?")
	res, err = querier(ctx, []string{"nft"}, query)
	require.Error(t, err)
	require.Nil(t, res)

	params := types.NewQueryNFTParams(denom2, id2)
	bz, err2 := app.Codec().MarshalJSON(params)
	require.Nil(t, err2)

	query.Data = bz
	res, err = querier(ctx, []string{"nft"}, query)
	require.Error(t, err)
	require.Nil(t, res)

	params = types.NewQueryNFTParams(denom, id)
	bz, err2 = app.Codec().MarshalJSON(params)
	require.Nil(t, err2)

	query.Data = bz
	res, err = querier(ctx, []string{"nft"}, query)
	require.NoError(t, err)
	require.NotNil(t, res)

	var out exported.NFT
	app.Codec().MustUnmarshalJSON(res, &out)

	require.Equal(t, out.String(), nft.String())
}

func TestQueryDenoms(t *testing.T) {
	app, ctx := createTestApp(false)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	err = app.NFTKeeper.MintNFT(ctx, denom2, &nft)
	require.NoError(t, err)

	querier := keep.NewQuerier(app.NFTKeeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}
	var res []byte
	query.Path = "/custom/nft/denoms"

	res, err = querier(ctx, []string{"denoms"}, query)
	require.NoError(t, err)
	require.NotNil(t, res)

	denoms := []string{denom, denom2}

	var out []string
	app.Codec().MustUnmarshalJSON(res, &out)

	for key, denomInQuestion := range out {
		require.Equal(t, denomInQuestion, denoms[key])
	}
}
