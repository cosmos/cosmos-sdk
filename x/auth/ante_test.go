package auth

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// run the tx through the anteHandler and ensure its valid
func checkValidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, simulate bool) {
	_, result, abort := anteHandler(ctx, tx, simulate)
	require.Equal(t, "", result.Log)
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
		stdTx, ok := tx.(types.StdTx)
		require.True(t, ok, "tx must be in form auth.types.StdTx")
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
	anteHandler := NewAnteHandler(input.ak, input.sk, DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()

	// msg and signatures
	var tx sdk.Tx
	msg1 := types.NewTestMsg(addr1, addr2)
	msg2 := types.NewTestMsg(addr1, addr3)
	fee := types.NewTestStdFee()

	msgs := []sdk.Msg{msg1, msg2}

	// test no signatures
	privs, accNums, seqs := []crypto.PrivKey{}, []uint64{}, []uint64{}
	tx = types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	// tx.GetSigners returns addresses in correct order: addr1, addr2, addr3
	expectedSigners := []sdk.AccAddress{addr1, addr2, addr3}
	stdTx := tx.(types.StdTx)
	require.Equal(t, expectedSigners, stdTx.GetSigners())

	// Check no signatures fails
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeNoSignatures)

	// test num sigs dont match GetSigners
	privs, accNums, seqs = []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// test an unrecognized account
	privs, accNums, seqs = []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)
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
	anteHandler := NewAnteHandler(input.ak, input.sk, DefaultSigVerificationGasConsumer)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	input.ak.SetAccount(ctx, acc2)

	// msg and signatures
	var tx sdk.Tx
	msg := types.NewTestMsg(addr1)
	fee := types.NewTestStdFee()

	msgs := []sdk.Msg{msg}

	// test good tx from one signer
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx from wrong account number
	seqs = []uint64{1}
	tx = types.NewTestTx(ctx, msgs, privs, []uint64{1}, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// from correct account number
	seqs = []uint64{1}
	tx = types.NewTestTx(ctx, msgs, privs, []uint64{0}, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx with another signer and incorrect account numbers
	msg1 := types.NewTestMsg(addr1, addr2)
	msg2 := types.NewTestMsg(addr2, addr1)
	msgs = []sdk.Msg{msg1, msg2}
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{1, 0}, []uint64{2, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// correct account numbers
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{2, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around account number checking with many signers when BlockHeight is 0.
func TestAnteHandlerAccountNumbersAtBlockHeightZero(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.sk, DefaultSigVerificationGasConsumer)
	ctx := input.ctx.WithBlockHeight(0)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()

	// set the accounts, we don't need the acc numbers as it is in the genesis block
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	input.ak.SetAccount(ctx, acc2)

	// msg and signatures
	var tx sdk.Tx
	msg := types.NewTestMsg(addr1)
	fee := types.NewTestStdFee()

	msgs := []sdk.Msg{msg}

	// test good tx from one signer
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx from wrong account number
	seqs = []uint64{1}
	tx = types.NewTestTx(ctx, msgs, privs, []uint64{1}, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// from correct account number
	seqs = []uint64{1}
	tx = types.NewTestTx(ctx, msgs, privs, []uint64{0}, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx with another signer and incorrect account numbers
	msg1 := types.NewTestMsg(addr1, addr2)
	msg2 := types.NewTestMsg(addr2, addr1)
	msgs = []sdk.Msg{msg1, msg2}
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{1, 0}, []uint64{2, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// correct account numbers
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 0}, []uint64{2, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around sequence checking with one signer and many signers.
func TestAnteHandlerSequences(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.sk, DefaultSigVerificationGasConsumer)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	input.ak.SetAccount(ctx, acc2)
	acc3 := input.ak.NewAccountWithAddress(ctx, addr3)
	acc3.SetCoins(types.NewTestCoins())
	require.NoError(t, acc3.SetAccountNumber(2))
	input.ak.SetAccount(ctx, acc3)

	// msg and signatures
	var tx sdk.Tx
	msg := types.NewTestMsg(addr1)
	fee := types.NewTestStdFee()

	msgs := []sdk.Msg{msg}

	// test good tx from one signer
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// test sending it again fails (replay protection)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// fix sequence, should pass
	seqs = []uint64{1}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx with another signer and correct sequences
	msg1 := types.NewTestMsg(addr1, addr2)
	msg2 := types.NewTestMsg(addr3, addr1)
	msgs = []sdk.Msg{msg1, msg2}

	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{2, 0, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// replay fails
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// tx from just second signer with incorrect sequence fails
	msg = types.NewTestMsg(addr2)
	msgs = []sdk.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{1}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// fix the sequence and it passes
	tx = types.NewTestTx(ctx, msgs, []crypto.PrivKey{priv2}, []uint64{1}, []uint64{1}, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// another tx from both of them that passes
	msg = types.NewTestMsg(addr1, addr2)
	msgs = []sdk.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{3, 2}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around fee deduction.
func TestAnteHandlerFees(t *testing.T) {
	// setup
	input := setupTestInput()
	ctx := input.ctx
	anteHandler := NewAnteHandler(input.ak, input.sk, DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	input.ak.SetAccount(ctx, acc1)

	// msg and signatures
	var tx sdk.Tx
	msg := types.NewTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := types.NewTestStdFee()
	msgs := []sdk.Msg{msg}

	// signer does not have enough funds to pay the fee
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInsufficientFunds)

	acc1.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("atom", 149)))
	input.ak.SetAccount(ctx, acc1)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInsufficientFunds)

	require.True(t, input.sk.GetModuleAccount(ctx, types.FeeCollectorName).GetCoins().Empty())
	require.True(sdk.IntEq(t, input.ak.GetAccount(ctx, addr1).GetCoins().AmountOf("atom"), sdk.NewInt(149)))

	acc1.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("atom", 150)))
	input.ak.SetAccount(ctx, acc1)
	checkValidTx(t, anteHandler, ctx, tx, false)

	require.True(sdk.IntEq(t, input.sk.GetModuleAccount(ctx, types.FeeCollectorName).GetCoins().AmountOf("atom"), sdk.NewInt(150)))
	require.True(sdk.IntEq(t, input.ak.GetAccount(ctx, addr1).GetCoins().AmountOf("atom"), sdk.NewInt(0)))
}

// Test logic around memo gas consumption.
func TestAnteHandlerMemoGas(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.sk, DefaultSigVerificationGasConsumer)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	require.NoError(t, acc1.SetAccountNumber(0))
	input.ak.SetAccount(ctx, acc1)

	// msg and signatures
	var tx sdk.Tx
	msg := types.NewTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := NewStdFee(0, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))

	// tx does not have enough gas
	tx = types.NewTestTx(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeOutOfGas)

	// tx with memo doesn't have enough gas
	fee = NewStdFee(801, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))
	tx = types.NewTestTxWithMemo(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee, "abcininasidniandsinasindiansdiansdinaisndiasndiadninsd")
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeOutOfGas)

	// memo too large
	fee = NewStdFee(9000, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))
	tx = types.NewTestTxWithMemo(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee, strings.Repeat("01234567890", 500))
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeMemoTooLarge)

	// tx with memo has enough gas
	fee = NewStdFee(9000, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))
	tx = types.NewTestTxWithMemo(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee, strings.Repeat("0123456789", 10))
	checkValidTx(t, anteHandler, ctx, tx, false)
}

