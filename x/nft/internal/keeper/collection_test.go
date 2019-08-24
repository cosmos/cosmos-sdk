package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

func TestSetCollection(t *testing.T) {
	app, ctx := createTestApp(false)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	// collection should exist
	collection, exists := app.NFTKeeper.GetCollection(ctx, denom)
	require.True(t, exists)

	nft2 := types.NewBaseNFT(id2, address, tokenURI)
	collection, err = collection.AddNFT(&nft2)
	require.NoError(t, err)
	app.NFTKeeper.SetCollection(ctx, denom, collection)

	collection, exists = app.NFTKeeper.GetCollection(ctx, denom)
	require.True(t, exists)
	require.Len(t, collection.NFTs, 2)

}
func TestGetCollection(t *testing.T) {
	app, ctx := createTestApp(false)

	// collection shouldn't exist
	collection, exists := app.NFTKeeper.GetCollection(ctx, denom)
	require.Empty(t, collection)
	require.False(t, exists)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	// collection should exist
	collection, exists = app.NFTKeeper.GetCollection(ctx, denom)
	require.True(t, exists)
	require.NotEmpty(t, collection)
}
func TestGetCollections(t *testing.T) {
	app, ctx := createTestApp(false)

	// collections should be empty
	collections := app.NFTKeeper.GetCollections(ctx)
	require.Empty(t, collections)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	// collections should equal 1
	collections = app.NFTKeeper.GetCollections(ctx)
	require.NotEmpty(t, collections)
	require.Equal(t, len(collections), 1)
}
