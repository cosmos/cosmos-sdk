package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

func TestGetOwners(t *testing.T) {
	app, ctx := createTestApp(false)

	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	nft2 := types.NewBaseNFT(id2, address2, tokenURI)
	err = app.NFTKeeper.MintNFT(ctx, denom, &nft2)
	require.NoError(t, err)

	nft3 := types.NewBaseNFT(id3, address3, tokenURI)
	err = app.NFTKeeper.MintNFT(ctx, denom, &nft3)
	require.NoError(t, err)

	owners := app.NFTKeeper.GetOwners(ctx)
	require.Equal(t, 3, len(owners))

	nft = types.NewBaseNFT(id, address, tokenURI)
	err = app.NFTKeeper.MintNFT(ctx, denom2, &nft)
	require.NoError(t, err)

	nft2 = types.NewBaseNFT(id2, address2, tokenURI)
	err = app.NFTKeeper.MintNFT(ctx, denom2, &nft2)
	require.NoError(t, err)

	nft3 = types.NewBaseNFT(id3, address3, tokenURI)
	err = app.NFTKeeper.MintNFT(ctx, denom2, &nft3)
	require.NoError(t, err)

	owners = app.NFTKeeper.GetOwners(ctx)
	require.Equal(t, 3, len(owners))
}

func TestSetOwner(t *testing.T) {
	app, ctx := createTestApp(false)

	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	idCollection := types.NewIDCollection(denom, []string{id, id2, id3})
	owner := types.NewOwner(address, idCollection)

	oldOwner := app.NFTKeeper.GetOwner(ctx, address)

	app.NFTKeeper.SetOwner(ctx, owner)

	newOwner := app.NFTKeeper.GetOwner(ctx, address)
	require.NotEqual(t, oldOwner.String(), newOwner.String())
	require.Equal(t, owner.String(), newOwner.String())
}

func TestSetOwners(t *testing.T) {
	app, ctx := createTestApp(false)

	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	nft = types.NewBaseNFT(id2, address2, tokenURI)
	err = app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	idCollection := types.NewIDCollection(denom, []string{id, id2, id3})
	owner := types.NewOwner(address, idCollection)
	owner2 := types.NewOwner(address2, idCollection)

	oldOwner := app.NFTKeeper.GetOwner(ctx, address)
	oldOwner2 := app.NFTKeeper.GetOwner(ctx, address2)

	app.NFTKeeper.SetOwners(ctx, []types.Owner{owner, owner2})

	newOwner := app.NFTKeeper.GetOwner(ctx, address)
	require.NotEqual(t, oldOwner.String(), newOwner.String())
	require.Equal(t, owner.String(), newOwner.String())

	newOwner2 := app.NFTKeeper.GetOwner(ctx, address2)
	require.NotEqual(t, oldOwner2.String(), newOwner2.String())
	require.Equal(t, owner2.String(), newOwner2.String())
}

func TestSwapOwners(t *testing.T) {
	app, ctx := createTestApp(false)

	nft := types.NewBaseNFT(id, address, tokenURI)
	err := app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	err = app.NFTKeeper.SwapOwners(ctx, denom, id, address, address2)
	require.NoError(t, err)

	err = app.NFTKeeper.SwapOwners(ctx, denom, id, address, address2)
	require.Error(t, err)

	err = app.NFTKeeper.SwapOwners(ctx, denom2, id, address, address2)
	require.Error(t, err)
}
