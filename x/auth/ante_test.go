package auth

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// run the tx through the anteHandler and ensure its valid
func checkValidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, simulate bool) {
	_, result, abort := anteHandler(ctx, tx, simulate)
	require.False(t, abort)
	require.Equal(t, sdk.CodeOK, result.Code)
	require.True(t, result.IsOK())
}

// run the tx through the anteHandler and ensure it fails with the given code
func checkInvalidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, simulate bool, code sdk.CodeType) {
	newCtx, result, abort := anteHandler(ctx, tx, simulate)
	require.True(t, abort)

	require.Equal(t, code, result.Code, fmt.Sprintf("Expected %v, got %v", code, result))
	require.Equal(t, sdk.CodespaceRoot, result.Codespace)

	if code == sdk.CodeOutOfGas {
		stdTx, ok := tx.(StdTx)
		require.True(t, ok, "tx must be in form auth.StdTx")
		// GasWanted set correctly
		require.Equal(t, stdTx.Fee.Gas, result.GasWanted, "Gas wanted not set correctly")
		require.True(t, result.GasUsed > result.GasWanted, "GasUsed not greated than GasWanted")
		// Check that context is set correctly
		require.Equal(t, result.GasUsed, newCtx.GasMeter().GasConsumed(), "Context not updated correctly")
	}
}

// Test various error cases in the AnteHandler control flow.
func TestAnteHandlerSigErrors(t *testing.T) {
	// setup
	input := setupTestInput()
	ctx := input.ctx
	anteHandler := NewAnteHandler(input.ak, input.fck)

	// keys and addresses
	priv1, _, addr1 := keyPubAddr()
	priv2, _, addr2 := keyPubAddr()
	priv3, _, addr3 := keyPubAddr()

	// msg and signatures
	var tx sdk.Tx
	msg1 := newTestMsg(addr1, addr2)
	msg2 := newTestMsg(addr1, addr3)
	fee := newStdFee()

	msgs := []sdk.Msg{msg1, msg2}

	// test no signatures
	privs, accNums, seqs := []crypto.PrivKey{}, []uint64{}, []uint64{}
	tx = newTestTx(ctx, msgs, privs, accNums, seqs, fee)

	// tx.GetSigners returns addresses in correct order: addr1, addr2, addr3
	expectedSigners := []sdk.AccAddress{addr1, addr2, addr3}
	stdTx := tx.(StdTx)
	require.Equal(t, expectedSigners, stdTx.GetSigners())

	// Check no signatures fails
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeNoSignatures)

	// test num sigs dont match GetSigners
	privs, accNums, seqs = []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = newTestTx(ctx, msgs, privs, accNums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// test an unrecognized account
	privs, accNums, seqs = []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	tx = newTestTx(ctx, msgs, privs, accNums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnknownAddress)

	// save the first account, but second is still unrecognized
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(fee.Amount)
	input.ak.SetAccount(ctx, acc1)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnknownAddress)
}

