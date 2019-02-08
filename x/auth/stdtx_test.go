package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	priv = ed25519.GenPrivKey()
	addr = sdk.AccAddress(priv.PubKey().Address())
)

func TestStdTx(t *testing.T) {
	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	fee := newStdFee()
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
	defaultFee := newStdFee()
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
	priv1, _, addr1 := keyPubAddr()
	priv2, _, addr2 := keyPubAddr()
	priv3, _, addr3 := keyPubAddr()
	priv4, _, addr4 := keyPubAddr()
	priv5, _, addr5 := keyPubAddr()
	priv6, _, addr6 := keyPubAddr()
	priv7, _, addr7 := keyPubAddr()
	priv8, _, addr8 := keyPubAddr()

	// msg and signatures
	msg1 := newTestMsg(addr1, addr2)
	fee := newStdFee()

	msgs := []sdk.Msg{msg1}

	// require to fail validation upon invalid fee
	badFee := newStdFee()
	badFee.Amount[0].Amount = sdk.NewInt(-5)
	tx := newTestTx(ctx, nil, nil, nil, nil, badFee)

	err := tx.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, sdk.CodeInsufficientFee, err.Result().Code)

	// require to fail validation when no signatures exist
	privs, accNums, seqs := []crypto.PrivKey{}, []uint64{}, []uint64{}
	tx = newTestTx(ctx, msgs, privs, accNums, seqs, fee)

	err = tx.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, sdk.CodeNoSignatures, err.Result().Code)

	// require to fail validation when signatures do not match expected signers
	privs, accNums, seqs = []crypto.PrivKey{priv1}, []uint64{0, 1}, []uint64{0, 0}
	tx = newTestTx(ctx, msgs, privs, accNums, seqs, fee)

	err = tx.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, sdk.CodeUnauthorized, err.Result().Code)

	// require to fail validation when there are too many signatures
	privs = []crypto.PrivKey{priv1, priv2, priv3, priv4, priv5, priv6, priv7, priv8}
	accNums, seqs = []uint64{0, 0, 0, 0, 0, 0, 0, 0}, []uint64{0, 0, 0, 0, 0, 0, 0, 0}
	badMsg := newTestMsg(addr1, addr2, addr3, addr4, addr5, addr6, addr7, addr8)
	badMsgs := []sdk.Msg{badMsg}
	tx = newTestTx(ctx, badMsgs, privs, accNums, seqs, fee)

	err = tx.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, sdk.CodeTooManySignatures, err.Result().Code)

	// require to fail with invalid gas supplied
	badFee = newStdFee()
	badFee.Gas = 9223372036854775808
	tx = newTestTx(ctx, nil, nil, nil, nil, badFee)

	err = tx.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, sdk.CodeGasOverflow, err.Result().Code)

	// require to pass when above criteria are matched
	privs, accNums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{0, 0}
	tx = newTestTx(ctx, msgs, privs, accNums, seqs, fee)

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
	fee := newStdFee()
	sigs := []StdSignature{}

	tx := NewStdTx(msgs, fee, sigs, "")

	cdcBytes, err := cdc.MarshalBinaryLengthPrefixed(tx)

	require.NoError(t, err)
	encoderBytes, err := encoder(tx)

	require.NoError(t, err)
	require.Equal(t, cdcBytes, encoderBytes)
}
