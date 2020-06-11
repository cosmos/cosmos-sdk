package tx

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	priv = ed25519.GenPrivKey()
	addr = sdk.AccAddress(priv.PubKey().Address())
)

func TestTxWrapper(t *testing.T) {
	// TODO:
	// - verify that body and authInfo bytes encoded with DefaultTxEncoder and decoded with DefaultTxDecoder can be
	//   retrieved from GetBodyBytes and GetAuthInfoBytes
	// - create a TxWrapper using NewTxWrapper and:
	//   - verify that calling the SetBody results in the correct GetBodyBytes
	//   - verify that calling the SetAuthInfo results in the correct GetAuthInfoBytes and GetPubKeys
	//   - verify no nil panics
	cdc := std.NewAppCodec(codec.New(), codectypes.NewInterfaceRegistry())

	tx := NewTxWrapper(cdc.Marshaler, std.DefaultPublicKeyCodec{})

	memo := "sometestmemo"
	msgs := []sdk.Msg{types.NewTestMsg(addr)}

	txBody := TxBody{
		Messages: msgs,
		Memo:     memo,
	}

	authInfo := AuthInfo{
		SignerInfos: make([]*SignerInfo, 0),
		Fee:         &Fee{Amount: sdk.NewCoins(sdk.NewCoin("test, 0")), GasLimit: 20000},
	}

	tx.SetBody(&txBody)

	tx.SetAuthInfo(&authInfo)

	bodyBz := tx.GetBodyBytes()
}