// Test logic around account number checking with one signer and many signers.
func TestAnteHandlerAccountNumbers(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.fck)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := keyPubAddr()
	priv2, _, addr2 := keyPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc2)

	// msg and signatures
	var tx sdk.Tx
	msg := newTestMsg(addr1)
	fee := newStdFee()

	msgs := []sdk.Msg{msg}

	// test good tx from one signer
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx from wrong account number
	seqs = []uint64{1}
	tx = newTestTx(ctx, msgs, privs, []uint64{1}, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// from correct account number
	seqs = []uint64{1}
	tx = newTestTx(ctx, msgs, privs, []uint64{0}, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx with another signer and incorrect account numbers
	msg1 := newTestMsg(addr1, addr2)
	msg2 := newTestMsg(addr2, addr1)
	msgs = []sdk.Msg{msg1, msg2}
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{1, 0}, []uint64{2, 0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// correct account numbers
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{2, 0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around account number checking with many signers when BlockHeight is 0.
func TestAnteHandlerAccountNumbersAtBlockHeightZero(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.fck)
	ctx := input.ctx.WithBlockHeight(0)

	// keys and addresses
	priv1, _, addr1 := keyPubAddr()
	priv2, _, addr2 := keyPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc2)

	// msg and signatures
	var tx sdk.Tx
	msg := newTestMsg(addr1)
	fee := newStdFee()

	msgs := []sdk.Msg{msg}

	// test good tx from one signer
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx from wrong account number
	seqs = []uint64{1}
	tx = newTestTx(ctx, msgs, privs, []uint64{1}, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// from correct account number
	seqs = []uint64{1}
	tx = newTestTx(ctx, msgs, privs, []uint64{0}, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx with another signer and incorrect account numbers
	msg1 := newTestMsg(addr1, addr2)
	msg2 := newTestMsg(addr2, addr1)
	msgs = []sdk.Msg{msg1, msg2}
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{1, 0}, []uint64{2, 0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// correct account numbers
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 0}, []uint64{2, 0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around sequence checking with one signer and many signers.
func TestAnteHandlerSequences(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.fck)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := keyPubAddr()
	priv2, _, addr2 := keyPubAddr()
	priv3, _, addr3 := keyPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc2)
	acc3 := input.ak.NewAccountWithAddress(ctx, addr3)
	acc3.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc3)

	// msg and signatures
	var tx sdk.Tx
	msg := newTestMsg(addr1)
	fee := newStdFee()

	msgs := []sdk.Msg{msg}

	// test good tx from one signer
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// test sending it again fails (replay protection)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// fix sequence, should pass
	seqs = []uint64{1}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx with another signer and correct sequences
	msg1 := newTestMsg(addr1, addr2)
	msg2 := newTestMsg(addr3, addr1)
	msgs = []sdk.Msg{msg1, msg2}

	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{2, 0, 0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// replay fails
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// tx from just second signer with incorrect sequence fails
	msg = newTestMsg(addr2)
	msgs = []sdk.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{1}, []uint64{0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// fix the sequence and it passes
	tx = newTestTx(ctx, msgs, []crypto.PrivKey{priv2}, []uint64{1}, []uint64{1}, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// another tx from both of them that passes
	msg = newTestMsg(addr1, addr2)
	msgs = []sdk.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{3, 2}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around fee deduction.
func TestAnteHandlerFees(t *testing.T) {
	// setup
	input := setupTestInput()
	ctx := input.ctx
	anteHandler := NewAnteHandler(input.ak, input.fck)

	// keys and addresses
	priv1, _, addr1 := keyPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	input.ak.SetAccount(ctx, acc1)

	// msg and signatures
	var tx sdk.Tx
	msg := newTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := newStdFee()
	msgs := []sdk.Msg{msg}

	// signer does not have enough funds to pay the fee
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInsufficientFunds)

	acc1.SetCoins(sdk.Coins{sdk.NewInt64Coin("atom", 149)})
	input.ak.SetAccount(ctx, acc1)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInsufficientFunds)

	require.True(t, input.fck.GetCollectedFees(ctx).IsEqual(emptyCoins))
	require.True(t, input.ak.GetAccount(ctx, addr1).GetCoins().AmountOf("atom").Equal(sdk.NewInt(149)))

	acc1.SetCoins(sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	input.ak.SetAccount(ctx, acc1)
	checkValidTx(t, anteHandler, ctx, tx, false)

	require.True(t, input.fck.GetCollectedFees(ctx).IsEqual(sdk.Coins{sdk.NewInt64Coin("atom", 150)}))
	require.True(t, input.ak.GetAccount(ctx, addr1).GetCoins().AmountOf("atom").Equal(sdk.NewInt(0)))
}

// Test logic around memo gas consumption.
func TestAnteHandlerMemoGas(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.fck)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := keyPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	input.ak.SetAccount(ctx, acc1)

	// msg and signatures
	var tx sdk.Tx
	msg := newTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := NewStdFee(0, sdk.Coins{sdk.NewInt64Coin("atom", 0)})

	// tx does not have enough gas
	tx = newTestTx(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeOutOfGas)

	// tx with memo doesn't have enough gas
	fee = NewStdFee(801, sdk.Coins{sdk.NewInt64Coin("atom", 0)})
	tx = newTestTxWithMemo(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee, "abcininasidniandsinasindiansdiansdinaisndiasndiadninsd")
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeOutOfGas)

	// memo too large
	fee = NewStdFee(9000, sdk.Coins{sdk.NewInt64Coin("atom", 0)})
	tx = newTestTxWithMemo(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee, strings.Repeat("01234567890", 500))
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeMemoTooLarge)

	// tx with memo has enough gas
	fee = NewStdFee(9000, sdk.Coins{sdk.NewInt64Coin("atom", 0)})
	tx = newTestTxWithMemo(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee, strings.Repeat("0123456789", 10))
	checkValidTx(t, anteHandler, ctx, tx, false)
}

func TestAnteHandlerMultiSigner(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.fck)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := keyPubAddr()
	priv2, _, addr2 := keyPubAddr()
	priv3, _, addr3 := keyPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc2)
	acc3 := input.ak.NewAccountWithAddress(ctx, addr3)
	acc3.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc3)

	// set up msgs and fee
	var tx sdk.Tx
	msg1 := newTestMsg(addr1, addr2)
	msg2 := newTestMsg(addr3, addr1)
	msg3 := newTestMsg(addr2, addr3)
	msgs := []sdk.Msg{msg1, msg2, msg3}
	fee := newStdFee()

	// signers in order
	privs, accnums, seqs := []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	tx = newTestTxWithMemo(ctx, msgs, privs, accnums, seqs, fee, "Check signers are in expected order and different account numbers works")

	checkValidTx(t, anteHandler, ctx, tx, false)

	// change sequence numbers
	tx = newTestTx(ctx, []sdk.Msg{msg1}, []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{1, 1}, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
	tx = newTestTx(ctx, []sdk.Msg{msg2}, []crypto.PrivKey{priv3, priv1}, []uint64{2, 0}, []uint64{1, 2}, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// expected seqs = [3, 2, 2]
	tx = newTestTxWithMemo(ctx, msgs, privs, accnums, []uint64{3, 2, 2}, fee, "Check signers are in expected order and different account numbers and sequence numbers works")
	checkValidTx(t, anteHandler, ctx, tx, false)
}

func TestAnteHandlerBadSignBytes(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.fck)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := keyPubAddr()
	priv2, _, addr2 := keyPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc2)

	var tx sdk.Tx
	msg := newTestMsg(addr1)
	msgs := []sdk.Msg{msg}
	fee := newStdFee()
	fee2 := newStdFee()
	fee2.Gas += 100
	fee3 := newStdFee()
	fee3.Amount[0].Amount = fee3.Amount[0].Amount.AddRaw(100)

	// test good tx and signBytes
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	chainID := ctx.ChainID()
	chainID2 := chainID + "somemorestuff"
	codeUnauth := sdk.CodeUnauthorized

	cases := []struct {
		chainID string
		accnum  uint64
		seq     uint64
		fee     StdFee
		msgs    []sdk.Msg
		code    sdk.CodeType
	}{
		{chainID2, 0, 1, fee, msgs, codeUnauth},                        // test wrong chain_id
		{chainID, 0, 2, fee, msgs, codeUnauth},                         // test wrong seqs
		{chainID, 1, 1, fee, msgs, codeUnauth},                         // test wrong accnum
		{chainID, 0, 1, fee, []sdk.Msg{newTestMsg(addr2)}, codeUnauth}, // test wrong msg
		{chainID, 0, 1, fee2, msgs, codeUnauth},                        // test wrong fee
		{chainID, 0, 1, fee3, msgs, codeUnauth},                        // test wrong fee
	}

	privs, seqs = []crypto.PrivKey{priv1}, []uint64{1}
	for _, cs := range cases {
		tx := newTestTxWithSignBytes(
			msgs, privs, accnums, seqs, fee,
			StdSignBytes(cs.chainID, cs.accnum, cs.seq, cs.fee, cs.msgs, ""),
			"",
		)
		checkInvalidTx(t, anteHandler, ctx, tx, false, cs.code)
	}

	// test wrong signer if public key exist
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{0}, []uint64{1}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// test wrong signer if public doesn't exist
	msg = newTestMsg(addr2)
	msgs = []sdk.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv1}, []uint64{1}, []uint64{0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInvalidPubKey)
}

