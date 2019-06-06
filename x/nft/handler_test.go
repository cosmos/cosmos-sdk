package nft

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	// "github.com/stretchr/testify/require"
)

// TODO: finish

// type MsgTransferNFT struct {
// 	Sender    sdk.AccAddress
// 	Recipient sdk.AccAddress
// 	Denom     string
// 	ID        uint64
// }
func TestTransferNFTMsg(t *testing.T) {
	ctx, ak, nftK := CreateTestInput(t)

	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.ZeroInt()))
	accounts := createTestAccs(ctx, 2, coins, &ak)

	nft := nftk.New

	transferNft := MsgTransferNFT{
		accounts[0].GetAddress(),
		accounts[1].GetAddress(),
	}

}
