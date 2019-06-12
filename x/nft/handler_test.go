package nft

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
	"github.com/stretchr/testify/require"
)

func createInput() (k Keeper, addrs []sdk.AccAddress) {
	cdc := codec.New()
	types.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	addrs = keeper.CreateTestAddrs(2)
	keyNFT := sdk.NewKVStoreKey(StoreKey)

	k = keeper.NewKeeper(cdc, keyNFT)
	return
}
func TestInvalidMsg(t *testing.T) {
	ctx, k, _ := keeper.Initialize()
	h := NewHandler(k)
	res := h(ctx, sdk.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized nft message type"))
}

func TestTransferNFTMsg(t *testing.T) {
	denom := "some-denom"
	id := "somd-id"

	ctx, k, _ := keeper.Initialize()
	addresses := keeper.CreateTestAddrs(2)
	originalOwner := addresses[0]
	nextOwner := addresses[1]
	h := NewHandler(k)

	// An NFT to be transferred
	nft := types.NewBaseNFT(
		id,
		originalOwner,
		"Name",
		"Description",
		"ImageURI",
		"TokenURI",
	)

	// Define MsgTransferNft
	transferNftMsg := types.NewMsgTransferNFT(
		originalOwner,
		nextOwner,
		denom,
		id,
	)

	// handle should fail trying to transfer NFT that doesn't exist
	res := h(ctx, transferNftMsg)
	require.False(t, res.IsOK(), "%v", res)

	// Create token (collection and owner)
	k.MintNFT(ctx, denom, nft)

	// handle should succeed when nft exists and is transferred by owner
	res = h(ctx, transferNftMsg)
	require.True(t, res.IsOK(), "%v", res)

	// nft should have been transferred as a result of the message
	nftAfterwards, err := k.GetNFT(ctx, denom, id)
	require.Nil(t, err)
	require.True(t, nftAfterwards.GetOwner().Equals(addresses[1]))

	// handle should fail when nft exists and is transferred by not owner
	res = h(ctx, transferNftMsg)
	require.False(t, res.IsOK(), "%v", res)

	tags := res.Tags
	fmt.Println(tags)
	// TODO: check for proper tags
}

func TestEditNFTMetadataMsg(t *testing.T) {
	denom := "some-denom"
	id := "somd-id"
	name := "Name"
	tokenURI := "TokenURI"
	description := "Description"
	image := "ImageURI"

	_name := "Name2"
	_tokenURI := "TokenURI2"
	_description := "Description2"
	_image := "ImageURI2"

	ctx, k, _ := keeper.Initialize()
	addresses := keeper.CreateTestAddrs(2)
	owner := addresses[0]

	h := NewHandler(k)

	// An NFT to be edited
	nft := types.NewBaseNFT(
		id,
		owner,
		name,
		description,
		image,
		tokenURI,
	)

	// Create token (collection and owner)
	k.MintNFT(ctx, denom, nft)

	// Define MsgTransferNft
	editNFTMetadata := types.NewMsgEditNFTMetadata(
		owner,
		id,
		denom,
		_name,
		_description,
		_image,
		_tokenURI,
	)

	res := h(ctx, editNFTMetadata)
	require.True(t, res.IsOK(), "%v", res)

	nftAfterwards, err := k.GetNFT(ctx, denom, id)
	require.Nil(t, err)
	require.Equal(t, _name, nftAfterwards.GetName())
	require.Equal(t, _description, nftAfterwards.GetDescription())
	require.Equal(t, _image, nftAfterwards.GetImage())
	require.Equal(t, _tokenURI, nftAfterwards.GetTokenURI())

}

func TestMintNFTMsg(t *testing.T) {
	denom := "some-denom"
	id := "somd-id"
	name := "Name"
	tokenURI := "TokenURI"
	description := "Description"
	image := "ImageURI"

	ctx, k, _ := keeper.Initialize()
	addresses := keeper.CreateTestAddrs(2)
	owner := addresses[0]

	h := NewHandler(k)

	// Define MsgMintNFT
	mintNFT := types.NewMsgMintNFT(
		owner,
		owner,
		id,
		denom,
		name,
		description,
		image,
		tokenURI,
	)

	// minting a token should succeed
	res := h(ctx, mintNFT)
	require.True(t, res.IsOK(), "%v", res)

	nftAfterwards, err := k.GetNFT(ctx, denom, id)

	require.Nil(t, err)
	require.Equal(t, name, nftAfterwards.GetName())
	require.Equal(t, description, nftAfterwards.GetDescription())
	require.Equal(t, image, nftAfterwards.GetImage())
	require.Equal(t, tokenURI, nftAfterwards.GetTokenURI())

	// minting the same token should fail
	res = h(ctx, mintNFT)
	require.False(t, res.IsOK(), "%v", res)

}

func TestBurnNFTMsg(t *testing.T) {
	denom := "some-denom"
	id := "somd-id"
	name := "Name"
	tokenURI := "TokenURI"
	description := "Description"
	image := "ImageURI"

	ctx, k, _ := keeper.Initialize()
	addresses := keeper.CreateTestAddrs(2)
	owner := addresses[0]
	notOwner := addresses[1]

	h := NewHandler(k)

	// An NFT to be burned
	nft := types.NewBaseNFT(
		id,
		owner,
		name,
		description,
		image,
		tokenURI,
	)

	// Create token (collection and owner)
	k.MintNFT(ctx, denom, nft)

	exists := k.IsNFT(ctx, denom, id)
	require.True(t, exists)

	// burning an NFT without being the owner should fail
	failBurnNFT := types.NewMsgBurnNFT(
		notOwner,
		id,
		denom,
	)
	res := h(ctx, failBurnNFT)
	require.False(t, res.IsOK(), "%s", res.Log)

	// NFT should still exist
	exists = k.IsNFT(ctx, denom, id)
	require.True(t, exists)

	// burning the NFt should succeed
	burnNFT := types.NewMsgBurnNFT(
		owner,
		id,
		denom,
	)

	res = h(ctx, burnNFT)
	require.True(t, res.IsOK(), "%v", res)

	// the NFT should not exist after burn
	exists = k.IsNFT(ctx, denom, id)
	require.False(t, exists)
}