func TestAnteHandlerSetPubKey(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.fck)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := keyPubAddr()
	_, _, addr2 := keyPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc2)

	var tx sdk.Tx

	// test good tx and set public key
	msg := newTestMsg(addr1)
	msgs := []sdk.Msg{msg}
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := newStdFee()
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	acc1 = input.ak.GetAccount(ctx, addr1)
	require.Equal(t, acc1.GetPubKey(), priv1.PubKey())

	// test public key not found
	msg = newTestMsg(addr2)
	msgs = []sdk.Msg{msg}
	tx = newTestTx(ctx, msgs, privs, []uint64{1}, seqs, fee)
	sigs := tx.(StdTx).GetSignatures()
	sigs[0].PubKey = nil
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInvalidPubKey)

	acc2 = input.ak.GetAccount(ctx, addr2)
	require.Nil(t, acc2.GetPubKey())

	// test invalid signature and public key
	tx = newTestTx(ctx, msgs, privs, []uint64{1}, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInvalidPubKey)

	acc2 = input.ak.GetAccount(ctx, addr2)
	require.Nil(t, acc2.GetPubKey())
}

func TestProcessPubKey(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	// keys
	_, _, addr1 := keyPubAddr()
	priv2, _, addr2 := keyPubAddr()
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)

	acc2.SetPubKey(priv2.PubKey())

	type args struct {
		acc      Account
		sig      StdSignature
		simulate bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"no sigs, simulate off", args{acc1, StdSignature{}, false}, true},
		{"no sigs, simulate on", args{acc1, StdSignature{}, true}, false},
		{"no sigs, account with pub, simulate on", args{acc2, StdSignature{}, true}, false},
		{"pubkey doesn't match addr, simulate off", args{acc1, StdSignature{PubKey: priv2.PubKey()}, false}, true},
		{"pubkey doesn't match addr, simulate on", args{acc1, StdSignature{PubKey: priv2.PubKey()}, true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ProcessPubKey(tt.args.acc, tt.args.sig, tt.args.simulate)
			require.Equal(t, tt.wantErr, !err.IsOK())
		})
	}
}

