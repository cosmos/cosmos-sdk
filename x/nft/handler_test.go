package nft

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/require"
)

// TODO: finish

// type MsgTransferNFT struct {
// 	Sender    sdk.AccAddress
// 	Recipient sdk.AccAddress
// 	Denom     string
// 	ID        uint64
// }

func createTestAddrs(numAddrs int) []sdk.AccAddress {
	var addresses []sdk.AccAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (numAddrs + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") //base address string

		buffer.WriteString(numString) //adding on final two digits to make addresses unique
		res, _ := sdk.AccAddressFromHex(buffer.String())
		bech := res.String()
		addresses = append(addresses, testAddr(buffer.String(), bech))
		buffer.Reset()
	}
	return addresses
}


func createInput() (k keeper, addrs []sdk.AccAddress) {
	cdc := codec.New()
	types.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	
	addrs = createTestAddrs(2)
	keyNFT := sdk.NewKVStoreKey(StoreKey)
	
	k = keeper.NewKeeper(storeKey, cdc)
	return
}
func TestInvalidMsg(t *testing.T) {
	h := NewHandler(nil)

	res := h(sdk.Context{}, sdk.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized nft message type"))
}

func TestTransferNFTMsg(t *testing.T) {
	nftK, addresses := CreateTestInput(t)

	h := NewHandler(nftK)

	// An NFT to be transferred
	nfts := types.NewNFTs(
		types.NewBaseNFT(
			"hello",
			addresses[0].Address()
			"Berlin",
			"Berlin NFT",
			"ImageURI",
			"TokenURI"
		)
)

	require.equal(t, nft)

	// Create collection 
	nftK.SetCollection(cdc, "Williams", nfts)

	// Define MsgTransferNft
	transferNft := MsgTransferNFT{
		addresses[0].GetAddress(),
		addresses[1].GetAddress(),
		"Williams"
		"hello"
	}

}
