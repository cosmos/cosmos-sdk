package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

func TestGetOwners(t *testing.T) {
	ctx, keeper, _ := Initialize()

	nft := types.NewBaseNFT(id, address, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	nft2 := types.NewBaseNFT(id2, address2, tokenURI)
	err = keeper.MintNFT(ctx, denom, &nft2)
	require.NoError(t, err)

	nft3 := types.NewBaseNFT(id3, address3, tokenURI)
	err = keeper.MintNFT(ctx, denom, &nft3)
	require.NoError(t, err)

	owners := keeper.GetOwners(ctx)
	require.Equal(t, 3, len(owners))

	nft = types.NewBaseNFT(id, address, tokenURI)
	err = keeper.MintNFT(ctx, denom2, &nft)
	require.NoError(t, err)

	nft2 = types.NewBaseNFT(id2, address2, tokenURI)
	err = keeper.MintNFT(ctx, denom2, &nft2)
	require.NoError(t, err)

	nft3 = types.NewBaseNFT(id3, address3, tokenURI)
	err = keeper.MintNFT(ctx, denom2, &nft3)
	require.NoError(t, err)

	owners = keeper.GetOwners(ctx)
	require.Equal(t, 3, len(owners))
}

func TestSetOwner(t *testing.T) {
	ctx, keeper, _ := Initialize()

	nft := types.NewBaseNFT(id, address, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	idCollection := types.NewIDCollection(denom, []string{id, id2, id3})
	owner := types.NewOwner(address, idCollection)

	oldOwner := keeper.GetOwner(ctx, address)

	keeper.SetOwner(ctx, owner)

	newOwner := keeper.GetOwner(ctx, address)
	require.NotEqual(t, oldOwner.String(), newOwner.String())
	require.Equal(t, owner.String(), newOwner.String())
}

func TestSetOwners(t *testing.T) {
	ctx, keeper, _ := Initialize()

	nft := types.NewBaseNFT(id, address, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	nft = types.NewBaseNFT(id2, address2, tokenURI)
	err = keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	idCollection := types.NewIDCollection(denom, []string{id, id2, id3})
	owner := types.NewOwner(address, idCollection)
	owner2 := types.NewOwner(address2, idCollection)

	oldOwner := keeper.GetOwner(ctx, address)
	oldOwner2 := keeper.GetOwner(ctx, address2)

	keeper.SetOwners(ctx, []types.Owner{owner, owner2})

	newOwner := keeper.GetOwner(ctx, address)
	require.NotEqual(t, oldOwner.String(), newOwner.String())
	require.Equal(t, owner.String(), newOwner.String())

	newOwner2 := keeper.GetOwner(ctx, address2)
	require.NotEqual(t, oldOwner2.String(), newOwner2.String())
	require.Equal(t, owner2.String(), newOwner2.String())
}

func TestSwapOwners(t *testing.T) {
	ctx, keeper, _ := Initialize()

	nft := types.NewBaseNFT(id, address, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	err = keeper.SwapOwners(ctx, denom, id, address, address2)
	require.NoError(t, err)

	err = keeper.SwapOwners(ctx, denom, id, address, address2)
	require.Error(t, err)

	err = keeper.SwapOwners(ctx, denom2, id, address, address2)
	require.Error(t, err)
}
