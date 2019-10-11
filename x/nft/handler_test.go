package nft_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

const (
	module    = "module"
	denom     = "denom"
	nftID     = "nft-id"
	sender    = "sender"
	recipient = "recipient"
	tokenURI  = "token-uri"
)

func TestInvalidMsg(t *testing.T) {
	app, ctx := createTestApp(false)
	h := nft.GenericHandler(app.NFTKeeper)
	res := h(ctx, sdk.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized nft message type"))
}

func TestTransferNFTMsg(t *testing.T) {

	app, ctx := createTestApp(false)
	h := nft.GenericHandler(app.NFTKeeper)

	// An NFT to be transferred
	nft := types.NewBaseNFT(id, address, "TokenURI")

	// Define MsgTransferNft
	transferNftMsg := types.NewMsgTransferNFT(address, address2, denom, id)

	// handle should fail trying to transfer NFT that doesn't exist
	res := h(ctx, transferNftMsg)
	require.False(t, res.IsOK(), "%v", res)

	// Create token (collection and owner)
	app.NFTKeeper.MintNFT(ctx, denom, &nft)
	require.True(t, CheckInvariants(app.NFTKeeper, ctx))

	// handle should succeed when nft exists and is transferred by owner
	res = h(ctx, transferNftMsg)
	require.True(t, res.IsOK(), "%v", res)
	require.True(t, CheckInvariants(app.NFTKeeper, ctx))

	// event events should be emitted correctly
	for _, event := range res.Events {
		for _, attribute := range event.Attributes {
			value := string(attribute.Value)
			switch key := string(attribute.Key); key {
			case module:
				require.Equal(t, value, types.ModuleName)
			case denom:
				require.Equal(t, value, denom)
			case nftID:
				require.Equal(t, value, id)
			case sender:
				require.Equal(t, value, address.String())
			case recipient:
				require.Equal(t, value, address2.String())
			default:
				require.Fail(t, fmt.Sprintf("unrecognized event %s", key))
			}
		}
	}

	// nft should have been transferred as a result of the message
	nftAfterwards, err := app.NFTKeeper.GetNFT(ctx, denom, id)
	require.NoError(t, err)
	require.True(t, nftAfterwards.GetOwner().Equals(address2))

	transferNftMsg = types.NewMsgTransferNFT(address2, address3, denom, id)

	// handle should succeed when nft exists and is transferred by owner
	res = h(ctx, transferNftMsg)
	require.True(t, res.IsOK(), "%v", res)
	require.True(t, CheckInvariants(app.NFTKeeper, ctx))

	// Create token (collection and owner)
	app.NFTKeeper.MintNFT(ctx, denom2, &nft)
	require.True(t, CheckInvariants(app.NFTKeeper, ctx))

	transferNftMsg = types.NewMsgTransferNFT(address2, address3, denom2, id)

	// handle should succeed when nft exists and is transferred by owner
	res = h(ctx, transferNftMsg)
	require.True(t, res.IsOK(), "%v", res)
	require.True(t, CheckInvariants(app.NFTKeeper, ctx))
}

func TestEditNFTMetadataMsg(t *testing.T) {
	app, ctx := createTestApp(false)
	h := nft.GenericHandler(app.NFTKeeper)

	// An NFT to be edited
	nft := types.NewBaseNFT(id, address, tokenURI)

	// Create token (collection and address)
	app.NFTKeeper.MintNFT(ctx, denom, &nft)

	// Define MsgTransferNft
	failingEditNFTMetadata := types.NewMsgEditNFTMetadata(address, id, denom2, tokenURI2)

	res := h(ctx, failingEditNFTMetadata)
	require.False(t, res.IsOK(), "%v", res)

	// Define MsgTransferNft
	editNFTMetadata := types.NewMsgEditNFTMetadata(address, id, denom, tokenURI2)

	res = h(ctx, editNFTMetadata)
	require.True(t, res.IsOK(), "%v", res)

	// event events should be emitted correctly
	for _, event := range res.Events {
		for _, attribute := range event.Attributes {
			value := string(attribute.Value)
			switch key := string(attribute.Key); key {
			case module:
				require.Equal(t, value, types.ModuleName)
			case denom:
				require.Equal(t, value, denom)
			case nftID:
				require.Equal(t, value, id)
			case sender:
				require.Equal(t, value, address.String())
			case tokenURI:
				require.Equal(t, value, tokenURI2)
			default:
				require.Fail(t, fmt.Sprintf("unrecognized event %s", key))
			}
		}
	}

	nftAfterwards, err := app.NFTKeeper.GetNFT(ctx, denom, id)
	require.NoError(t, err)
	require.Equal(t, tokenURI2, nftAfterwards.GetTokenURI())

}

func TestMintNFTMsg(t *testing.T) {
	app, ctx := createTestApp(false)
	h := nft.GenericHandler(app.NFTKeeper)

	// Define MsgMintNFT
	mintNFT := types.NewMsgMintNFT(address, address, id, denom, tokenURI)

	// minting a token should succeed
	res := h(ctx, mintNFT)
	require.True(t, res.IsOK(), "%v", res)

	// event events should be emitted correctly
	for _, event := range res.Events {
		for _, attribute := range event.Attributes {
			value := string(attribute.Value)
			switch key := string(attribute.Key); key {
			case module:
				require.Equal(t, value, types.ModuleName)
			case denom:
				require.Equal(t, value, denom)
			case nftID:
				require.Equal(t, value, id)
			case sender:
				require.Equal(t, value, address.String())
			case recipient:
				require.Equal(t, value, address.String())
			case tokenURI:
				require.Equal(t, value, tokenURI)
			default:
				require.Fail(t, fmt.Sprintf("unrecognized event %s", key))
			}
		}
	}

	nftAfterwards, err := app.NFTKeeper.GetNFT(ctx, denom, id)

	require.NoError(t, err)
	require.Equal(t, tokenURI, nftAfterwards.GetTokenURI())

	// minting the same token should fail
	res = h(ctx, mintNFT)
	require.False(t, res.IsOK(), "%v", res)

	require.True(t, CheckInvariants(app.NFTKeeper, ctx))

}

func TestBurnNFTMsg(t *testing.T) {
	app, ctx := createTestApp(false)
	h := nft.GenericHandler(app.NFTKeeper)

	// An NFT to be burned
	nft := types.NewBaseNFT(id, address, tokenURI)

	// Create token (collection and address)
	app.NFTKeeper.MintNFT(ctx, denom, &nft)

	exists := app.NFTKeeper.IsNFT(ctx, denom, id)
	require.True(t, exists)

	// burning a non-existent NFT should fail
	failBurnNFT := types.NewMsgBurnNFT(address, id2, denom)
	res := h(ctx, failBurnNFT)
	require.False(t, res.IsOK(), "%s", res.Log)

	// NFT should still exist
	exists = app.NFTKeeper.IsNFT(ctx, denom, id)
	require.True(t, exists)

	// burning the NFt should succeed
	burnNFT := types.NewMsgBurnNFT(address, id, denom)

	res = h(ctx, burnNFT)
	require.True(t, res.IsOK(), "%v", res)

	// event events should be emitted correctly
	for _, event := range res.Events {
		for _, attribute := range event.Attributes {
			value := string(attribute.Value)
			switch key := string(attribute.Key); key {
			case module:
				require.Equal(t, value, types.ModuleName)
			case denom:
				require.Equal(t, value, denom)
			case nftID:
				require.Equal(t, value, id)
			case sender:
				require.Equal(t, value, address.String())
			default:
				require.Fail(t, fmt.Sprintf("unrecognized event %s", key))
			}
		}
	}

	// the NFT should not exist after burn
	exists = app.NFTKeeper.IsNFT(ctx, denom, id)
	require.False(t, exists)

	ownerReturned := app.NFTKeeper.GetOwner(ctx, address)
	require.Equal(t, 0, ownerReturned.Supply())

	require.True(t, CheckInvariants(app.NFTKeeper, ctx))
}
