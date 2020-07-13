package ante_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec/testdata"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
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

// Increment the sequences in a sequences array. To be called after each
// IncrementSequenceDecorator.
func incrementSeq(seqs []uint64) []uint64 {
	var newSeqs []uint64

	for _, value := range seqs {
		newSeqs = append(newSeqs, value+1)
	}

	return newSeqs
}

// TestAccount represents an account used in the tests below
type TestAccount struct {
	priv crypto.PrivKey
	acc  types.AccountI
}

type AnteTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	anteHandler sdk.AnteHandler
	ctx         sdk.Context
	clientCtx   client.Context
	txBuilder   client.TxBuilder
}

func (suite *AnteTestSuite) SetupTest() {
	suite.app, suite.ctx = createTestApp(true)
	suite.ctx = suite.ctx.WithBlockHeight(1)
	suite.anteHandler = ante.NewAnteHandler(suite.app.AccountKeeper, suite.app.BankKeeper, *suite.app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

	// set up the TxBuilder
	encodingConfig := simappparams.MakeEncodingConfig()
	suite.clientCtx = client.Context{}.
		WithTxGenerator(encodingConfig.TxGenerator).
		WithJSONMarshaler(encodingConfig.Marshaler)
	suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()
}

// Create `numAccs` accounts, and return all relevant information about them.
func (suite *AnteTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := types.KeyTestPubAddr()
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr)
		err := acc.SetAccountNumber(uint64(i))
		suite.Require().NoError(err)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		suite.app.BankKeeper.SetBalances(suite.ctx, acc.GetAddress(), types.NewTestCoins())

		accounts = append(accounts, TestAccount{priv, acc})

	}

	return accounts
}

// Helper function to create a tx given multiple inputs.
func (suite *AnteTestSuite) CreateTestTx(privs []crypto.PrivKey, accNums []uint64, seqs []uint64) xauthsigning.SigFeeMemoTx {
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2, err := tx.SignWithPrivKey(priv, accNums[i], seqs[i], suite.ctx.ChainID(), suite.clientCtx.TxGenerator, suite.txBuilder)
		suite.Require().NoError(err)

		sigsV2 = append(sigsV2, sigV2)
	}
	suite.txBuilder.SetSignatures(sigsV2...)

	return suite.txBuilder.GetTx()
}