func TestAnteHandlerMultiSigner(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.sk, DefaultSigVerificationGasConsumer)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	input.ak.SetAccount(ctx, acc2)
	acc3 := input.ak.NewAccountWithAddress(ctx, addr3)
	acc3.SetCoins(types.NewTestCoins())
	require.NoError(t, acc3.SetAccountNumber(2))
	input.ak.SetAccount(ctx, acc3)

	// set up msgs and fee
	var tx sdk.Tx
	msg1 := types.NewTestMsg(addr1, addr2)
	msg2 := types.NewTestMsg(addr3, addr1)
	msg3 := types.NewTestMsg(addr2, addr3)
	msgs := []sdk.Msg{msg1, msg2, msg3}
	fee := types.NewTestStdFee()

	// signers in order
	privs, accnums, seqs := []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	tx = types.NewTestTxWithMemo(ctx, msgs, privs, accnums, seqs, fee, "Check signers are in expected order and different account numbers works")

	checkValidTx(t, anteHandler, ctx, tx, false)

	// change sequence numbers
	tx = types.NewTestTx(ctx, []sdk.Msg{msg1}, []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{1, 1}, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
	tx = types.NewTestTx(ctx, []sdk.Msg{msg2}, []crypto.PrivKey{priv3, priv1}, []uint64{2, 0}, []uint64{1, 2}, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// expected seqs = [3, 2, 2]
	tx = types.NewTestTxWithMemo(ctx, msgs, privs, accnums, []uint64{3, 2, 2}, fee, "Check signers are in expected order and different account numbers and sequence numbers works")
	checkValidTx(t, anteHandler, ctx, tx, false)
}

func TestAnteHandlerBadSignBytes(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.sk, DefaultSigVerificationGasConsumer)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	input.ak.SetAccount(ctx, acc2)

	var tx sdk.Tx
	msg := types.NewTestMsg(addr1)
	msgs := []sdk.Msg{msg}
	fee := types.NewTestStdFee()
	fee2 := types.NewTestStdFee()
	fee2.Gas += 100
	fee3 := types.NewTestStdFee()
	fee3.Amount[0].Amount = fee3.Amount[0].Amount.AddRaw(100)

	// test good tx and signBytes
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
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
		{chainID2, 0, 1, fee, msgs, codeUnauth},                              // test wrong chain_id
		{chainID, 0, 2, fee, msgs, codeUnauth},                               // test wrong seqs
		{chainID, 1, 1, fee, msgs, codeUnauth},                               // test wrong accnum
		{chainID, 0, 1, fee, []sdk.Msg{types.NewTestMsg(addr2)}, codeUnauth}, // test wrong msg
		{chainID, 0, 1, fee2, msgs, codeUnauth},                              // test wrong fee
		{chainID, 0, 1, fee3, msgs, codeUnauth},                              // test wrong fee
	}

	privs, seqs = []crypto.PrivKey{priv1}, []uint64{1}
	for _, cs := range cases {
		tx := types.NewTestTxWithSignBytes(
			msgs, privs, accnums, seqs, fee,
			StdSignBytes(cs.chainID, cs.accnum, cs.seq, cs.fee, cs.msgs, ""),
			"",
		)
		checkInvalidTx(t, anteHandler, ctx, tx, false, cs.code)
	}

	// test wrong signer if public key exist
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{0}, []uint64{1}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeUnauthorized)

	// test wrong signer if public doesn't exist
	msg = types.NewTestMsg(addr2)
	msgs = []sdk.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv1}, []uint64{1}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInvalidPubKey)
}

