package direct

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

func TestDirectModeHandler(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	var directModeHandler DirectModeHandler

	require.Equal(t, directModeHandler.DefaultMode(), types.SignMode_SIGN_MODE_DIRECT)
	require.Len(t, directModeHandler.Modes(), 1)

	signingData := signing.SignerData{
		ChainID:         "test-chain",
		AccountNumber:   1,
		AccountSequence: 1,
	}

	priv, _, addr := authtypes.KeyTestPubAddr()

	msgs := []sdk.Msg{authtypes.NewTestMsg(addr)}
	fee := authtypes.NewTestStdFee()
	privs, accNums, seqs := []crypto.PrivKey{priv}, []uint64{0}, []uint64{0}
	_ = authtypes.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)
	// signBytes, err := directModeHandler.GetSignBytes(types.SignMode_SIGN_MODE_DIRECT, signingData, tx)

	// fmt.Println("signBytes, err", string(signBytes), err)

	txWrapper := types.NewTxWrapper(&codec.ProtoCodec{}, std.DefaultPublicKeyCodec{})

	// setbody throwing error, since it is using any type
	txWrapper.SetBody(&types.TxBody{
		Messages: msgs,
	})

	// SetAuthInfo throwing error, since it is using any type
	// txWrapper.SetAuthInfo(&types.AuthInfo{
	// 	SignerInfos: &[]types.SignerInfo{
	// 		PublicKey: priv.PubKey(),
	// 		ModeInfo: &types.ModeInfo{
	// 			Sum: &types.ModeInfo_Single_{
	// 				Single: &types.ModeInfo_Single{Mode: types.SignMode_SIGN_MODE_DIRECT},
	// 			},
	// 		},
	// 	},
	// 	Fee: fee,
	// })

	signBytes, err := directModeHandler.GetSignBytes(types.SignMode_SIGN_MODE_DIRECT, signingData, txWrapper)

	fmt.Println("signBytes, err", string(signBytes), err)

	require.NoError(t, err)
	// TODO:
	// - verify DefaultMode and Modes
	// - verify GetSignBytes using a test transaction vs manually generating sign bytes by marshaling SignDoc
}
