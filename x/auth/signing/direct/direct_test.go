package direct

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
)

func TestDirectModeHandler(t *testing.T) {
	app := simapp.Setup(false)
	//ctx := app.BaseApp.NewContext(false, abci.Header{})

	_, pubkey, addr := authtypes.KeyTestPubAddr()
	cdc := std.DefaultPublicKeyCodec{}

	tx := types.NewTxWrapper(app.AppCodec().Marshaler, std.DefaultPublicKeyCodec{})
	memo := "sometestmemo"
	msgs := []sdk.Msg{authtypes.NewTestMsg(addr)}

	pk, err := cdc.Encode(pubkey)
	require.NoError(t, err)

	var signerInfo []*types.SignerInfo
	signerInfo = append(signerInfo, &types.SignerInfo{
		PublicKey: pk,
		ModeInfo: &types.ModeInfo{
			Sum: &types.ModeInfo_Single_{
				Single: &types.ModeInfo_Single{
					Mode: types.SignMode_SIGN_MODE_DIRECT,
				},
			},
		},
	})

	fee := types.Fee{Amount: sdk.NewCoins(sdk.NewInt64Coin("atom", 150)), GasLimit: 20000}

	tx.SetMsgs(msgs)
	tx.SetMemo(memo)
	tx.SetFee(fee)
	tx.SetSignerInfos(signerInfo)

	var directModeHandler DirectModeHandler
	require.Equal(t, directModeHandler.DefaultMode(), types.SignMode_SIGN_MODE_DIRECT)
	require.Len(t, directModeHandler.Modes(), 1)

	signingData := signing.SignerData{
		ChainID:         "test-chain",
		AccountNumber:   1,
		AccountSequence: 1,
	}

	signBytes, err := directModeHandler.GetSignBytes(types.SignMode_SIGN_MODE_DIRECT, signingData, tx)

	require.NoError(t, err)
	require.NotNil(t, signBytes)

	signDoc := types.SignDoc{
		AccountNumber: 1,
		AccountSequence: 1,
		AuthInfoBytes: tx.GetAuthInfoBytes(),
		BodyBytes: tx.GetBodyBytes(),
		ChainId: "test-chain",
	}

	signDocBytes, err := signDoc.Marshal()
	require.NoError(t, err)
	require.Equal(t, signDocBytes, signBytes)

	signDoc.BodyBytes = []byte("dfafdasfds")
	signDocBytes, err = signDoc.Marshal()
	require.NoError(t, err)
	require.NotEqual(t, signDocBytes, signBytes)
}