// Test that simulate transaction accurately estimates gas cost
func (suite *AnteTestSuite) TestSimulateGasCost() {
	suite.SetupTest() // reset
	accounts := suite.CreateTestAccounts(3)

	fmt.Println("chainID", suite.ctx.ChainID())

	// Same data for every test cases
	msgs := []sdk.Msg{
		testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress()),
		testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress()),
		testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[2].acc.GetAddress()),
	}
	fee := types.NewTestStdFee()
	seqs := []uint64{0, 0, 0}
	privs := []crypto.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}
	accNums := []uint64{0, 1, 2}

	testCases := []struct {
		msg      string
		malleate func()
		simulate bool
		expPass  bool
	}{
		{
			"tx with 150atom fee",
			func() {
				suite.txBuilder.SetFeeAmount(fee.GetAmount())
			},
			true,
			true,
		},
		{
			"with previously estimated gas",
			func() {
				simulatedGas := suite.ctx.GasMeter().GasConsumed()
				fee.Gas = simulatedGas

				// Round 2: update tx with exact simulated gas estimate
				suite.txBuilder.SetGasLimit(fee.Gas)
			},
			false,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			// Create new txBuilder for each test case
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()
			suite.txBuilder.SetMsgs(msgs...)

			tc.malleate()

			fmt.Println("SEQS", seqs)
			tx := suite.CreateTestTx(privs, accNums, seqs)
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
				seqs = incrementSeq(seqs)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// Test various error cases in the AnteHandler control flow.
func (suite *AnteTestSuite) TestAnteHandlerSigErrors() {
	suite.SetupTest() // reset

	var (
		accounts []TestAccount
		privs    []crypto.PrivKey
		accNums  []uint64
		seqs     []uint64
	)

	testCases := []struct {
		msg      string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"no signatures",
			func() {
				privs, accNums, seqs = []crypto.PrivKey{}, []uint64{}, []uint64{}

				// tx := suite.CreateTestTx(privs, accNums, seqs)
				// tx.GetSigners returns addresses in correct order: addr1, addr2, addr3
				expectedSigners := []sdk.AccAddress{
					accounts[0].acc.GetAddress(),
					accounts[1].acc.GetAddress(),
					accounts[2].acc.GetAddress(),
				}
				stdTx := suite.txBuilder.GetTx().(types.StdTx)
				suite.Require().Equal(expectedSigners, stdTx.GetSigners())
			},
			false,
			false,
			sdkerrors.ErrNoSignatures,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()
			accounts = suite.CreateTestAccounts(3)

			msgs := []sdk.Msg{
				testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress()),
				testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[2].acc.GetAddress()),
			}
			suite.txBuilder.SetMsgs(msgs...)

			fee := types.NewTestStdFee()
			suite.txBuilder.SetFeeAmount(fee.GetAmount())

			tc.malleate()

			tx := suite.CreateTestTx(privs, accNums, seqs)
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
				seqs = incrementSeq(seqs)
			} else {
				suite.Require().Error(err)
			}
		})
	}

	// // test num sigs dont match GetSigners
	// privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
	// tx = createTestTx(privs, accNums, seqs, ctx, clientCtx.TxGenerator, txBuilder, t)
	// checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnauthorized)

	// // test an unrecognized account
	// privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	// tx = createTestTx(privs, accNums, seqs, ctx, clientCtx.TxGenerator, txBuilder, t)
	// checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnknownAddress)

	// // save the first account, but second is still unrecognized
	// acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, accounts[0].acc.GetAddress())
	// app.AccountKeeper.SetAccount(ctx, acc1)
	// err := app.BankKeeper.SetBalances(ctx, accounts[0].acc.GetAddress(), fee.Amount)
	// require.NoError(t, err)
	// checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnknownAddress)
}

