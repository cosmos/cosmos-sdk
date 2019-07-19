package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	priv = ed25519.GenPrivKey()
	addr = sdk.AccAddress(priv.PubKey().Address())
)

func TestStdTx(t *testing.T) {
	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	fee := NewTestStdFee()
	sigs := []StdSignature{}

	tx := NewStdTx(msgs, fee, sigs, "")
	require.Equal(t, msgs, tx.GetMsgs())
	require.Equal(t, sigs, tx.GetSignatures())

	feePayer := tx.GetSigners()[0]
	require.Equal(t, addr, feePayer)
}

func TestStdSignBytes(t *testing.T) {
	type args struct {
		chainID  string
		accnum   uint64
		sequence uint64
		fee      StdFee
		msgs     []sdk.Msg
		memo     string
	}
	defaultFee := NewTestStdFee()
	tests := []struct {
		args args
		want string
	}{
		{
			args{"1234", 3, 6, defaultFee, []sdk.Msg{sdk.NewTestMsg(addr)}, "memo"},
			fmt.Sprintf("{\"account_number\":\"3\",\"chain_id\":\"1234\",\"fee\":{\"amount\":[{\"amount\":\"150\",\"denom\":\"atom\"}],\"gas\":\"50000\"},\"memo\":\"memo\",\"msgs\":[[\"%s\"]],\"sequence\":\"6\"}", addr),
		},
	}
	for i, tc := range tests {
		got := string(StdSignBytes(tc.args.chainID, tc.args.accnum, tc.args.sequence, tc.args.fee, tc.args.msgs, tc.args.memo))
		require.Equal(t, tc.want, got, "Got unexpected result on test case i: %d", i)
	}
}

func TestTxValidateBasic(t *testing.T) {
	ctx := sdk.NewContext(nil, abci.Header{ChainID: "mychainid"}, false, log.NewNopLogger())

	// keys and addresses
	priv1, _, addr1 := KeyTestPubAddr()
	priv2, _, addr2 := KeyTestPubAddr()

	// msg and signatures
	msg1 := NewTestMsg(addr1, addr2)
	fee := NewTestStdFee()

	msgs := []sdk.Msg{msg1}

	// require to fail validation upon invalid fee
	badFee := NewTestStdFee()
	badFee.Amount[0].Amount = sdk.NewInt(-5)
	tx := NewTestTx(ctx, nil, nil, nil, nil, badFee)

	err := tx.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, sdk.CodeInsufficientFee, err.Result().Code)

	// require to fail validation when no signatures exist
	privs, accNums, seqs := []crypto.PrivKey{}, []uint64{}, []uint64{}
	tx = NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	err = tx.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, sdk.CodeNoSignatures, err.Result().Code)

	// require to fail validation when signatures do not match expected signers
	privs, accNums, seqs = []crypto.PrivKey{priv1}, []uint64{0, 1}, []uint64{0, 0}
	tx = NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	err = tx.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, sdk.CodeUnauthorized, err.Result().Code)

	// require to fail with invalid gas supplied
	badFee = NewTestStdFee()
	badFee.Gas = 9223372036854775808
	tx = NewTestTx(ctx, nil, nil, nil, nil, badFee)

	err = tx.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, sdk.CodeGasOverflow, err.Result().Code)

	// require to pass when above criteria are matched
	privs, accNums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{0, 0}
	tx = NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	err = tx.ValidateBasic()
	require.NoError(t, err)
}

func TestDefaultTxEncoder(t *testing.T) {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	RegisterCodec(cdc)
	cdc.RegisterConcrete(sdk.TestMsg{}, "cosmos-sdk/Test", nil)
	encoder := DefaultTxEncoder(cdc)

	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	fee := NewTestStdFee()
	sigs := []StdSignature{}

	tx := NewStdTx(msgs, fee, sigs, "")

	cdcBytes, err := cdc.MarshalBinaryLengthPrefixed(tx)

	require.NoError(t, err)
	encoderBytes, err := encoder(tx)

	require.NoError(t, err)
	require.Equal(t, cdcBytes, encoderBytes)
}

func TestStdSignatureMarshalYAML(t *testing.T) {
	_, pubKey, _ := KeyTestPubAddr()

	testCases := []struct {
		sig    StdSignature
		output string
	}{
		{
			StdSignature{},
			"|\n  pubkey: \"\"\n  signature: \"\"\n",
		},
		{
			StdSignature{PubKey: pubKey, Signature: []byte("dummySig")},
			fmt.Sprintf("|\n  pubkey: %s\n  signature: dummySig\n", sdk.MustBech32ifyAccPub(pubKey)),
		},
		{
			StdSignature{PubKey: pubKey, Signature: nil},
			fmt.Sprintf("|\n  pubkey: %s\n  signature: \"\"\n", sdk.MustBech32ifyAccPub(pubKey)),
		},
	}

	for i, tc := range testCases {
		bz, err := yaml.Marshal(tc.sig)
		require.NoError(t, err)
		require.Equal(t, tc.output, string(bz), "test case #%d", i)
	}
}
