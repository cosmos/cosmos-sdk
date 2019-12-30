package ante_test

import (
	"encoding/json"
	"errors"
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
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// run the tx through the anteHandler and ensure its valid
func checkValidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, simulate bool) {
	_, err := anteHandler(ctx, tx, simulate)
	require.Nil(t, err)
}

// run the tx through the anteHandler and ensure it fails with the given code
func checkInvalidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, simulate bool, expErr error) {
	_, err := anteHandler(ctx, tx, simulate)
	require.NotNil(t, err)
	require.True(t, errors.Is(expErr, err))
}

// Test that simulate transaction accurately estimates gas cost
func TestSimulateGasCost(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	ctx = ctx.WithBlockHeight(1)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)
	acc3 := app.AccountKeeper.NewAccountWithAddress(ctx, addr3)
	acc3.SetCoins(types.NewTestCoins())
	require.NoError(t, acc3.SetAccountNumber(2))
	app.AccountKeeper.SetAccount(ctx, acc3)

	// set up msgs and fee
	var tx sdk.Tx
	msg1 := types.NewTestMsg(addr1, addr2)
	msg2 := types.NewTestMsg(addr3, addr1)
	msg3 := types.NewTestMsg(addr2, addr3)
	msgs := []sdk.Msg{msg1, msg2, msg3}
	fee := types.NewTestStdFee()

	// signers in order. accnums are all 0 because it is in genesis block
	privs, accnums, seqs := []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)

	cc, _ := ctx.CacheContext()
	newCtx, err := anteHandler(cc, tx, true)
	require.Nil(t, err, "transaction failed on simulate mode")

	simulatedGas := newCtx.GasMeter().GasConsumed()
	fee.Gas = simulatedGas

	// update tx with simulated gas estimate
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	_, err = anteHandler(ctx, tx, false)

	require.Nil(t, err, "transaction failed with gas estimate")
}

// Test various error cases in the AnteHandler control flow.
func TestAnteHandlerSigErrors(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

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
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrNoSignatures)

	// test num sigs dont match GetSigners
	privs, accNums, seqs = []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnauthorized)

	// test an unrecognized account
	privs, accNums, seqs = []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnknownAddress)

	// save the first account, but second is still unrecognized
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(fee.Amount)
	app.AccountKeeper.SetAccount(ctx, acc1)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnknownAddress)
}

// Test logic around account number checking with one signer and many signers.
func TestAnteHandlerAccountNumbers(t *testing.T) {
	// setup
	app, ctx := createTestApp(false)
	ctx = ctx.WithBlockHeight(1)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)

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
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnauthorized)

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
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnauthorized)

	// correct account numbers
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{2, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around account number checking with many signers when BlockHeight is 0.
func TestAnteHandlerAccountNumbersAtBlockHeightZero(t *testing.T) {
	// setup
	app, ctx := createTestApp(false)
	ctx = ctx.WithBlockHeight(0)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()

	// set the accounts, we don't need the acc numbers as it is in the genesis block
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	app.AccountKeeper.SetAccount(ctx, acc1)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)

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
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnauthorized)

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
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnauthorized)

	// correct account numbers
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 0}, []uint64{2, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around sequence checking with one signer and many signers.
func TestAnteHandlerSequences(t *testing.T) {
	// setup
	app, ctx := createTestApp(false)
	ctx = ctx.WithBlockHeight(1)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)
	acc3 := app.AccountKeeper.NewAccountWithAddress(ctx, addr3)
	acc3.SetCoins(types.NewTestCoins())
	require.NoError(t, acc3.SetAccountNumber(2))
	app.AccountKeeper.SetAccount(ctx, acc3)

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
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnauthorized)

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
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnauthorized)

	// tx from just second signer with incorrect sequence fails
	msg = types.NewTestMsg(addr2)
	msgs = []sdk.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{1}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnauthorized)

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
	app, ctx := createTestApp(true)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)

	// msg and signatures
	var tx sdk.Tx
	msg := types.NewTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := types.NewTestStdFee()
	msgs := []sdk.Msg{msg}

	// signer does not have enough funds to pay the fee
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrInsufficientFunds)

	acc1.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("atom", 149)))
	app.AccountKeeper.SetAccount(ctx, acc1)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrInsufficientFunds)

	require.True(t, app.SupplyKeeper.GetModuleAccount(ctx, types.FeeCollectorName).GetCoins().Empty())
	require.True(sdk.IntEq(t, app.AccountKeeper.GetAccount(ctx, addr1).GetCoins().AmountOf("atom"), sdk.NewInt(149)))

	acc1.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("atom", 150)))
	app.AccountKeeper.SetAccount(ctx, acc1)
	checkValidTx(t, anteHandler, ctx, tx, false)

	require.True(sdk.IntEq(t, app.SupplyKeeper.GetModuleAccount(ctx, types.FeeCollectorName).GetCoins().AmountOf("atom"), sdk.NewInt(150)))
	require.True(sdk.IntEq(t, app.AccountKeeper.GetAccount(ctx, addr1).GetCoins().AmountOf("atom"), sdk.NewInt(0)))
}