func TestConsumeSignatureVerificationGas(t *testing.T) {
	params := DefaultParams()

	type args struct {
		meter  sdk.GasMeter
		pubkey crypto.PubKey
		params Params
	}
	tests := []struct {
		name        string
		args        args
		gasConsumed uint64
		wantPanic   bool
	}{
		{"PubKeyEd25519", args{sdk.NewInfiniteGasMeter(), ed25519.GenPrivKey().PubKey(), params}, DefaultSigVerifyCostED25519, false},
		{"PubKeySecp256k1", args{sdk.NewInfiniteGasMeter(), secp256k1.GenPrivKey().PubKey(), params}, DefaultSigVerifyCostSecp256k1, false},
		{"unknown key", args{sdk.NewInfiniteGasMeter(), nil, params}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() { consumeSignatureVerificationGas(tt.args.meter, tt.args.pubkey, tt.args.params) })
			} else {
				consumeSignatureVerificationGas(tt.args.meter, tt.args.pubkey, tt.args.params)
				require.Equal(t, tt.args.meter.GasConsumed(), tt.gasConsumed)
			}
		})
	}
}

func TestAdjustFeesByGas(t *testing.T) {
	type args struct {
		fee sdk.Coins
		gas uint64
	}
	tests := []struct {
		name string
		args args
		want sdk.Coins
	}{
		{"nil coins", args{sdk.Coins{}, 100000}, sdk.Coins{}},
		{"nil coins", args{sdk.Coins{sdk.NewInt64Coin("a", 10), sdk.NewInt64Coin("b", 0)}, 100000}, sdk.Coins{sdk.NewInt64Coin("a", 20), sdk.NewInt64Coin("b", 10)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.True(t, tt.want.IsEqual(adjustFeesByGas(tt.args.fee, tt.args.gas)))
		})
	}
}

func TestCountSubkeys(t *testing.T) {
	genPubKeys := func(n int) []crypto.PubKey {
		var ret []crypto.PubKey
		for i := 0; i < n; i++ {
			ret = append(ret, secp256k1.GenPrivKey().PubKey())
		}
		return ret
	}
	genMultiKey := func(n, k int, keysGen func(n int) []crypto.PubKey) crypto.PubKey {
		return multisig.NewPubKeyMultisigThreshold(k, keysGen(n))
	}
	type args struct {
		pub crypto.PubKey
	}
	mkey := genMultiKey(5, 4, genPubKeys)
	mkeyType := mkey.(*multisig.PubKeyMultisigThreshold)
	mkeyType.PubKeys = append(mkeyType.PubKeys, genMultiKey(6, 5, genPubKeys))
	tests := []struct {
		name string
		args args
		want int
	}{
		{"single key", args{secp256k1.GenPrivKey().PubKey()}, 1},
		{"multi sig key", args{genMultiKey(5, 4, genPubKeys)}, 5},
		{"multi multi sig", args{mkey}, 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(T *testing.T) {
			require.Equal(t, tt.want, countSubKeys(tt.args.pub))
		})
	}
}

func TestAnteHandlerSigLimitExceeded(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.fck)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := keyPubAddr()
	priv2, _, addr2 := keyPubAddr()
	priv3, _, addr3 := keyPubAddr()
	priv4, _, addr4 := keyPubAddr()
	priv5, _, addr5 := keyPubAddr()
	priv6, _, addr6 := keyPubAddr()
	priv7, _, addr7 := keyPubAddr()
	priv8, _, addr8 := keyPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(newCoins())
	input.ak.SetAccount(ctx, acc2)

	var tx sdk.Tx
	msg := newTestMsg(addr1, addr2, addr3, addr4, addr5, addr6, addr7, addr8)
	msgs := []sdk.Msg{msg}
	fee := newStdFee()

	// test rejection logic
	privs, accnums, seqs := []crypto.PrivKey{priv1, priv2, priv3, priv4, priv5, priv6, priv7, priv8},
		[]uint64{0, 0, 0, 0, 0, 0, 0, 0}, []uint64{0, 0, 0, 0, 0, 0, 0, 0}
	tx = newTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeTooManySignatures)
}
