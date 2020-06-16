package types

import (
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//var (
//	priv = ed25519.GenPrivKey()
//	addr = sdk.AccAddress(priv.PubKey().Address())
//)

func initTxBuilder () TxBuilder {
	return NewTxBuilder(
		DefaultTxEncoder(makeCodec()), 1, 1,
		0, 0, false,
		"foo-chain", "", nil, nil,
	)
}

func makeCodec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	RegisterCodec(cdc)
	cdc.RegisterConcrete(sdk.TestMsg{}, "cosmos-sdk/Test", nil)
	return cdc
}

func initTxBuilderWithKeybase(t *testing.T, from string) (TxBuilder, keyring.Info) {
	// Now add a temporary keybase
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	kr, err := keyring.New(t.Name(), "test", dir, nil)
	require.NoError(t, err)
	path := hd.CreateHDPath(118, 0, 0).String()

	_, seed, err := kr.NewMnemonic(from, keyring.English, path, hd.Secp256k1)
	require.NoError(t, err)
	require.NoError(t, kr.Delete(from))

	info, err := kr.NewAccount(from, seed, "", path, hd.Secp256k1)
	require.NoError(t, err)

	return NewTxBuilder(
		DefaultTxEncoder(makeCodec()), 1, 1,
		0, 0, false,
		"foo-chain", "", nil, nil,
	).WithKeybase(kr), info
}

func TestTxBuilderBuild(t *testing.T) {
	type fields struct {
		TxEncoder     sdk.TxEncoder
		AccountNumber uint64
		Sequence      uint64
		Gas           uint64
		GasAdjustment float64
		SimulateGas   bool
		ChainID       string
		Memo          string
		Fees          sdk.Coins
		GasPrices     sdk.DecCoins
	}
	defaultMsg := []sdk.Msg{sdk.NewTestMsg(addr)}
	tests := []struct {
		name    string
		fields  fields
		msgs    []sdk.Msg
		want    StdSignMsg
		wantErr bool
	}{
		{
			"builder without fees and gas",
			fields{
				TxEncoder:     DefaultTxEncoder(codec.New()),
				AccountNumber: 1,
				Sequence:      1,
				SimulateGas:   false,
				ChainID:       "test-chain",
				Memo:          "hello from Voyager 1!",
			},
			defaultMsg,
			StdSignMsg{
				ChainID:       "test-chain",
				AccountNumber: 1,
				Sequence:      1,
				Memo:          "hello from Voyager 1!",
				Msgs:          defaultMsg,
				Fee:           NewStdFee(0, nil),
			},
			false,
		},
		{
			"builder with fees",
			fields{
				TxEncoder:     DefaultTxEncoder(codec.New()),
				AccountNumber: 1,
				Sequence:      1,
				Gas:           200000,
				GasAdjustment: 1.1,
				SimulateGas:   false,
				ChainID:       "test-chain",
				Memo:          "hello from Voyager 1!",
				Fees:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))),
			},
			defaultMsg,
			StdSignMsg{
				ChainID:       "test-chain",
				AccountNumber: 1,
				Sequence:      1,
				Memo:          "hello from Voyager 1!",
				Msgs:          defaultMsg,
				Fee:           NewStdFee(200000, sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))}),
			},
			false,
		},
		{
			"builder with gas prices",
			fields{
				TxEncoder:     DefaultTxEncoder(codec.New()),
				AccountNumber: 1,
				Sequence:      1,
				Gas:           200000,
				GasAdjustment: 1.1,
				SimulateGas:   false,
				ChainID:       "test-chain",
				Memo:          "hello from Voyager 2!",
				GasPrices:     sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10000, sdk.Precision))},
			},
			defaultMsg,
			StdSignMsg{
				ChainID:       "test-chain",
				AccountNumber: 1,
				Sequence:      1,
				Memo:          "hello from Voyager 2!",
				Msgs:          defaultMsg,
				Fee:           NewStdFee(200000, sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))}),
			},
			false,
		},
		{
			"no chain-id supplied",
			fields{
				TxEncoder:     DefaultTxEncoder(codec.New()),
				AccountNumber: 1,
				Sequence:      1,
				Gas:           200000,
				GasAdjustment: 1.1,
				SimulateGas:   false,
				ChainID:       "",
				Memo:          "hello from Voyager 1!",
				Fees:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))),
			},
			defaultMsg,
			StdSignMsg{
				ChainID:       "test-chain",
				AccountNumber: 1,
				Sequence:      1,
				Memo:          "hello from Voyager 1!",
				Msgs:          defaultMsg,
				Fee:           NewStdFee(200000, sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))}),
			},
			true,
		},
		{
			"builder w/ fees and gas prices",
			fields{
				TxEncoder:     DefaultTxEncoder(codec.New()),
				AccountNumber: 1,
				Sequence:      1,
				Gas:           200000,
				GasAdjustment: 1.1,
				SimulateGas:   false,
				ChainID:       "test-chain",
				Memo:          "hello from Voyager 1!",
				Fees:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))),
				GasPrices:     sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10000, sdk.Precision))},
			},
			defaultMsg,
			StdSignMsg{
				ChainID:       "test-chain",
				AccountNumber: 1,
				Sequence:      1,
				Memo:          "hello from Voyager 1!",
				Msgs:          defaultMsg,
				Fee:           NewStdFee(200000, sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))}),
			},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			bldr := NewTxBuilder(
				tt.fields.TxEncoder, tt.fields.AccountNumber, tt.fields.Sequence,
				tt.fields.Gas, tt.fields.GasAdjustment, tt.fields.SimulateGas,
				tt.fields.ChainID, tt.fields.Memo, tt.fields.Fees, tt.fields.GasPrices,
			)
			got, err := bldr.BuildSignMsg(tt.msgs)
			require.Equal(t, tt.wantErr, (err != nil))
			if err == nil {
				require.True(t, reflect.DeepEqual(tt.want, got))
			}
		})
	}
}

