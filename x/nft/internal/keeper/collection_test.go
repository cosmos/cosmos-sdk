package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

func TestSetCollection(t *testing.T) {
	ctx, keeper, _ := Initialize()

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	// collection should exist
	collection, exists := keeper.GetCollection(ctx, denom)
	require.True(t, exists)

	nft2 := types.NewBaseNFT(id2, address, tokenURI)
	collection, err = collection.AddNFT(&nft2)
	require.NoError(t, err)
	keeper.SetCollection(ctx, denom, collection)

	collection, exists = keeper.GetCollection(ctx, denom)
	require.True(t, exists)
	require.Len(t, collection.NFTs, 2)

}
func TestGetCollection(t *testing.T) {
	ctx, keeper, _ := Initialize()

	// collection shouldn't exist
	collection, exists := keeper.GetCollection(ctx, denom)
	require.Empty(t, collection)
	require.False(t, exists)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	// collection should exist
	collection, exists = keeper.GetCollection(ctx, denom)
	require.True(t, exists)
	require.NotEmpty(t, collection)
}
func TestGetCollections(t *testing.T) {
	ctx, keeper, _ := Initialize()

	// collections should be empty
	collections := keeper.GetCollections(ctx)
	require.Empty(t, collections)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	// collections should equal 1
	collections = keeper.GetCollections(ctx)
	require.NotEmpty(t, collections)
	require.Equal(t, len(collections), 1)
}