func TestAnteHandlerSetPubKey(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.sk, DefaultSigVerificationGasConsumer)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	_, _, addr2 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	input.ak.SetAccount(ctx, acc2)

	var tx sdk.Tx

	// test good tx and set public key
	msg := types.NewTestMsg(addr1)
	msgs := []sdk.Msg{msg}
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := types.NewTestStdFee()
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	acc1 = input.ak.GetAccount(ctx, addr1)
	require.Equal(t, acc1.GetPubKey(), priv1.PubKey())

	// test public key not found
	msg = types.NewTestMsg(addr2)
	msgs = []sdk.Msg{msg}
	tx = types.NewTestTx(ctx, msgs, privs, []uint64{1}, seqs, fee)
	sigs := tx.(types.StdTx).GetSignatures()
	sigs[0].PubKey = nil
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInvalidPubKey)

	acc2 = input.ak.GetAccount(ctx, addr2)
	require.Nil(t, acc2.GetPubKey())

	// test invalid signature and public key
	tx = types.NewTestTx(ctx, msgs, privs, []uint64{1}, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInvalidPubKey)

	acc2 = input.ak.GetAccount(ctx, addr2)
	require.Nil(t, acc2.GetPubKey())
}

func TestProcessPubKey(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	// keys
	_, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
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
	msg := []byte{1, 2, 3, 4}

	pkSet1, sigSet1 := generatePubKeysAndSignatures(5, msg, false)
	multisigKey1 := multisig.NewPubKeyMultisigThreshold(2, pkSet1)
	multisignature1 := multisig.NewMultisig(len(pkSet1))
	expectedCost1 := expectedGasCostByKeys(pkSet1)
	for i := 0; i < len(pkSet1); i++ {
		multisignature1.AddSignatureFromPubKey(sigSet1[i], pkSet1[i], pkSet1)
	}

	type args struct {
		meter  sdk.GasMeter
		sig    []byte
		pubkey crypto.PubKey
		params Params
	}
	tests := []struct {
		name        string
		args        args
		gasConsumed uint64
		shouldErr   bool
	}{
		{"PubKeyEd25519", args{sdk.NewInfiniteGasMeter(), nil, ed25519.GenPrivKey().PubKey(), params}, DefaultSigVerifyCostED25519, true},
		{"PubKeySecp256k1", args{sdk.NewInfiniteGasMeter(), nil, secp256k1.GenPrivKey().PubKey(), params}, DefaultSigVerifyCostSecp256k1, false},
		{"Multisig", args{sdk.NewInfiniteGasMeter(), multisignature1.Marshal(), multisigKey1, params}, expectedCost1, false},
		{"unknown key", args{sdk.NewInfiniteGasMeter(), nil, nil, params}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := DefaultSigVerificationGasConsumer(tt.args.meter, tt.args.sig, tt.args.pubkey, tt.args.params)

			if tt.shouldErr {
				require.False(t, res.IsOK())
			} else {
				require.True(t, res.IsOK())
				require.Equal(t, tt.gasConsumed, tt.args.meter.GasConsumed(), fmt.Sprintf("%d != %d", tt.gasConsumed, tt.args.meter.GasConsumed()))
			}
		})
	}
}