// Test logic around account number checking with one signer and many signers.
func TestAnteHandlerAccountNumbers(t *testing.T) {
	// setup
	app, ctx := createTestApp(false)
	ctx = ctx.WithBlockHeight(1)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.BankKeeper, *app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	err := app.BankKeeper.SetBalances(ctx, addr1, types.NewTestCoins())
	require.NoError(t, err)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)
	err = app.BankKeeper.SetBalances(ctx, addr2, types.NewTestCoins())
	require.NoError(t, err)

	// msg and signatures
	var tx sdk.Tx
	msg := testdata.NewTestMsg(addr1)
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
	msg1 := testdata.NewTestMsg(addr1, addr2)
	msg2 := testdata.NewTestMsg(addr2, addr1)
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
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.BankKeeper, *app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()

	// set the accounts, we don't need the acc numbers as it is in the genesis block
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	err := app.BankKeeper.SetBalances(ctx, addr1, types.NewTestCoins())
	require.NoError(t, err)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)
	err = app.BankKeeper.SetBalances(ctx, addr2, types.NewTestCoins())
	require.NoError(t, err)

	// msg and signatures
	var tx sdk.Tx
	msg := testdata.NewTestMsg(addr1)
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
	msg1 := testdata.NewTestMsg(addr1, addr2)
	msg2 := testdata.NewTestMsg(addr2, addr1)
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
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.BankKeeper, *app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	app.BankKeeper.SetBalances(ctx, addr1, types.NewTestCoins())
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)
	app.BankKeeper.SetBalances(ctx, addr2, types.NewTestCoins())
	acc3 := app.AccountKeeper.NewAccountWithAddress(ctx, addr3)
	require.NoError(t, acc3.SetAccountNumber(2))
	app.AccountKeeper.SetAccount(ctx, acc3)
	app.BankKeeper.SetBalances(ctx, addr3, types.NewTestCoins())

	// msg and signatures
	var tx sdk.Tx
	msg := testdata.NewTestMsg(addr1)
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
	msg1 := testdata.NewTestMsg(addr1, addr2)
	msg2 := testdata.NewTestMsg(addr3, addr1)
	msgs = []sdk.Msg{msg1, msg2}

	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{2, 0, 0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// replay fails
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnauthorized)

	// tx from just second signer with incorrect sequence fails
	msg = testdata.NewTestMsg(addr2)
	msgs = []sdk.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{1}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrUnauthorized)

	// fix the sequence and it passes
	tx = types.NewTestTx(ctx, msgs, []crypto.PrivKey{priv2}, []uint64{1}, []uint64{1}, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// another tx from both of them that passes
	msg = testdata.NewTestMsg(addr1, addr2)
	msgs = []sdk.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{3, 2}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around fee deduction.
func TestAnteHandlerFees(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.BankKeeper, *app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)

	// msg and signatures
	var tx sdk.Tx
	msg := testdata.NewTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := types.NewTestStdFee()
	msgs := []sdk.Msg{msg}

	// signer does not have enough funds to pay the fee
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrInsufficientFunds)

	app.AccountKeeper.SetAccount(ctx, acc1)
	app.BankKeeper.SetBalances(ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("atom", 149)))
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrInsufficientFunds)

	modAcc := app.AccountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)

	require.True(t, app.BankKeeper.GetAllBalances(ctx, modAcc.GetAddress()).Empty())
	require.True(sdk.IntEq(t, app.BankKeeper.GetAllBalances(ctx, addr1).AmountOf("atom"), sdk.NewInt(149)))

	app.AccountKeeper.SetAccount(ctx, acc1)
	app.BankKeeper.SetBalances(ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("atom", 150)))
	checkValidTx(t, anteHandler, ctx, tx, false)

	require.True(sdk.IntEq(t, app.BankKeeper.GetAllBalances(ctx, modAcc.GetAddress()).AmountOf("atom"), sdk.NewInt(150)))
	require.True(sdk.IntEq(t, app.BankKeeper.GetAllBalances(ctx, addr1).AmountOf("atom"), sdk.NewInt(0)))
}

// Test logic around memo gas consumption.
func TestAnteHandlerMemoGas(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	ctx = ctx.WithBlockHeight(1)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.BankKeeper, *app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)

	// msg and signatures
	var tx sdk.Tx
	msg := testdata.NewTestMsg(addr1)
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
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.BankKeeper, *app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	priv3, _, addr3 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	app.BankKeeper.SetBalances(ctx, addr1, types.NewTestCoins())
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)
	app.BankKeeper.SetBalances(ctx, addr2, types.NewTestCoins())
	acc3 := app.AccountKeeper.NewAccountWithAddress(ctx, addr3)
	require.NoError(t, acc3.SetAccountNumber(2))
	app.AccountKeeper.SetAccount(ctx, acc3)
	app.BankKeeper.SetBalances(ctx, addr3, types.NewTestCoins())

	// set up msgs and fee
	var tx sdk.Tx
	msg1 := testdata.NewTestMsg(addr1, addr2)
	msg2 := testdata.NewTestMsg(addr3, addr1)
	msg3 := testdata.NewTestMsg(addr2, addr3)
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
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.BankKeeper, *app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	app.BankKeeper.SetBalances(ctx, addr1, types.NewTestCoins())
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)
	app.BankKeeper.SetBalances(ctx, addr2, types.NewTestCoins())

	var tx sdk.Tx
	msg := testdata.NewTestMsg(addr1)
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
		{chainID2, 0, 1, fee, msgs, errUnauth},                                 // test wrong chain_id
		{chainID, 0, 2, fee, msgs, errUnauth},                                  // test wrong seqs
		{chainID, 1, 1, fee, msgs, errUnauth},                                  // test wrong accnum
		{chainID, 0, 1, fee, []sdk.Msg{testdata.NewTestMsg(addr2)}, errUnauth}, // test wrong msg
		{chainID, 0, 1, fee2, msgs, errUnauth},                                 // test wrong fee
		{chainID, 0, 1, fee3, msgs, errUnauth},                                 // test wrong fee
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
	msg = testdata.NewTestMsg(addr2)
	msgs = []sdk.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv1}, []uint64{1}, []uint64{0}
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdkerrors.ErrInvalidPubKey)
}

