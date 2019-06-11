package nft

import (
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
	ctx, k := keeper.Initialize()
	h := NewHandler(k)
	res := h(ctx, sdk.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized nft message type"))
}

func TestTransferNFTMsg(t *testing.T) {
	denom := "some-denom"
	id := "somd-id"

	ctx, k := keeper.Initialize()
	addresses := keeper.CreateTestAddrs(2)
	originalOwner := addresses[0]
	nextOwner := addresses[1]
	h := NewHandler(k)

	// An NFT to be transferred
	nft := types.NewBaseNFT(
		id,
		originalOwner,
		"TokenURI",
		"Description",
		"ImageURI",
		"Name",
	)

	// Create token (collection and owner)
	k.MintNFT(ctx, denom, nft)

	// Define MsgTransferNft
	transferNftMsg := MsgTransferNFT{
		originalOwner,
		nextOwner,
		denom,
		id,
	}

	res := h(ctx, transferNftMsg)
	require.True(t, res.IsOK())

	nftAfterwards, err := k.GetNFT(ctx, denom, id)
	require.Nil(t, err)

	require.True(t, nftAfterwards.GetOwner().Equals(addresses[1]))

}
