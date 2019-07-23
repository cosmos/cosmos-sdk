package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

func TestMintNFT(t *testing.T) {
	ctx, keeper, _ := Initialize()

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	// MintNFT shouldn't fail when collection exists
	nft2 := types.NewBaseNFT(id2, address, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom, &nft2)
	require.NoError(t, err)
}

func TestGetNFT(t *testing.T) {
	ctx, keeper, _ := Initialize()

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	// GetNFT should get the NFT
	receivedNFT, err := keeper.GetNFT(ctx, denom, id)
	require.NoError(t, err)
	require.Equal(t, receivedNFT.GetID(), id)
	require.True(t, receivedNFT.GetOwner().Equals(address))
	require.Equal(t, receivedNFT.GetTokenURI(), tokenURI)
	require.Equal(t, receivedNFT.GetDescription(), description)
	require.Equal(t, receivedNFT.GetImage(), image)
	require.Equal(t, receivedNFT.GetName(), name)

	// MintNFT shouldn't fail when collection exists
	nft2 := types.NewBaseNFT(id2, address, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom, &nft2)
	require.NoError(t, err)

	// GetNFT should get the NFT when collection exists
	receivedNFT2, err := keeper.GetNFT(ctx, denom, id2)
	require.NoError(t, err)
	require.Equal(t, receivedNFT2.GetID(), id2)
	require.True(t, receivedNFT2.GetOwner().Equals(address))
	require.Equal(t, receivedNFT2.GetTokenURI(), tokenURI)
	require.Equal(t, receivedNFT2.GetDescription(), description)
	require.Equal(t, receivedNFT2.GetImage(), image)
	require.Equal(t, receivedNFT2.GetName(), name)

}

func TestUpdateNFT(t *testing.T) {
	ctx, keeper, _ := Initialize()

	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)

	// UpdateNFT should fail when NFT doesn't exists
	err := keeper.UpdateNFT(ctx, denom, &nft)
	require.Error(t, err)

	// MintNFT shouldn't fail when collection does not exist
	err = keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	nonnft := types.NewBaseNFT(id2, address, name, description, image, tokenURI)
	// UpdateNFT should fail when NFT doesn't exists
	err = keeper.UpdateNFT(ctx, denom, &nonnft)
	require.Error(t, err)

	// UpdateNFT shouldn't fail when NFT exists
	nft2 := types.NewBaseNFT(id, address, name, description, image, tokenURI2)
	err = keeper.UpdateNFT(ctx, denom, &nft2)
	require.NoError(t, err)

	// UpdateNFT shouldn't fail when NFT exists
	nft2 = types.NewBaseNFT(id, address2, name, description, image, tokenURI2)
	err = keeper.UpdateNFT(ctx, denom, &nft2)
	require.NoError(t, err)

	// GetNFT should get the NFT with new tokenURI
	receivedNFT, err := keeper.GetNFT(ctx, denom, id)
	require.NoError(t, err)
	require.Equal(t, receivedNFT.GetTokenURI(), tokenURI2)

}

func TestDeleteNFT(t *testing.T) {
	ctx, keeper, _ := Initialize()

	// DeleteNFT should fail when NFT doesn't exist and collection doesn't exist
	err := keeper.DeleteNFT(ctx, denom, id)
	require.Error(t, err)

	// MintNFT should not fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	// DeleteNFT should fail when NFT doesn't exist but collection does exist
	err = keeper.DeleteNFT(ctx, denom, id2)
	require.Error(t, err)

	// DeleteNFT should not fail when NFT and collection exist
	err = keeper.DeleteNFT(ctx, denom, id)
	require.NoError(t, err)

	// NFT should no longer exist
	isNFT := keeper.IsNFT(ctx, denom, id)
	require.False(t, isNFT)

	owner := keeper.GetOwner(ctx, address)
	require.Equal(t, 0, owner.Supply())
}

func TestIsNFT(t *testing.T) {
	ctx, keeper, _ := Initialize()

	// IsNFT should return false
	isNFT := keeper.IsNFT(ctx, denom, id)
	require.False(t, isNFT)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, &nft)
	require.NoError(t, err)

	// IsNFT should return true
	isNFT = keeper.IsNFT(ctx, denom, id)
	require.True(t, isNFT)
}