func TestAnteHandlerSetPubKey(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	ctx = ctx.WithBlockHeight(1)
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.BankKeeper, *app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()
	_, _, addr2 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	app.BankKeeper.SetBalances(ctx, addr1, types.NewTestCoins())
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)
	app.BankKeeper.SetBalances(ctx, addr2, types.NewTestCoins())

	var tx sdk.Tx

	// test good tx and set public key
	msg := testdata.NewTestMsg(addr1)
	msgs := []sdk.Msg{msg}
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := types.NewTestStdFee()
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	acc1 = app.AccountKeeper.GetAccount(ctx, addr1)
	require.Equal(t, acc1.GetPubKey(), priv1.PubKey())

	// test public key not found
	msg = testdata.NewTestMsg(addr2)
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

func generatePubKeysAndSignatures(n int, msg []byte, _ bool) (pubkeys []crypto.PubKey, signatures [][]byte) {
	pubkeys = make([]crypto.PubKey, n)
	signatures = make([][]byte, n)
	for i := 0; i < n; i++ {
		var privkey crypto.PrivKey
		privkey = secp256k1.GenPrivKey()

		// TODO: also generate ed25519 keys as below when ed25519 keys are
		//  actually supported, https://github.com/cosmos/cosmos-sdk/issues/4789
		// for now this fails:
		//if rand.Int63()%2 == 0 {
		//	privkey = ed25519.GenPrivKey()
		//} else {
		//	privkey = secp256k1.GenPrivKey()
		//}

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
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.BankKeeper, *app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

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
		err := acc.SetAccountNumber(uint64(i))
		require.NoError(t, err)
		app.AccountKeeper.SetAccount(ctx, acc)
		app.BankKeeper.SetBalances(ctx, addr, types.NewTestCoins())
	}

	var tx sdk.Tx
	msg := testdata.NewTestMsg(addr1, addr2, addr3, addr4, addr5, addr6, addr7, addr8)
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
	anteHandler := ante.NewAnteHandler(app.AccountKeeper, app.BankKeeper, *app.IBCKeeper, func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error {
		switch pubkey := sig.PubKey.(type) {
		case ed25519.PubKeyEd25519:
			meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
			return nil
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
		}
	}, types.LegacyAminoJSONHandler{})

	// verify that an secp256k1 account gets rejected
	priv1, _, addr1 := types.KeyTestPubAddr()
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	app.BankKeeper.SetBalances(ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("atom", 150)))

	var tx sdk.Tx
	msg := testdata.NewTestMsg(addr1)
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

	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr2, sdk.NewCoins(sdk.NewInt64Coin("atom", 150))))
	require.NoError(t, acc2.SetAccountNumber(1))
	app.AccountKeeper.SetAccount(ctx, acc2)
	msg = testdata.NewTestMsg(addr2)
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
	require.NoError(t, acc1.SetAccountNumber(0))
	app.AccountKeeper.SetAccount(ctx, acc1)
	app.BankKeeper.SetBalances(ctx, addr1, types.NewTestCoins())

	antehandler := ante.NewAnteHandler(app.AccountKeeper, app.BankKeeper, *app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

	// test that operations skipped on recheck do not run

	msg := testdata.NewTestMsg(addr1)
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
	app.AccountKeeper.SetAccount(ctx, acc1)
	app.BankKeeper.SetBalances(ctx, addr1, sdk.NewCoins())

	_, err = antehandler(ctx, tx, false)
	require.NotNil(t, err, "antehandler on recheck did not fail once feePayer no longer has sufficient funds")
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}