// Test logic around memo gas consumption.
func TestAnteHandlerMemoGas(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	ctx = ctx.WithBlockHeight(1)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)

	// msg and signatures
	var tx sdk.Tx
	msg := types.NewTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := types.NewStdFee(0, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))

	// tx does not have enough gas
	tx = types.NewTestTx(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrOutOfGas)

	// tx with memo doesn't have enough gas
	fee = types.NewStdFee(801, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))
	tx = types.NewTestTxWithMemo(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee, "abcininasidniandsinasindiansdiansdinaisndiasndiadninsd")
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrOutOfGas)

	// memo too large
	fee = types.NewStdFee(50000, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))
	tx = types.NewTestTxWithMemo(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee, strings.Repeat("01234567890", 500))
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrMemoTooLarge)

	// tx with memo has enough gas
	fee = types.NewStdFee(50000, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))
	tx = types.NewTestTxWithMemo(ctx, []sdk.Msg{msg}, privs, accnums, seqs, fee, strings.Repeat("0123456789", 10))
	checkValidTx(t, anteHandler, ctx, tx, false)
}

func TestAnteHandlerMultiSigner(t *testing.T) {
	// setup
	app, ctx := createTestApp(false)
	ctx = ctx.WithBlockHeight(1)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)
	acc3 := app.AccountKeeper.NewAccountWithAddress(ctx, addr3)
	acc3.SetCoins(types.NewTestCoins())
	require.NoError(t, acc3.SetAccountNumber(2))
	app.AccountKeeper.SetAccount(ctx, acc3)

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
	app, ctx := createTestApp(true)
	ctx = ctx.WithBlockHeight(1)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)

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
	errUnauth := sdkerrors.ErrUnauthorized

	cases := []struct {
		chainID string
		accnum  uint64
		seq     uint64
		fee     types.StdFee
		msgs    []sdk.Msg
		err     error
	}{
		{chainID2, 0, 1, fee, msgs, errUnauth},                              // test wrong chain_id
		{chainID, 0, 2, fee, msgs, errUnauth},                               // test wrong seqs
		{chainID, 1, 1, fee, msgs, errUnauth},                               // test wrong accnum
		{chainID, 0, 1, fee, []sdk.Msg{types.NewTestMsg(addr2)}, errUnauth}, // test wrong msg
		{chainID, 0, 1, fee2, msgs, errUnauth},                              // test wrong fee
		{chainID, 0, 1, fee3, msgs, errUnauth},                              // test wrong fee
	}

	privs, seqs = []crypto.PrivKey{priv1}, []uint64{1}
	for _, cs := range cases {
		tx := types.NewTestTxWithSignBytes(
			msgs, privs, accnums, seqs, fee,
			types.StdSignBytes(cs.chainID, cs.accnum, cs.seq, cs.fee, cs.msgs, ""),
			"",
		)
		checkInvalidTx(t, anteHandler, ctx, tx, false, cs.err)
	}

	// test wrong signer if public key exist
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{0}, []uint64{1}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrInvalidPubKey)

	// test wrong signer if public doesn't exist
	msg = types.NewTestMsg(addr2)
	msgs = []sdk.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv1}, []uint64{1}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrInvalidPubKey)
}

