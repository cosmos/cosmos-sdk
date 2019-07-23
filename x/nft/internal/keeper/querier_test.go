package keeper

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/x/nft/exported"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

func TestNewQuerier(t *testing.T) {
	ctx, keeper, _ := Initialize()
	querier := NewQuerier(keeper)
	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}
	_, err := querier(ctx, []string{"foo", "bar"}, query)
	require.Error(t, err)
}

func TestQuerySupply(t *testing.T) {
	ctx, keeper, cdc := Initialize()

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	querier := NewQuerier(keeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	query.Path = "/custom/nft/supply"
	query.Data = []byte("?")

	res, err := querier(ctx, []string{"supply"}, query)
	require.Error(t, err)

	queryCollectionParams := types.NewQueryCollectionParams(denom2)
	bz, errRes := cdc.MarshalJSON(queryCollectionParams)
	require.Nil(t, errRes)
	query.Data = bz
	res, err = querier(ctx, []string{"supply"}, query)
	require.Error(t, err)

	queryCollectionParams = types.NewQueryCollectionParams(denom)
	bz, errRes = cdc.MarshalJSON(queryCollectionParams)
	require.Nil(t, errRes)
	query.Data = bz

	res, err = querier(ctx, []string{"supply"}, query)
	supplyResp := binary.LittleEndian.Uint64(res)
	require.Equal(t, 1, int(supplyResp))
}

func TestQueryCollection(t *testing.T) {
	ctx, keeper, cdc := Initialize()

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	querier := NewQuerier(keeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	query.Path = "/custom/nft/collection"

	query.Data = []byte("?")
	res, err := querier(ctx, []string{"collection"}, query)
	require.Error(t, err)

	queryCollectionParams := types.NewQueryCollectionParams(denom2)
	bz, errRes := cdc.MarshalJSON(queryCollectionParams)
	require.Nil(t, errRes)

	query.Data = bz
	res, err = querier(ctx, []string{"collection"}, query)
	require.Error(t, err)

	queryCollectionParams = types.NewQueryCollectionParams(denom)
	bz, errRes = cdc.MarshalJSON(queryCollectionParams)
	require.Nil(t, errRes)

	query.Data = bz
	res, err = querier(ctx, []string{"collection"}, query)

	var collections types.Collections
	types.ModuleCdc.MustUnmarshalJSON(res, &collections)
	require.Len(t, collections, 1)
	require.Len(t, collections[0].NFTs, 1)
}

func TestQueryOwner(t *testing.T) {
	ctx, keeper, cdc := Initialize()

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	denom2 := "test_denom2"
	err = keeper.MintNFT(ctx, denom2, &nft)
	require.NoError(t, err)

	querier := NewQuerier(keeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}
	query.Path = "/custom/nft/ownerByDenom"

	query.Data = []byte("?")
	_, err = querier(ctx, []string{"ownerByDenom"}, query)
	require.Error(t, err)

	// query the balance using the first denom
	params := types.NewQueryBalanceParams(address, denom)
	bz, err2 := cdc.MarshalJSON(params)
	require.Nil(t, err2)
	query.Data = bz

	var res []byte
	res, err = querier(ctx, []string{"ownerByDenom"}, query)
	require.NoError(t, err)
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
	query.Data = []byte("?")
	_, err = querier(ctx, []string{"owner"}, query)
	require.Error(t, err)

	query.Data = bz
	res, err = querier(ctx, []string{"owner"}, query)

	require.NoError(t, err)
	cdc.MustUnmarshalJSON(res, &out)

	// build the owner using both denoms TODO: add sorting to ensure the objects are the same
	idCollection2 := types.NewIDCollection(denom2, []string{id})
	owner = types.NewOwner(address, idCollection2, idCollection1)

	require.Equal(t, out.String(), owner.String())
}

func TestQueryNFT(t *testing.T) {
	ctx, keeper, cdc := Initialize()

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	querier := NewQuerier(keeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}
	query.Path = "/custom/nft/nft"
	var res []byte

	query.Data = []byte("?")
	res, err = querier(ctx, []string{"nft"}, query)
	require.Error(t, err)

	params := types.NewQueryNFTParams(denom2, id2)
	bz, err2 := cdc.MarshalJSON(params)
	require.Nil(t, err2)

	query.Data = bz
	res, err = querier(ctx, []string{"nft"}, query)
	require.Error(t, err)

	params = types.NewQueryNFTParams(denom, id)
	bz, err2 = cdc.MarshalJSON(params)
	require.Nil(t, err2)

	query.Data = bz
	res, err = querier(ctx, []string{"nft"}, query)

	var out exported.NFT
	cdc.MustUnmarshalJSON(res, &out)

	require.Equal(t, out.String(), nft.String())
}

func TestQueryDenoms(t *testing.T) {
	ctx, keeper, cdc := Initialize()

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	err = keeper.MintNFT(ctx, denom2, &nft)
	require.NoError(t, err)

	querier := NewQuerier(keeper)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}
	var res []byte
	query.Path = "/custom/nft/denoms"

	res, err = querier(ctx, []string{"denoms"}, query)

	denoms := []string{denom, denom2}

	var out []string
	cdc.MustUnmarshalJSON(res, &out)

	for key, denomInQuestion := range out {
		require.Equal(t, denomInQuestion, denoms[key])
	}
}
