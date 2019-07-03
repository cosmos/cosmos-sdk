package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
	"github.com/stretchr/testify/require"
)

func TestMintNFT(t *testing.T) {
	ctx, keeper, _ := Initialize()

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	// MintNFT shouldn't fail when collection exists
	nft2 := types.NewBaseNFT(id2, address, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom, nft2)
	require.Nil(t, err)
}

func TestGetNFT(t *testing.T) {
	ctx, keeper, _ := Initialize()

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	// GetNFT should get the NFT
	var receivedNFT types.NFT
	receivedNFT, err = keeper.GetNFT(ctx, denom, id)
	require.Nil(t, err)
	require.Equal(t, receivedNFT.GetID(), id)
	require.True(t, receivedNFT.GetOwner().Equals(address))
	require.Equal(t, receivedNFT.GetTokenURI(), tokenURI)
	require.Equal(t, receivedNFT.GetDescription(), description)
	require.Equal(t, receivedNFT.GetImage(), image)
	require.Equal(t, receivedNFT.GetName(), name)

	// MintNFT shouldn't fail when collection exists
	nft2 := types.NewBaseNFT(id2, address, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom, nft2)
	require.Nil(t, err)

	// GetNFT should get the NFT when collection exists
	var receivedNFT2 types.NFT
	receivedNFT2, err = keeper.GetNFT(ctx, denom, id2)
	require.Nil(t, err)
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
	err := keeper.UpdateNFT(ctx, denom, nft)
	require.NotNil(t, err)

	// MintNFT shouldn't fail when collection does not exist
	err = keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	nonnft := types.NewBaseNFT(id2, address, name, description, image, tokenURI)
	// UpdateNFT should fail when NFT doesn't exists
	err = keeper.UpdateNFT(ctx, denom, nonnft)
	require.NotNil(t, err)

	// UpdateNFT shouldn't fail when NFT exists
	nft2 := types.NewBaseNFT(id, address, name, description, image, tokenURI2)
	err = keeper.UpdateNFT(ctx, denom, nft2)
	require.Nil(t, err)

	// UpdateNFT shouldn't fail when NFT exists
	nft2 = types.NewBaseNFT(id, address2, name, description, image, tokenURI2)
	err = keeper.UpdateNFT(ctx, denom, nft2)
	require.Nil(t, err)

	// GetNFT should get the NFT with new tokenURI
	var receivedNFT types.NFT
	receivedNFT, err = keeper.GetNFT(ctx, denom, id)
	require.Nil(t, err)
	require.Equal(t, receivedNFT.GetTokenURI(), tokenURI2)

}

func TestDeleteNFT(t *testing.T) {
	ctx, keeper, _ := Initialize()

	// DeleteNFT should fail when NFT doesn't exist and collection doesn't exist
	err := keeper.DeleteNFT(ctx, denom, id)
	require.NotNil(t, err)

	// MintNFT should not fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	// DeleteNFT should fail when NFT doesn't exist but collection does exist
	err = keeper.DeleteNFT(ctx, denom, id2)
	require.NotNil(t, err)

	// DeleteNFT should not fail when NFT and collection exist
	err = keeper.DeleteNFT(ctx, denom, id)
	require.Nil(t, err)

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
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	// IsNFT should return true
	isNFT = keeper.IsNFT(ctx, denom, id)
	require.True(t, isNFT)
}

func TestSetCollection(t *testing.T) {
	ctx, keeper, _ := Initialize()

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	// collection should exist
	collection, exists := keeper.GetCollection(ctx, denom)
	require.True(t, exists)

	nft2 := types.NewBaseNFT(id2, address, name, description, image, tokenURI)
	collection, err = collection.AddNFT(nft2)
	require.Nil(t, err)
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
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

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
	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	// collections should equal 1
	collections = keeper.GetCollections(ctx)
	require.NotEmpty(t, collections)
	require.Equal(t, len(collections), 1)
}

func TestSwapOwners(t *testing.T) {
	ctx, keeper, _ := Initialize()

	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	err = keeper.SwapOwners(ctx, denom, id, address, address2)
	require.Nil(t, err)

	err = keeper.SwapOwners(ctx, denom, id, address, address2)
	require.NotNil(t, err)

	err = keeper.SwapOwners(ctx, denom2, id, address, address2)
	require.NotNil(t, err)
}

func TestGetOwners(t *testing.T) {
	ctx, keeper, _ := Initialize()

	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	nft2 := types.NewBaseNFT(id2, address2, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom, nft2)
	require.Nil(t, err)

	nft3 := types.NewBaseNFT(id3, address3, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom, nft3)
	require.Nil(t, err)

	owners := keeper.GetOwners(ctx)
	require.Equal(t, 3, len(owners))

	nft = types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom2, nft)
	require.Nil(t, err)

	nft2 = types.NewBaseNFT(id2, address2, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom2, nft2)
	require.Nil(t, err)

	nft3 = types.NewBaseNFT(id3, address3, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom2, nft3)
	require.Nil(t, err)
	owners = keeper.GetOwners(ctx)
	require.Equal(t, 3, len(owners))
}

func TestSetOwner(t *testing.T) {
	ctx, keeper, _ := Initialize()

	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

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

	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	nft = types.NewBaseNFT(id2, address2, name, description, image, tokenURI)
	err = keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

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
