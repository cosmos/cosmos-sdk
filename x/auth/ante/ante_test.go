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

	"github.com/cosmos/cosmos-sdk/codec/testdata"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Test that simulate transaction accurately estimates gas cost
func (suite *AnteTestSuite) TestSimulateGasCost() {
	suite.SetupTest(true) // reset

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(3)
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
		desc     string
		malleate func()
		simulate bool
		expPass  bool
	}{
		{
			"tx with 150atom fee",
			func() {
				suite.txBuilder.SetFeeAmount(fee.GetAmount())
				suite.txBuilder.SetGasLimit(fee.GetGas())
			},
			true,
			true,
		},
		{
			"with previously estimated gas",
			func() {
				simulatedGas := suite.ctx.GasMeter().GasConsumed()
				fee.Gas = simulatedGas

				seqs = []uint64{1, 1, 1}
				suite.txBuilder.SetFeeAmount(fee.GetAmount())
				suite.txBuilder.SetGasLimit(fee.Gas)
			},
			false,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			suite.txBuilder.SetMsgs(msgs...)

			tc.malleate()

			tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// Test various error cases in the AnteHandler control flow.
func (suite *AnteTestSuite) TestAnteHandlerSigErrors() {
	suite.SetupTest(true) // reset

	// Same data for every test cases
	priv0, _, addr0 := types.KeyTestPubAddr()
	priv1, _, addr1 := types.KeyTestPubAddr()
	priv2, _, addr2 := types.KeyTestPubAddr()
	msgs := []sdk.Msg{
		testdata.NewTestMsg(addr0, addr1),
		testdata.NewTestMsg(addr0, addr2),
	}
	fee := types.NewTestStdFee()

	// Variable data per test case
	var (
		privs   []crypto.PrivKey
		accNums []uint64
		seqs    []uint64
	)

	testCases := []struct {
		desc     string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"no signatures fails",
			func() {
				privs, accNums, seqs = []crypto.PrivKey{}, []uint64{}, []uint64{}

				// tx := suite.CreateTestTx(privs, accNums, seqs)
				// tx.GetSigners returns addresses in correct order: addr1, addr2, addr3
				expectedSigners := []sdk.AccAddress{addr0, addr1, addr2}
				stdTx := suite.txBuilder.GetTx().(types.StdTx)
				suite.Require().Equal(expectedSigners, stdTx.GetSigners())
			},
			false,
			false,
			sdkerrors.ErrNoSignatures,
		},
		{
			"num sigs dont match GetSigners",
			func() {
				privs, accNums, seqs = []crypto.PrivKey{priv0}, []uint64{0}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"unrecognized account",
			func() {
				privs, accNums, seqs = []crypto.PrivKey{priv0, priv1, priv2}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
			},
			false,
			false,
			sdkerrors.ErrUnknownAddress,
		},
		{
			"save the first account, but second is still unrecognized",
			func() {
				acc1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr0)
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc1)
				err := suite.app.BankKeeper.SetBalances(suite.ctx, addr0, fee.Amount)
				suite.Require().NoError(err)
			},
			false,
			false,
			sdkerrors.ErrUnknownAddress,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			suite.txBuilder.SetMsgs(msgs...)
			suite.txBuilder.SetFeeAmount(fee.GetAmount())
			suite.txBuilder.SetGasLimit(fee.GetGas())

			tc.malleate()

			tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
}

// Test logic around account number checking with one signer and many signers.
func (suite *AnteTestSuite) TestAnteHandlerAccountNumbers() {
	suite.SetupTest(false) // reset

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(2)
	fee := types.NewTestStdFee()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []crypto.PrivKey
		seqs    []uint64
	)

	testCases := []struct {
		desc     string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"good tx from one signer",
			func() {
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			true,
			nil,
		},
		{
			"new tx from wrong account number",
			func() {
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{1}, []uint64{1}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx from correct account number",
			func() {
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{1}
			},
			false,
			true,
			nil,
		},
		{
			"new tx with another signer and incorrect account numbers",
			func() {
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2}
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{1, 0}, []uint64{2, 0}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx with correct account numbers",
			func() {
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{0, 1}, []uint64{2, 0}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			tc.malleate()

			suite.txBuilder.SetMsgs(msgs...)
			suite.txBuilder.SetFeeAmount(fee.GetAmount())
			suite.txBuilder.SetGasLimit(fee.GetGas())

			tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
}

// Test logic around account number checking with many signers when BlockHeight is 0.
func (suite *AnteTestSuite) TestAnteHandlerAccountNumbersAtBlockHeightZero() {
	suite.SetupTest(false) // setup
	suite.ctx = suite.ctx.WithBlockHeight(0)

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(2)
	fee := types.NewTestStdFee()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []crypto.PrivKey
		seqs    []uint64
	)

	testCases := []struct {
		desc     string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"good tx from one signer",
			func() {
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			true,
			nil,
		},
		{
			"new tx from wrong account number",
			func() {
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{1}, []uint64{1}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx from correct account number",
			func() {
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{1}
			},
			false,
			true,
			nil,
		},
		{
			"new tx with another signer and incorrect account numbers",
			func() {
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2}

				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{1, 0}, []uint64{2, 0}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx with another signer and correct account numbers",
			func() {
				// Note that accNums is [0,0] at block 0.
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{0, 0}, []uint64{2, 0}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			tc.malleate()

			suite.txBuilder.SetMsgs(msgs...)
			suite.txBuilder.SetFeeAmount(fee.GetAmount())
			suite.txBuilder.SetGasLimit(fee.GetGas())

			tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
}

// Test logic around sequence checking with one signer and many signers.
func (suite *AnteTestSuite) TestAnteHandlerSequences() {
	suite.SetupTest(false) // setup

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(3)
	fee := types.NewTestStdFee()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []crypto.PrivKey
		seqs    []uint64
	)

	testCases := []struct {
		desc     string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"good tx from one signer",
			func() {
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			true,
			nil,
		},
		{
			"test sending it again fails (replay protection)",
			func() {
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"fix sequence, should pass",
			func() {
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{1}
			},
			false,
			true,
			nil,
		},
		{
			"new tx with another signer and correct sequences",
			func() {
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2}

				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{2, 0, 0}
			},
			false,
			true,
			nil,
		},
		{
			"replay fails",
			func() {},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"tx from just second signer with incorrect sequence fails",
			func() {
				msg := testdata.NewTestMsg(accounts[1].acc.GetAddress())
				msgs = []sdk.Msg{msg}
				privs, accNums, seqs = []crypto.PrivKey{accounts[1].priv}, []uint64{1}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"fix the sequence and it passes",
			func() {
				seqs = []uint64{1}
			},
			false,
			true,
			nil,
		},
		{
			"fix the sequence and it passes",
			func() {
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{0, 1}, []uint64{3, 2}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			tc.malleate()

			suite.txBuilder.SetMsgs(msgs...)
			suite.txBuilder.SetFeeAmount(fee.GetAmount())
			suite.txBuilder.SetGasLimit(fee.GetGas())

			tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
}

// Test logic around fee deduction.
func (suite *AnteTestSuite) TestAnteHandlerFees() {
	suite.SetupTest(true) // setup

	// Same data for every test cases
	priv0, _, addr0 := types.KeyTestPubAddr()

	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr0)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc1)
	msg := testdata.NewTestMsg(addr0)
	fee := types.NewTestStdFee()
	privs, accNums, seqs := []crypto.PrivKey{priv0}, []uint64{0}, []uint64{0}

	testCases := []struct {
		desc     string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"signer has no funds",
			func() {
				seqs = []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"signer does not have enough funds to pay the fee",
			func() {
				suite.app.BankKeeper.SetBalances(suite.ctx, addr0, sdk.NewCoins(sdk.NewInt64Coin("atom", 149)))
			},
			false,
			false,
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"signer as enough funds, should pass",
			func() {
				modAcc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, types.FeeCollectorName)

				suite.Require().True(suite.app.BankKeeper.GetAllBalances(suite.ctx, modAcc.GetAddress()).Empty())
				require.True(sdk.IntEq(suite.T(), suite.app.BankKeeper.GetAllBalances(suite.ctx, addr0).AmountOf("atom"), sdk.NewInt(149)))

				suite.app.BankKeeper.SetBalances(suite.ctx, addr0, sdk.NewCoins(sdk.NewInt64Coin("atom", 150)))
			},
			false,
			true,
			nil,
		},
		{
			"signer doesn't have any more funds",
			func() {
				modAcc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, types.FeeCollectorName)

				require.True(sdk.IntEq(suite.T(), suite.app.BankKeeper.GetAllBalances(suite.ctx, modAcc.GetAddress()).AmountOf("atom"), sdk.NewInt(150)))
				require.True(sdk.IntEq(suite.T(), suite.app.BankKeeper.GetAllBalances(suite.ctx, addr0).AmountOf("atom"), sdk.NewInt(0)))
			},
			false,
			false,
			sdkerrors.ErrInsufficientFunds,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			tc.malleate()

			suite.txBuilder.SetMsgs(msg)
			suite.txBuilder.SetFeeAmount(fee.GetAmount())
			suite.txBuilder.SetGasLimit(fee.GetGas())

			tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
}

// Test logic around memo gas consumption.
func (suite *AnteTestSuite) TestAnteHandlerMemoGas() {
	suite.SetupTest(false) // setup

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(1)
	msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
	privs, accNums, seqs := []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}

	// Variable data per test case
	var (
		fee types.StdFee
	)

	testCases := []struct {
		desc     string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"tx does not have enough gas",
			func() {
				fee = types.NewStdFee(0, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))
			},
			false,
			false,
			sdkerrors.ErrOutOfGas,
		},
		{
			"tx with memo doesn't have enough gas",
			func() {
				fee = types.NewStdFee(801, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))
				suite.txBuilder.SetMemo("abcininasidniandsinasindiansdiansdinaisndiasndiadninsd")
			},
			false,
			false,
			sdkerrors.ErrOutOfGas,
		},
		{
			"memo too large",
			func() {
				fee = types.NewStdFee(50000, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))
				suite.txBuilder.SetMemo(strings.Repeat("01234567890", 500))
			},
			false,
			false,
			sdkerrors.ErrMemoTooLarge,
		},
		{
			"tx with memo has enough gas",
			func() {
				fee = types.NewStdFee(50000, sdk.NewCoins(sdk.NewInt64Coin("atom", 0)))
				suite.txBuilder.SetMemo(strings.Repeat("0123456789", 10))
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			tc.malleate()

			suite.txBuilder.SetMsgs(msg)
			suite.txBuilder.SetFeeAmount(fee.GetAmount())
			suite.txBuilder.SetGasLimit(fee.GetGas())

			tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
}

func (suite *AnteTestSuite) TestAnteHandlerMultiSigner() {
	suite.SetupTest(false) // setup

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(3)
	msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
	msg2 := testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress())
	msg3 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[2].acc.GetAddress())
	fee := types.NewTestStdFee()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []crypto.PrivKey
		seqs    []uint64
	)

	testCases := []struct {
		desc     string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"signers in order",
			func() {
				msgs = []sdk.Msg{msg1, msg2, msg3}
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
				suite.txBuilder.SetMemo("Check signers are in expected order and different account numbers works")
			},
			false,
			true,
			nil,
		},
		{
			"change sequence numbers (only accounts 0 and 1 sign)",
			func() {
				msgs = []sdk.Msg{msg1}
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{0, 1}, []uint64{1, 1}
			},
			false,
			true,
			nil,
		},
		{
			"change sequence numbers (only accounts 1 and 2 sign)",
			func() {
				msgs = []sdk.Msg{msg2}
				privs, accNums, seqs = []crypto.PrivKey{accounts[2].priv, accounts[0].priv}, []uint64{2, 0}, []uint64{1, 2}
			},
			false,
			true,
			nil,
		},
		{
			"everyone signs again",
			func() {
				msgs = []sdk.Msg{msg1, msg2, msg3}
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{3, 2, 2}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			tc.malleate()

			suite.txBuilder.SetMsgs(msgs...)
			suite.txBuilder.SetFeeAmount(fee.GetAmount())
			suite.txBuilder.SetGasLimit(fee.GetGas())

			tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
}

func (suite *AnteTestSuite) TestAnteHandlerBadSignBytes() {
	suite.SetupTest(false) // setup

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(2)
	msg0 := testdata.NewTestMsg(accounts[0].acc.GetAddress())

	// Variable data per test case
	var (
		accNums []uint64
		chainID string
		fee     types.StdFee
		msgs    []sdk.Msg
		privs   []crypto.PrivKey
		seqs    []uint64
	)

	testCases := []struct {
		desc     string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"test good tx and signBytes",
			func() {
				chainID = suite.ctx.ChainID()
				fee = types.NewTestStdFee()
				msgs = []sdk.Msg{msg0}
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			true,
			nil,
		},
		{
			"test wrong chainID",
			func() {
				seqs = []uint64{1} // Back to correct seqs
				chainID = "chain-foo"
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"test wrong seqs",
			func() {
				chainID = suite.ctx.ChainID() // Back to correct chainID
				seqs = []uint64{2}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"test wrong accNums",
			func() {
				seqs = []uint64{1} // Back to correct seqs
				accNums = []uint64{1}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		// {
		// 	"test wrong msg",
		// 	func() {
		// 		accNums = []uint64{0} // Back to correct accNums
		// 		msgs = []sdk.Msg{testdata.NewTestMsg(accounts[1].acc.GetAddress())}
		// 	},
		// 	false,
		// 	false,
		// 	sdkerrors.ErrUnauthorized,
		// },
		{
			"test wrong fee gas",
			func() {
				msgs = []sdk.Msg{msg0} // Back to correct msgs
				fee = types.NewTestStdFee()
				fee.Gas += 100
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"test wrong fee amount",
			func() {
				fee = types.NewTestStdFee()
				fee.Amount[0].Amount = fee.Amount[0].Amount.AddRaw(100)
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"test wrong signer if public key exist",
			func() {
				fee = types.NewTestStdFee()
				privs, accNums, seqs = []crypto.PrivKey{accounts[1].priv}, []uint64{0}, []uint64{1}
			},
			false,
			false,
			sdkerrors.ErrInvalidPubKey,
		},
		{
			"test wrong signer if public doesn't exist",
			func() {
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[1].acc.GetAddress())}
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{1}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrInvalidPubKey,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			tc.malleate()

			suite.txBuilder.SetMsgs(msgs...)
			suite.txBuilder.SetFeeAmount(fee.GetAmount())
			suite.txBuilder.SetGasLimit(fee.GetGas())

			tx := suite.CreateTestTx(privs, accNums, seqs, chainID)
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
}

func (suite *AnteTestSuite) TestAnteHandlerSetPubKey() {
	suite.SetupTest(false) // setup

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(2)
	fee := types.NewTestStdFee()
	privs, accNums, seqs := []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}

	// Variable data per test case
	var (
		msgs []sdk.Msg
	)

	testCases := []struct {
		desc     string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"test good tx",
			func() {
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
			},
			false,
			true,
			nil,
		},
		{
			"make sure public key has been set (tx should fail because of replay protection)",
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, accounts[0].acc.GetAddress())
				suite.Require().Equal(acc.GetPubKey(), accounts[0].priv.PubKey())
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		// {
		// 	"test public key not found",
		// 	func() {
		// 		msgs = []sdk.Msg{testdata.NewTestMsg(accounts[1].acc.GetAddress())}
		// 		tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
		// 		sigs := tx.(types.StdTx).Signatures
		// 		sigs[0].PubKey = nil

		// 		// Run anteHandler manually here, because we manually modified the tx.
		// 		_, err := suite.anteHandler(suite.ctx, tx, false)
		// 		suite.Require().Error(err)
		// 		fmt.Println("ERR", err)
		// 		suite.Require().True(errors.Is(err, sdkerrors.ErrInvalidPubKey))
		// 	},
		// 	false,
		// 	true,
		// 	nil,
		// },
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			tc.malleate()

			suite.txBuilder.SetMsgs(msgs...)
			suite.txBuilder.SetFeeAmount(fee.GetAmount())
			suite.txBuilder.SetGasLimit(fee.GetGas())

			tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
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
	testCases := []struct {
		name string
		args args
		want int
	}{
		{"single key", args{singleKey}, 1},
		{"single level multikey", args{singleLevelMultiKey}, 5},
		{"multi level multikey", args{multiLevelMultiKey}, 11},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(T *testing.T) {
			require.Equal(t, tc.want, types.CountSubKeys(tc.args.pub))
		})
	}
}

func (suite *AnteTestSuite) TestAnteHandlerSigLimitExceeded() {
	suite.SetupTest(true) // setup

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(8)
	var addrs []sdk.AccAddress
	var privs []crypto.PrivKey
	for i := 0; i < 8; i++ {
		addrs = append(addrs, accounts[i].acc.GetAddress())
		privs = append(privs, accounts[i].priv)
	}
	msgs := []sdk.Msg{testdata.NewTestMsg(addrs...)}
	accNums, seqs := []uint64{0, 1, 2, 3, 4, 5, 6, 7}, []uint64{0, 0, 0, 0, 0, 0, 0, 0}
	fee := types.NewTestStdFee()

	testCases := []struct {
		desc     string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"test rejection logic",
			func() {},
			false,
			false,
			sdkerrors.ErrTooManySignatures,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			tc.malleate()

			suite.txBuilder.SetMsgs(msgs...)
			suite.txBuilder.SetFeeAmount(fee.GetAmount())
			suite.txBuilder.SetGasLimit(fee.GetGas())

			tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
}

// Test custom SignatureVerificationGasConsumer
func (suite *AnteTestSuite) TestCustomSignatureVerificationGasConsumer() {
	suite.SetupTest(true) // setup

	// setup an ante handler that only accepts PubKeyEd25519
	suite.anteHandler = ante.NewAnteHandler(suite.app.AccountKeeper, suite.app.BankKeeper, *suite.app.IBCKeeper, func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error {
		switch pubkey := sig.PubKey.(type) {
		case ed25519.PubKeyEd25519:
			meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
			return nil
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
		}
	}, types.LegacyAminoJSONHandler{})

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(1)
	fee := types.NewTestStdFee()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []crypto.PrivKey
		seqs    []uint64
	)

	testCases := []struct {
		desc     string
		malleate func()
		simulate bool
		expPass  bool
		expErr   error
	}{
		{
			"verify that an secp256k1 account gets rejected",
			func() {
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				privs, accNums, seqs = []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrInvalidPubKey,
		},
		{
			"verify that an ed25519 account gets accepted",
			func() {
				priv1 := ed25519.GenPrivKey()
				pub1 := priv1.PubKey()
				addr1 := sdk.AccAddress(pub1.Address())
				acc1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)

				suite.Require().NoError(suite.app.BankKeeper.SetBalances(suite.ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("atom", 150))))
				suite.Require().NoError(acc1.SetAccountNumber(1))
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc1)
				msg := testdata.NewTestMsg(addr1)
				privs, accNums, seqs = []crypto.PrivKey{priv1}, []uint64{1}, []uint64{0}
				msgs = []sdk.Msg{msg}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

			tc.malleate()

			suite.txBuilder.SetMsgs(msgs...)
			suite.txBuilder.SetFeeAmount(fee.GetAmount())
			suite.txBuilder.SetGasLimit(fee.GetGas())

			tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
			newCtx, err := suite.anteHandler(suite.ctx, tx, tc.simulate)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(newCtx)

				suite.ctx = newCtx
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
}

func (suite *AnteTestSuite) TestAnteHandlerReCheck() {
	suite.SetupTest(true) // setup
	// Set recheck=true
	suite.ctx = suite.ctx.WithIsReCheckTx(true)
	suite.txBuilder = suite.clientCtx.TxGenerator.NewTxBuilder()

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(1)

	fee := types.NewTestStdFee()
	suite.txBuilder.SetFeeAmount(fee.GetAmount())
	suite.txBuilder.SetGasLimit(fee.GetGas())

	msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
	msgs := []sdk.Msg{msg}
	suite.txBuilder.SetMsgs(msgs...)

	suite.txBuilder.SetMemo("thisisatestmemo")

	// test that operations skipped on recheck do not run
	privs, accNums, seqs := []crypto.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
	tx := suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())

	// make signature array empty which would normally cause ValidateBasicDecorator and SigVerificationDecorator fail
	// since these decorators don't run on recheck, the tx should pass the antehandler
	stdTx := tx.(types.StdTx)
	stdTx.Signatures = []types.StdSignature{}

	_, err := suite.anteHandler(suite.ctx, stdTx, false)
	suite.Require().Nil(err, "AnteHandler errored on recheck unexpectedly: %v", err)

	tx = suite.CreateTestTx(privs, accNums, seqs, suite.ctx.ChainID())
	txBytes, err := json.Marshal(tx)
	suite.Require().Nil(err, "Error marshalling tx: %v", err)
	suite.ctx = suite.ctx.WithTxBytes(txBytes)

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
		suite.app.AccountKeeper.SetParams(suite.ctx, tc.params)

		_, err := suite.anteHandler(suite.ctx, tx, false)

		suite.Require().NotNil(err, "tx does not fail on recheck with updated params in test case: %s", tc.name)

		// reset parameters to default values
		suite.app.AccountKeeper.SetParams(suite.ctx, types.DefaultParams())
	}

	// require that local mempool fee check is still run on recheck since validator may change minFee between check and recheck
	// create new minimum gas price so antehandler fails on recheck
	suite.ctx = suite.ctx.WithMinGasPrices([]sdk.DecCoin{{
		Denom:  "dnecoin", // fee does not have this denom
		Amount: sdk.NewDec(5),
	}})
	_, err = suite.anteHandler(suite.ctx, tx, false)
	suite.Require().NotNil(err, "antehandler on recheck did not fail when mingasPrice was changed")
	// reset min gasprice
	suite.ctx = suite.ctx.WithMinGasPrices(sdk.DecCoins{})

	// remove funds for account so antehandler fails on recheck
	suite.app.AccountKeeper.SetAccount(suite.ctx, accounts[0].acc)
	suite.app.BankKeeper.SetBalances(suite.ctx, accounts[0].acc.GetAddress(), sdk.NewCoins())

	_, err = suite.anteHandler(suite.ctx, tx, false)
	suite.Require().NotNil(err, "antehandler on recheck did not fail once feePayer no longer has sufficient funds")
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}