func generatePubKeysAndSignatures(n int, msg []byte, keyTypeed25519 bool) (pubkeys []crypto.PubKey, signatures [][]byte) {
	pubkeys = make([]crypto.PubKey, n)
	signatures = make([][]byte, n)
	for i := 0; i < n; i++ {
		var privkey crypto.PrivKey
		if rand.Int63()%2 == 0 {
			privkey = ed25519.GenPrivKey()
		} else {
			privkey = secp256k1.GenPrivKey()
		}
		pubkeys[i] = privkey.PubKey()
		signatures[i], _ = privkey.Sign(msg)
	}
	return
}

func expectedGasCostByKeys(pubkeys []crypto.PubKey) uint64 {
	cost := uint64(0)
	for _, pubkey := range pubkeys {
		pubkeyType := strings.ToLower(fmt.Sprintf("%T", pubkey))
		switch {
		case strings.Contains(pubkeyType, "ed25519"):
			cost += DefaultParams().SigVerifyCostED25519
		case strings.Contains(pubkeyType, "secp256k1"):
			cost += DefaultParams().SigVerifyCostSecp256k1
		default:
			panic("unexpected key type")
		}
	}
	return cost
}

func TestCountSubkeys(t *testing.T) {
	genPubKeys := func(n int) []crypto.PubKey {
		var ret []crypto.PubKey
		for i := 0; i < n; i++ {
			ret = append(ret, secp256k1.GenPrivKey().PubKey())
		}
		return ret
	}
	singleKey := secp256k1.GenPrivKey().PubKey()
	singleLevelMultiKey := multisig.NewPubKeyMultisigThreshold(4, genPubKeys(5))
	multiLevelSubKey1 := multisig.NewPubKeyMultisigThreshold(4, genPubKeys(5))
	multiLevelSubKey2 := multisig.NewPubKeyMultisigThreshold(4, genPubKeys(5))
	multiLevelMultiKey := multisig.NewPubKeyMultisigThreshold(2, []crypto.PubKey{
		multiLevelSubKey1, multiLevelSubKey2, secp256k1.GenPrivKey().PubKey()})
	type args struct {
		pub crypto.PubKey
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"single key", args{singleKey}, 1},
		{"single level multikey", args{singleLevelMultiKey}, 5},
		{"multi level multikey", args{multiLevelMultiKey}, 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(T *testing.T) {
			require.Equal(t, tt.want, CountSubKeys(tt.args.pub))
		})
	}
}

