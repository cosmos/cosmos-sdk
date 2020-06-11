package tx

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestTxWrapper(t *testing.T) {
	_, pubkey, addr := authtypes.KeyTestPubAddr()

	// TODO:
	// - verify that body and authInfo bytes encoded with DefaultTxEncoder and decoded with DefaultTxDecoder can be
	//   retrieved from GetBodyBytes and GetAuthInfoBytes
	// - create a TxWrapper using NewTxWrapper and:
	//   - verify that calling the SetBody results in the correct GetBodyBytes
	//   - verify that calling the SetAuthInfo results in the correct GetAuthInfoBytes and GetPubKeys
	//   - verify no nil panics
	appCodec := std.NewAppCodec(codec.New(), codectypes.NewInterfaceRegistry())
	tx := NewTxWrapper(appCodec.Marshaler, std.DefaultPublicKeyCodec{})

	cdc := std.DefaultPublicKeyCodec{}

	memo := "sometestmemo"
	msgs := []sdk.Msg{types.NewTestMsg(addr)}

	pk, err := cdc.Encode(pubkey)
	require.NoError(t, err)

	var signerInfo []*SignerInfo
	signerInfo = append(signerInfo, &SignerInfo{
		PublicKey: pk,
		ModeInfo: &ModeInfo{
			Sum: &ModeInfo_Single_{
				Single: &ModeInfo_Single{
					Mode: SignMode_SIGN_MODE_DIRECT,
				},
			},
		},
	})

	fee := Fee{Amount: sdk.NewCoins(sdk.NewInt64Coin("atom", 150)), GasLimit: 20000}

	tx.SetMsgs(msgs)
	require.Equal(t, len(msgs), len(tx.GetMsgs()))
	require.Equal(t, 0, len(tx.GetPubKeys()))

	tx.SetMemo(memo)
	tx.SetFee(fee.Amount)
	tx.SetGas(fee.GasLimit)

	tx.SetSignerInfos(signerInfo)
	require.Equal(t, len(msgs), len(tx.GetMsgs()))
	require.Equal(t, 1, len(tx.GetPubKeys()))
	require.Equal(t, pubkey.Bytes(), tx.GetPubKeys()[0].Bytes())
}