func TestAnteHandlerSetPubKey(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	ctx = ctx.WithBlockHeight(1)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	_, _, addr2 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(types.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)

	var tx sdk.Tx

	// test good tx and set public key
	msg := types.NewTestMsg(addr1)
	msgs := []sdk.Msg{msg}
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := types.NewTestStdFee()
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	acc1 = app.AccountKeeper.GetAccount(ctx, addr1)
	require.Equal(t, acc1.GetPubKey(), priv1.PubKey())

	// test public key not found
	msg = types.NewTestMsg(addr2)
	msgs = []sdk.Msg{msg}
	tx = types.NewTestTx(ctx, msgs, privs, []uint64{1}, seqs, fee)
	sigs := tx.(types.StdTx).Signatures
	sigs[0].PubKey = nil
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrInvalidPubKey)

	acc2 = app.AccountKeeper.GetAccount(ctx, addr2)
	require.Nil(t, acc2.GetPubKey())

	// test invalid signature and public key
	tx = types.NewTestTx(ctx, msgs, privs, []uint64{1}, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrInvalidPubKey)

	acc2 = app.AccountKeeper.GetAccount(ctx, addr2)
	require.Nil(t, acc2.GetPubKey())
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
			cost += types.DefaultParams().SigVerifyCostED25519
		case strings.Contains(pubkeyType, "secp256k1"):
			cost += types.DefaultParams().SigVerifyCostSecp256k1
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
		tt := tt
		t.Run(tt.name, func(T *testing.T) {
			require.Equal(t, tt.want, types.CountSubKeys(tt.args.pub))
		})
	}
}

func TestAnteHandlerSigLimitExceeded(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	ctx = ctx.WithBlockHeight(1)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()
	priv4, _, addr4 := types.KeyTestPubAddr()
	priv5, _, addr5 := types.KeyTestPubAddr()
	priv6, _, addr6 := types.KeyTestPubAddr()
	priv7, _, addr7 := types.KeyTestPubAddr()
	priv8, _, addr8 := types.KeyTestPubAddr()

	addrs := []sdk.AccAddress{addr1, addr2, addr3, addr4, addr5, addr6, addr7, addr8}

	// set the accounts
	for i, addr := range addrs {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		acc.SetCoins(types.NewTestCoins())
		acc.SetAccountNumber(uint64(i))
		app.AccountKeeper.SetAccount(ctx, acc)
	}

	var tx sdk.Tx
	msg := types.NewTestMsg(addr1, addr2, addr3, addr4, addr5, addr6, addr7, addr8)
	msgs := []sdk.Msg{msg}
	fee := types.NewTestStdFee()

	// test rejection logic
	privs, accnums, seqs := []crypto.PrivKey{priv1, priv2, priv3, priv4, priv5, priv6, priv7, priv8},
		[]uint64{0, 1, 2, 3, 4, 5, 6, 7}, []uint64{0, 0, 0, 0, 0, 0, 0, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrTooManySignatures)
}

// Test custom SignatureVerificationGasConsumer
func TestCustomSignatureVerificationGasConsumer(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	ctx = ctx.WithBlockHeight(1)
	// setup an ante handler that only accepts PubKeyEd25519
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, func(meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params types.Params) error {
		switch pubkey := pubkey.(type) {
		case ed25519.PubKeyEd25519:
			meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
			return nil
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
		}
	})

	// verify that an secp256k1 account gets rejected
	priv1, _, addr1 := types.KeyTestPubAddr()
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	_ = acc1.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("atom", 150)))
	app.AccountKeeper.SetAccount(ctx, acc1)

	var tx sdk.Tx
	msg := types.NewTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := types.NewTestStdFee()
	msgs := []sdk.Msg{msg}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrInvalidPubKey)

	// verify that an ed25519 account gets accepted
	priv2 := ed25519.GenPrivKey()
	pub2 := priv2.PubKey()
	addr2 := sdk.AccAddress(pub2.Address())
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	require.NoError(t, acc2.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("atom", 150))))
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)
	msg = types.NewTestMsg(addr2)
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{1}, []uint64{0}
	fee = types.NewTestStdFee()
	msgs = []sdk.Msg{msg}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