func TestTxBuilder_WithAccountNumber(t *testing.T) {
	txBuilder := initTxBuilder()
	txBuilder = txBuilder.WithAccountNumber(uint64(2))
	require.Equal(t, txBuilder.AccountNumber(), uint64(2))
}

func TestTxBuilder_ChainID(t *testing.T) {
	txBuilder := initTxBuilder()
	txBuilder = txBuilder.WithChainID("test-chain")
	require.Equal(t, txBuilder.ChainID(), "test-chain")
}

func TestTxBuilder_Fees(t *testing.T) {
	txBuilder := initTxBuilder()
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)))
	txBuilder = txBuilder.WithFees(fees.String())
	//require.DeepEqual(t, txBuilder.Fees(), fees)
	require.True(t, reflect.DeepEqual(txBuilder.Fees(), fees))
}

func TestTxBuilder_Gas(t *testing.T) {
	txBuilder := initTxBuilder()
	require.Equal(t, txBuilder.Gas(), uint64(0))
	txBuilder = txBuilder.WithGas(200000)
	require.Equal(t, txBuilder.Gas(), uint64(200000))
}

func TestTxBuilder_GasPrices(t *testing.T) {
	txBuilder := initTxBuilder()
	gasPrice := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom,
		sdk.NewDecWithPrec(10000, sdk.Precision))}
	txBuilder = txBuilder.WithGasPrices(gasPrice.String())
	require.Equal(t, txBuilder.GasPrices(), gasPrice)
}

func TestTxBuilder_Keybase(t *testing.T) {
	// Now add a temporary keybase
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	kr, err := keyring.New(t.Name(), "test", dir, nil)
	require.NoError(t, err)

	txBuilder := initTxBuilder()
	require.Equal(t, txBuilder.Keybase(), nil)
	txBuilder = txBuilder.WithKeybase(kr)
	require.Equal(t, txBuilder.Keybase(), kr)
	require.True(t, reflect.DeepEqual(txBuilder.Keybase(), kr))
}

func TestTxBuilder_Memo(t *testing.T) {
	txBuilder := initTxBuilder()
	require.Equal(t, txBuilder.Memo(), "")
	txBuilder = txBuilder.WithMemo("foo-memo")
	require.Equal(t, txBuilder.Memo(), "foo-memo")
}

func TestTxBuilder_BuildAndSign(t *testing.T) {
	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	var from = "from-key"
	txBuilder, _ := initTxBuilderWithKeybase(t, from)

	tx, err := txBuilder.BuildAndSign(from, msgs)

	require.NoError(t, err)
	require.NotNil(t, tx)
}

func TestTxBuilder_BuildSignMsg(t *testing.T) {
	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	txBuilder, _ := initTxBuilderWithKeybase(t, "from-key")

	tx, err := txBuilder.BuildSignMsg(msgs)

	require.NoError(t, err)
	require.NotNil(t, tx)
}

func TestTxBuilder_BuildTxForSim(t *testing.T) {
	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	txBuilder, _ := initTxBuilderWithKeybase(t, "from-key")
	tx, err := txBuilder.BuildTxForSim(msgs)
	require.NoError(t, err)
	require.NotNil(t, tx)
}

func TestTxBuilder_SimulateAndExecute(t *testing.T) {

}

func TestTxBuilder_SignStdTx(t *testing.T) {
	//const from = "from-key"
	//msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	//txBuilder, acc := initTxBuilderWithKeybase(t, from)
	//
	//stdTx := StdTx{
	//	Memo: txBuilder.Memo(),
	//	Fee: nil,
	//	Msgs: msgs,
	//	Signatures: []auth.StdSignature{
	//		{
	//			Signature: acc.GetPubKey().Bytes(),
	//			PubKey: acc.GetPubKey().Bytes(),
	//		},
	//	},
	//}
	//
	//signedStdTx, err := txBuilder.SignStdTx(from, stdTx, false)
	//require.NoError(t, err)
	//require.NotNil(t, signedStdTx)
	//require.Equal(t, signedStdTx.Signatures[0].PubKey, acc.GetPubKey().Bytes())
}

func TestTxBuilder_Sign(t *testing.T) {
	var from = "from-key"
	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	txBuilder, _ := initTxBuilderWithKeybase(t, from)
	stdSignMsg, err := txBuilder.BuildSignMsg(msgs)

	require.NoError(t, err)
	require.NotNil(t, stdSignMsg)

	tx, err := txBuilder.Sign(from, stdSignMsg)
	require.NoError(t, err)
	require.NotNil(t, tx)
}
