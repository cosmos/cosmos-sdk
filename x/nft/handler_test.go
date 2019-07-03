package nft

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
	"github.com/stretchr/testify/require"
)

func TestInvalidMsg(t *testing.T) {
	ctx, k, _ := keeper.Initialize()
	h := GenericHandler(k)
	res := h(ctx, sdk.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized nft message type"))
}

func TestTransferNFTMsg(t *testing.T) {

	ctx, k, _ := keeper.Initialize()
	h := GenericHandler(k)

	// An NFT to be transferred
	nft := types.NewBaseNFT(
		id,
		address,
		"Name",
		"Description",
		"ImageURI",
		"TokenURI",
	)

	// Define MsgTransferNft
	transferNftMsg := types.NewMsgTransferNFT(
		address,
		address2,
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

	// event events should be emmitted correctly
	for _, event := range res.Events {
		for _, attribute := range event.Attributes {
			value := string(attribute.Value)
			switch key := string(attribute.Key); key {
			case "denom":
				require.Equal(t, value, denom)
			case "nft-id":
				require.Equal(t, value, id)
			case "sender":
				require.Equal(t, value, address.String())
			case "recipient":
				require.Equal(t, value, address2.String())
			default:
				require.Fail(t, "unrecognized event %s", key)
			}
		}
	}

	// nft should have been transferred as a result of the message
	nftAfterwards, err := k.GetNFT(ctx, denom, id)
	require.Nil(t, err)
	require.True(t, nftAfterwards.GetOwner().Equals(address2))

	// handle should fail when nft exists and is transferred by not owner
	res = h(ctx, transferNftMsg)
	require.False(t, res.IsOK(), "%v", res)

}

func TestEditNFTMetadataMsg(t *testing.T) {

	ctx, k, _ := keeper.Initialize()

	h := GenericHandler(k)

	// An NFT to be edited
	nft := types.NewBaseNFT(
		id,
		address,
		name,
		description,
		image,
		tokenURI,
	)

	// Create token (collection and address)
	k.MintNFT(ctx, denom, nft)

	// Define MsgTransferNft
	failingEditNFTMetadata := types.NewMsgEditNFTMetadata(
		address,
		id,
		denom2,
		name2,
		description2,
		image2,
		tokenURI2,
	)

	res := h(ctx, failingEditNFTMetadata)
	require.False(t, res.IsOK(), "%v", res)

	// Define MsgTransferNft
	failingEditNFTMetadata = types.NewMsgEditNFTMetadata(
		address2,
		id,
		denom,
		name2,
		description2,
		image2,
		tokenURI2,
	)

	res = h(ctx, failingEditNFTMetadata)
	require.False(t, res.IsOK(), "%v", res)

	// Define MsgTransferNft
	editNFTMetadata := types.NewMsgEditNFTMetadata(
		address,
		id,
		denom,
		name2,
		description2,
		image2,
		tokenURI2,
	)

	res = h(ctx, editNFTMetadata)
	require.True(t, res.IsOK(), "%v", res)

	// event events should be emmitted correctly
	for _, event := range res.Events {
		for _, attribute := range event.Attributes {
			value := string(attribute.Value)
			switch key := string(attribute.Key); key {
			case "denom":
				require.Equal(t, value, denom)
			case "nft-id":
				require.Equal(t, value, id)
			case "nft-name":
				require.Equal(t, value, name2)
			case "nft-description":
				require.Equal(t, value, description2)
			case "nft-image":
				require.Equal(t, value, image2)
			case "nft-tokenURI":
				require.Equal(t, value, tokenURI2)
			case "sender":
				require.Equal(t, value, address.String())
			default:
				require.Fail(t, "unrecognized event %s", key)
			}
		}
	}

	nftAfterwards, err := k.GetNFT(ctx, denom, id)
	require.Nil(t, err)
	require.Equal(t, name2, nftAfterwards.GetName())
	require.Equal(t, description2, nftAfterwards.GetDescription())
	require.Equal(t, image2, nftAfterwards.GetImage())
	require.Equal(t, tokenURI2, nftAfterwards.GetTokenURI())

}

func TestMintNFTMsg(t *testing.T) {
	ctx, k, _ := keeper.Initialize()

	h := GenericHandler(k)

	// Define MsgMintNFT
	mintNFT := types.NewMsgMintNFT(
		address,
		address,
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

	// event events should be emmitted correctly
	for _, event := range res.Events {
		for _, attribute := range event.Attributes {
			value := string(attribute.Value)
			switch key := string(attribute.Key); key {
			case "denom":
				require.Equal(t, value, denom)
			case "nft-id":
				require.Equal(t, value, id)
			case "nft-name":
				require.Equal(t, value, name)
			case "nft-description":
				require.Equal(t, value, description)
			case "nft-image":
				require.Equal(t, value, image)
			case "nft-tokenURI":
				require.Equal(t, value, tokenURI)
			case "sender":
				require.Equal(t, value, address.String())
			case "recipient":
				require.Equal(t, value, address.String())
			default:
				require.Fail(t, "unrecognized event %s", key)
			}
		}
	}

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

	ctx, k, _ := keeper.Initialize()
	h := GenericHandler(k)

	// An NFT to be burned
	nft := types.NewBaseNFT(
		id,
		address,
		name,
		description,
		image,
		tokenURI,
	)

	// Create token (collection and address)
	k.MintNFT(ctx, denom, nft)

	exists := k.IsNFT(ctx, denom, id)
	require.True(t, exists)

	// burning an NFT without being the address should fail
	failBurnNFT := types.NewMsgBurnNFT(
		address2,
		id,
		denom,
	)
	res := h(ctx, failBurnNFT)
	require.False(t, res.IsOK(), "%s", res.Log)

	// burning a non-existant NFT should fail
	failBurnNFT = types.NewMsgBurnNFT(
		address,
		id2,
		denom,
	)
	res = h(ctx, failBurnNFT)
	require.False(t, res.IsOK(), "%s", res.Log)

	// NFT should still exist
	exists = k.IsNFT(ctx, denom, id)
	require.True(t, exists)

	// burning the NFt should succeed
	burnNFT := types.NewMsgBurnNFT(
		address,
		id,
		denom,
	)

	res = h(ctx, burnNFT)
	require.True(t, res.IsOK(), "%v", res)

	// event events should be emmitted correctly
	for _, event := range res.Events {
		for _, attribute := range event.Attributes {
			value := string(attribute.Value)
			switch key := string(attribute.Key); key {
			case "denom":
				require.Equal(t, value, denom)
			case "nft-id":
				require.Equal(t, value, id)
			case "sender":
				require.Equal(t, value, address.String())
			default:
				require.Fail(t, "unrecognized event %s", key)
			}
		}
	}

	// the NFT should not exist after burn
	exists = k.IsNFT(ctx, denom, id)
	require.False(t, exists)

	ownerReturned := k.GetOwner(ctx, address)
	require.Equal(t, 0, ownerReturned.Supply())
}