func TestAnteHandlerReCheck(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	// set blockheight and recheck=true
	ctx = ctx.WithBlockHeight(1)
	ctx = ctx.WithIsReCheckTx(true)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	// priv2, _, addr2 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(types.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)

	antehandler := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, ante.DefaultSigVerificationGasConsumer)

	// test that operations skipped on recheck do not run

	msg := types.NewTestMsg(addr1)
	msgs := []sdk.Msg{msg}
	fee := types.NewTestStdFee()

	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx := types.NewTestTxWithMemo(ctx, msgs, privs, accnums, seqs, fee, "thisisatestmemo")

	// make signature array empty which would normally cause ValidateBasicDecorator and SigVerificationDecorator fail
	// since these decorators don't run on recheck, the tx should pass the antehandler
	stdTx := tx.(types.StdTx)
	stdTx.Signatures = []types.StdSignature{}

	_, err := antehandler(ctx, stdTx, false)
	require.Nil(t, err, "AnteHandler errored on recheck unexpectedly: %v", err)

	tx = types.NewTestTxWithMemo(ctx, msgs, privs, accnums, seqs, fee, "thisisatestmemo")
	txBytes, err := json.Marshal(tx)
	require.Nil(t, err, "Error marshalling tx: %v", err)
	ctx = ctx.WithTxBytes(txBytes)

	// require that state machine param-dependent checking is still run on recheck since parameters can change between check and recheck
	testCases := []struct {
		name   string
		params types.Params
	}{
		{"memo size check", types.NewParams(1, types.DefaultTxSigLimit, types.DefaultTxSizeCostPerByte, types.DefaultSigVerifyCostED25519, types.DefaultSigVerifyCostSecp256k1)},
		{"txsize check", types.NewParams(types.DefaultMaxMemoCharacters, types.DefaultTxSigLimit, 10000000, types.DefaultSigVerifyCostED25519, types.DefaultSigVerifyCostSecp256k1)},
		{"sig verify cost check", types.NewParams(types.DefaultMaxMemoCharacters, types.DefaultTxSigLimit, types.DefaultTxSizeCostPerByte, types.DefaultSigVerifyCostED25519, 100000000)},
	}
	for _, tc := range testCases {
		// set testcase parameters
		app.AccountKeeper.SetParams(ctx, tc.params)

		_, err := antehandler(ctx, tx, false)

		require.NotNil(t, err, "tx does not fail on recheck with updated params in test case: %s", tc.name)

		// reset parameters to default values
		app.AccountKeeper.SetParams(ctx, types.DefaultParams())
	}

	// require that local mempool fee check is still run on recheck since validator may change minFee between check and recheck
	// create new minimum gas price so antehandler fails on recheck
	ctx = ctx.WithMinGasPrices([]sdk.DecCoin{{
		Denom:  "dnecoin", // fee does not have this denom
		Amount: sdk.NewDec(5),
	}})
	_, err = antehandler(ctx, tx, false)
	require.NotNil(t, err, "antehandler on recheck did not fail when mingasPrice was changed")
	// reset min gasprice
	ctx = ctx.WithMinGasPrices(sdk.DecCoins{})

	// remove funds for account so antehandler fails on recheck
	acc1.SetCoins(sdk.Coins{})
	app.AccountKeeper.SetAccount(ctx, acc1)

	_, err = antehandler(ctx, tx, false)
	require.NotNil(t, err, "antehandler on recheck did not fail once feePayer no longer has sufficient funds")
}