func TestAnteHandlerSigLimitExceeded(t *testing.T) {
	// setup
	input := setupTestInput()
	anteHandler := NewAnteHandler(input.ak, input.sk, DefaultSigVerificationGasConsumer)
	ctx := input.ctx.WithBlockHeight(1)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()
	priv4, _, addr4 := types.KeyTestPubAddr()
	priv5, _, addr5 := types.KeyTestPubAddr()
	priv6, _, addr6 := types.KeyTestPubAddr()
	priv7, _, addr7 := types.KeyTestPubAddr()
	priv8, _, addr8 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	input.ak.SetAccount(ctx, acc1)
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	input.ak.SetAccount(ctx, acc2)

	var tx sdk.Tx
	msg := types.NewTestMsg(addr1, addr2, addr3, addr4, addr5, addr6, addr7, addr8)
	msgs := []sdk.Msg{msg}
	fee := types.NewTestStdFee()

	// test rejection logic
	privs, accnums, seqs := []crypto.PrivKey{priv1, priv2, priv3, priv4, priv5, priv6, priv7, priv8},
		[]uint64{0, 0, 0, 0, 0, 0, 0, 0}, []uint64{0, 0, 0, 0, 0, 0, 0, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeTooManySignatures)
}

func TestEnsureSufficientMempoolFees(t *testing.T) {
	// setup
	input := setupTestInput()
	ctx := input.ctx.WithMinGasPrices(
		sdk.DecCoins{
			sdk.NewDecCoinFromDec("photino", sdk.NewDecWithPrec(50000000000000, sdk.Precision)), // 0.0001photino
			sdk.NewDecCoinFromDec("stake", sdk.NewDecWithPrec(10000000000000, sdk.Precision)),   // 0.000001stake
		},
	)

	testCases := []struct {
		input      StdFee
		expectedOK bool
	}{
		{NewStdFee(200000, sdk.Coins{}), false},
		{NewStdFee(200000, sdk.NewCoins(sdk.NewInt64Coin("photino", 5))), false},
		{NewStdFee(200000, sdk.NewCoins(sdk.NewInt64Coin("stake", 1))), false},
		{NewStdFee(200000, sdk.NewCoins(sdk.NewInt64Coin("stake", 2))), true},
		{NewStdFee(200000, sdk.NewCoins(sdk.NewInt64Coin("photino", 10))), true},
		{
			NewStdFee(
				200000,
				sdk.NewCoins(
					sdk.NewInt64Coin("photino", 10),
					sdk.NewInt64Coin("stake", 2),
				),
			),
			true,
		},
		{
			NewStdFee(
				200000,
				sdk.NewCoins(
					sdk.NewInt64Coin("atom", 5),
					sdk.NewInt64Coin("photino", 10),
					sdk.NewInt64Coin("stake", 2),
				),
			),
			true,
		},
	}

	for i, tc := range testCases {
		res := EnsureSufficientMempoolFees(ctx, tc.input)
		require.Equal(
			t, tc.expectedOK, res.IsOK(),
			"unexpected result; tc #%d, input: %v, log: %v", i, tc.input, res.Log,
		)
	}
}

// Test custom SignatureVerificationGasConsumer
func TestCustomSignatureVerificationGasConsumer(t *testing.T) {
	// setup
	input := setupTestInput()
	// setup an ante handler that only accepts PubKeyEd25519
	anteHandler := NewAnteHandler(input.ak, input.sk, func(meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params Params) sdk.Result {
		switch pubkey := pubkey.(type) {
		case ed25519.PubKeyEd25519:
			meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
			return sdk.Result{}
		default:
			return sdk.ErrInvalidPubKey(fmt.Sprintf("unrecognized public key type: %T", pubkey)).Result()
		}
	})
	ctx := input.ctx.WithBlockHeight(1)

	// verify that an secp256k1 account gets rejected
	priv1, _, addr1 := types.KeyTestPubAddr()
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	_ = acc1.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("atom", 150)))
	input.ak.SetAccount(ctx, acc1)

	var tx sdk.Tx
	msg := types.NewTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := types.NewTestStdFee()
	msgs := []sdk.Msg{msg}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInvalidPubKey)

	// verify that an ed25519 account gets accepted
	priv2 := ed25519.GenPrivKey()
	pub2 := priv2.PubKey()
	addr2 := sdk.AccAddress(pub2.Address())
	acc2 := input.ak.NewAccountWithAddress(ctx, addr2)
	require.NoError(t, acc2.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("atom", 150))))
	require.NoError(t, acc2.SetAccountNumber(1))
	input.ak.SetAccount(ctx, acc2)
	msg = types.NewTestMsg(addr2)
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{1}, []uint64{0}
	fee = types.NewTestStdFee()
	msgs = []sdk.Msg{msg}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}
