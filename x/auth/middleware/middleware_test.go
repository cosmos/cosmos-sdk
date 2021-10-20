package middleware_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Test that simulate transaction accurately estimates gas cost
func (s *MWTestSuite) TestSimulateGasCost() {
	ctx := s.SetupTest(false) // reset
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	accounts := s.createTestAccounts(ctx, 3)
	msgs := []sdk.Msg{
		testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress()),
		testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress()),
		testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[2].acc.GetAddress()),
	}
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	accSeqs := []uint64{0, 0, 0}
	privs := []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}
	accNums := []uint64{0, 1, 2}

	testCases := []TestCase{
		{
			"tx with 150atom fee",
			func() {
				txBuilder.SetFeeAmount(feeAmount)
				txBuilder.SetGasLimit(gasLimit)
			},
			true,
			true,
			nil,
		},
		{
			"with previously estimated gas",
			func() {
				simulatedGas := ctx.GasMeter().GasConsumed()

				accSeqs = []uint64{1, 1, 1}
				txBuilder.SetFeeAmount(feeAmount)
				txBuilder.SetGasLimit(simulatedGas)
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			s.runTestCase(ctx, txBuilder, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, ctx.ChainID(), tc)
		})
	}
}

// Test various error cases in the TxHandler control flow.
func (s *MWTestSuite) TestTxHandlerSigErrors() {
	ctx := s.SetupTest(false) // reset
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	priv0, _, addr0 := testdata.KeyTestPubAddr()
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	msgs := []sdk.Msg{
		testdata.NewTestMsg(addr0, addr1),
		testdata.NewTestMsg(addr0, addr2),
	}
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	// Variable data per test case
	var (
		privs   []cryptotypes.PrivKey
		accNums []uint64
		accSeqs []uint64
	)

	testCases := []TestCase{
		{
			"check no signatures fails",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{}, []uint64{}, []uint64{}

				// Create tx manually to test the tx's signers
				s.Require().NoError(txBuilder.SetMsgs(msgs...))
				tx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
				s.Require().NoError(err)
				// tx.GetSigners returns addresses in correct order: addr1, addr2, addr3
				expectedSigners := []sdk.AccAddress{addr0, addr1, addr2}
				s.Require().Equal(expectedSigners, tx.GetSigners())
			},
			false,
			false,
			sdkerrors.ErrNoSignatures,
		},
		{
			"num sigs dont match GetSigners",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"unrecognized account",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{priv0, priv1, priv2}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
			},
			false,
			false,
			sdkerrors.ErrUnknownAddress,
		},
		{
			"save the first account, but second is still unrecognized",
			func() {
				acc1 := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr0)
				s.app.AccountKeeper.SetAccount(ctx, acc1)
				err := s.app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, feeAmount)
				s.Require().NoError(err)
				err = s.app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr0, feeAmount)
				s.Require().NoError(err)
			},
			false,
			false,
			sdkerrors.ErrUnknownAddress,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			s.runTestCase(ctx, txBuilder, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, ctx.ChainID(), tc)
		})
	}
}

// Test logic around account number checking with one signer and many signers.
func (s *MWTestSuite) TestTxHandlerAccountNumbers() {
	ctx := s.SetupTest(false) // reset
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	accounts := s.createTestAccounts(ctx, 2)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []TestCase{
		{
			"good tx from one signer",
			func() {
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			true,
			nil,
		},
		{
			"new tx from wrong account number",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{1}, []uint64{1}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx from correct account number",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{1}
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
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{1, 0}, []uint64{2, 0}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx with correct account numbers",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{0, 1}, []uint64{2, 0}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			s.runTestCase(ctx, txBuilder, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, ctx.ChainID(), tc)
		})
	}
}

// Test logic around account number checking with many signers when BlockHeight is 0.
func (s *MWTestSuite) TestTxHandlerAccountNumbersAtBlockHeightZero() {
	ctx := s.SetupTest(false) // setup
	ctx = ctx.WithBlockHeight(0)
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	accounts := s.createTestAccounts(ctx, 2)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []TestCase{
		{
			"good tx from one signer",
			func() {
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			true,
			nil,
		},
		{
			"new tx from wrong account number",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{1}, []uint64{1}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx from correct account number",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{1}
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

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{1, 0}, []uint64{2, 0}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx with another signer and correct account numbers",
			func() {
				// Note that accNums is [0,0] at block 0.
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{0, 0}, []uint64{2, 0}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			s.runTestCase(ctx, txBuilder, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, ctx.ChainID(), tc)
		})
	}
}

// Test logic around sequence checking with one signer and many signers.
func (s *MWTestSuite) TestTxHandlerSequences() {
	ctx := s.SetupTest(false) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	accounts := s.createTestAccounts(ctx, 3)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []TestCase{
		{
			"good tx from one signer",
			func() {
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			true,
			nil,
		},
		{
			"test sending it again fails (replay protection)",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
		{
			"fix sequence, should pass",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{1}
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

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{2, 0, 0}
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
			sdkerrors.ErrWrongSequence,
		},
		{
			"tx from just second signer with incorrect sequence fails",
			func() {
				msg := testdata.NewTestMsg(accounts[1].acc.GetAddress())
				msgs = []sdk.Msg{msg}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv}, []uint64{1}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
		{
			"fix the sequence and it passes",
			func() {
				accSeqs = []uint64{1}
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

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{0, 1}, []uint64{3, 2}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			s.runTestCase(ctx, txBuilder, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, ctx.ChainID(), tc)
		})
	}
}

// Test logic around fee deduction.
func (s *MWTestSuite) TestTxHandlerFees() {
	ctx := s.SetupTest(false) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	priv0, _, addr0 := testdata.KeyTestPubAddr()

	acc1 := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr0)
	s.app.AccountKeeper.SetAccount(ctx, acc1)
	msgs := []sdk.Msg{testdata.NewTestMsg(addr0)}
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}

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
				accSeqs = []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"signer does not have enough funds to pay the fee",
			func() {
				err := testutil.FundAccount(s.app.BankKeeper, ctx, addr0, sdk.NewCoins(sdk.NewInt64Coin("atom", 149)))
				s.Require().NoError(err)
			},
			false,
			false,
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"signer as enough funds, should pass",
			func() {
				accNums = []uint64{acc1.GetAccountNumber()}

				modAcc := s.app.AccountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)

				s.Require().True(s.app.BankKeeper.GetAllBalances(ctx, modAcc.GetAddress()).Empty())
				require.True(sdk.IntEq(s.T(), s.app.BankKeeper.GetAllBalances(ctx, addr0).AmountOf("atom"), sdk.NewInt(149)))

				err := testutil.FundAccount(s.app.BankKeeper, ctx, addr0, sdk.NewCoins(sdk.NewInt64Coin("atom", 1)))
				s.Require().NoError(err)
			},
			false,
			true,
			nil,
		},
		{
			"signer doesn't have any more funds",
			func() {
				modAcc := s.app.AccountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)

				require.True(sdk.IntEq(s.T(), s.app.BankKeeper.GetAllBalances(ctx, modAcc.GetAddress()).AmountOf("atom"), sdk.NewInt(150)))
				require.True(sdk.IntEq(s.T(), s.app.BankKeeper.GetAllBalances(ctx, addr0).AmountOf("atom"), sdk.NewInt(0)))
			},
			false,
			false,
			sdkerrors.ErrInsufficientFunds,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			s.runTestCase(ctx, txBuilder, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, ctx.ChainID(), tc)
		})
	}
}

// Test logic around memo gas consumption.
func (s *MWTestSuite) TestTxHandlerMemoGas() {
	ctx := s.SetupTest(false) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	accounts := s.createTestAccounts(ctx, 1)
	msgs := []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
	privs, accNums, accSeqs := []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}

	// Variable data per test case
	var (
		feeAmount sdk.Coins
		gasLimit  uint64
	)

	testCases := []TestCase{
		{
			"tx does not have enough gas",
			func() {
				feeAmount = sdk.NewCoins(sdk.NewInt64Coin("atom", 0))
				gasLimit = 0
			},
			false,
			false,
			sdkerrors.ErrOutOfGas,
		},
		{
			"tx with memo doesn't have enough gas",
			func() {
				feeAmount = sdk.NewCoins(sdk.NewInt64Coin("atom", 0))
				gasLimit = 801
				txBuilder.SetMemo("abcininasidniandsinasindiansdiansdinaisndiasndiadninsd")
			},
			false,
			false,
			sdkerrors.ErrOutOfGas,
		},
		{
			"memo too large",
			func() {
				feeAmount = sdk.NewCoins(sdk.NewInt64Coin("atom", 0))
				gasLimit = 50000
				txBuilder.SetMemo(strings.Repeat("01234567890", 500))
			},
			false,
			false,
			sdkerrors.ErrMemoTooLarge,
		},
		{
			"tx with memo has enough gas",
			func() {
				feeAmount = sdk.NewCoins(sdk.NewInt64Coin("atom", 0))
				gasLimit = 50000
				txBuilder.SetMemo(strings.Repeat("0123456789", 10))
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			s.runTestCase(ctx, txBuilder, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, ctx.ChainID(), tc)
		})
	}
}

func (s *MWTestSuite) TestTxHandlerMultiSigner() {
	ctx := s.SetupTest(false) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	accounts := s.createTestAccounts(ctx, 3)
	msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
	msg2 := testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress())
	msg3 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[2].acc.GetAddress())
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []TestCase{
		{
			"signers in order",
			func() {
				msgs = []sdk.Msg{msg1, msg2, msg3}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
				txBuilder.SetMemo("Check signers are in expected order and different account numbers works")
			},
			false,
			true,
			nil,
		},
		{
			"change sequence numbers (only accounts 0 and 1 sign)",
			func() {
				msgs = []sdk.Msg{msg1}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{0, 1}, []uint64{1, 1}
			},
			false,
			true,
			nil,
		},
		{
			"change sequence numbers (only accounts 1 and 2 sign)",
			func() {
				msgs = []sdk.Msg{msg2}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[2].priv, accounts[0].priv}, []uint64{2, 0}, []uint64{1, 2}
			},
			false,
			true,
			nil,
		},
		{
			"everyone signs again",
			func() {
				msgs = []sdk.Msg{msg1, msg2, msg3}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{3, 2, 2}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			s.runTestCase(ctx, txBuilder, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, ctx.ChainID(), tc)
		})
	}
}

func (s *MWTestSuite) TestTxHandlerBadSignBytes() {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	accounts := s.createTestAccounts(ctx, 2)
	msg0 := testdata.NewTestMsg(accounts[0].acc.GetAddress())

	// Variable data per test case
	var (
		accNums   []uint64
		chainID   string
		feeAmount sdk.Coins
		gasLimit  uint64
		msgs      []sdk.Msg
		privs     []cryptotypes.PrivKey
		accSeqs   []uint64
	)

	testCases := []TestCase{
		{
			"test good tx and signBytes",
			func() {
				chainID = ctx.ChainID()
				feeAmount = testdata.NewTestFeeAmount()
				gasLimit = testdata.NewTestGasLimit()
				msgs = []sdk.Msg{msg0}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			true,
			nil,
		},
		{
			"test wrong chainID",
			func() {
				accSeqs = []uint64{1} // Back to correct accSeqs
				chainID = "chain-foo"
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"test wrong accSeqs",
			func() {
				chainID = ctx.ChainID() // Back to correct chainID
				accSeqs = []uint64{2}
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
		{
			"test wrong accNums",
			func() {
				accSeqs = []uint64{1} // Back to correct accSeqs
				accNums = []uint64{1}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"test wrong msg",
			func() {
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[1].acc.GetAddress())}
			},
			false,
			false,
			sdkerrors.ErrInvalidPubKey,
		},
		{
			"test wrong fee gas",
			func() {
				msgs = []sdk.Msg{msg0} // Back to correct msgs
				feeAmount = testdata.NewTestFeeAmount()
				gasLimit = testdata.NewTestGasLimit() + 100
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"test wrong fee amount",
			func() {
				feeAmount = testdata.NewTestFeeAmount()
				feeAmount[0].Amount = feeAmount[0].Amount.AddRaw(100)
				gasLimit = testdata.NewTestGasLimit()
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"test wrong signer if public key exist",
			func() {
				feeAmount = testdata.NewTestFeeAmount()
				gasLimit = testdata.NewTestGasLimit()
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv}, []uint64{0}, []uint64{1}
			},
			false,
			false,
			sdkerrors.ErrInvalidPubKey,
		},
		{
			"test wrong signer if public doesn't exist",
			func() {
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[1].acc.GetAddress())}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{1}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrInvalidPubKey,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			s.runTestCase(ctx, txBuilder, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, chainID, tc)
		})
	}
}

func (s *MWTestSuite) TestTxHandlerSetPubKey() {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	accounts := s.createTestAccounts(ctx, 2)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []TestCase{
		{
			"test good tx",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
			},
			false,
			true,
			nil,
		},
		{
			"make sure public key has been set (tx itself should fail because of replay protection)",
			func() {
				// Make sure public key has been set from previous test.
				acc0 := s.app.AccountKeeper.GetAccount(ctx, accounts[0].acc.GetAddress())
				s.Require().Equal(acc0.GetPubKey(), accounts[0].priv.PubKey())
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
		{
			"test public key not found",
			func() {
				// See above, `privs` still holds the private key of accounts[0].
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[1].acc.GetAddress())}
			},
			false,
			false,
			sdkerrors.ErrInvalidPubKey,
		},
		{
			"make sure public key is not set, when tx has no pubkey or signature",
			func() {
				// Make sure public key has not been set from previous test.
				acc1 := s.app.AccountKeeper.GetAccount(ctx, accounts[1].acc.GetAddress())
				s.Require().Nil(acc1.GetPubKey())

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv}, []uint64{1}, []uint64{0}
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[1].acc.GetAddress())}
				txBuilder.SetMsgs(msgs...)
				txBuilder.SetFeeAmount(feeAmount)
				txBuilder.SetGasLimit(gasLimit)

				// Manually create tx, and remove signature.
				tx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
				s.Require().NoError(err)
				txBuilder, err := s.clientCtx.TxConfig.WrapTxBuilder(tx)
				s.Require().NoError(err)
				s.Require().NoError(txBuilder.SetSignatures())

				// Run txHandler manually, expect ErrNoSignatures.
				_, err = s.txHandler.CheckTx(sdk.WrapSDKContext(ctx), txBuilder.GetTx(), abci.RequestCheckTx{})
				s.Require().Error(err)
				s.Require().True(errors.Is(err, sdkerrors.ErrNoSignatures))

				// Make sure public key has not been set.
				acc1 = s.app.AccountKeeper.GetAccount(ctx, accounts[1].acc.GetAddress())
				s.Require().Nil(acc1.GetPubKey())

				// Set incorrect accSeq, to generate incorrect signature.
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv}, []uint64{1}, []uint64{1}
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
		{
			"make sure previous public key has been set after wrong signature",
			func() {
				// Make sure public key has been set, as SetPubKeyMiddleware
				// is called before all signature verification middlewares.
				acc1 := s.app.AccountKeeper.GetAccount(ctx, accounts[1].acc.GetAddress())
				s.Require().Equal(acc1.GetPubKey(), accounts[1].priv.PubKey())
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			s.runTestCase(ctx, txBuilder, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, ctx.ChainID(), tc)
		})
	}
}

func generatePubKeysAndSignatures(n int, msg []byte, _ bool) (pubkeys []cryptotypes.PubKey, signatures [][]byte) {
	pubkeys = make([]cryptotypes.PubKey, n)
	signatures = make([][]byte, n)
	for i := 0; i < n; i++ {
		var privkey cryptotypes.PrivKey = secp256k1.GenPrivKey()

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

func expectedGasCostByKeys(pubkeys []cryptotypes.PubKey) uint64 {
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
	genPubKeys := func(n int) []cryptotypes.PubKey {
		var ret []cryptotypes.PubKey
		for i := 0; i < n; i++ {
			ret = append(ret, secp256k1.GenPrivKey().PubKey())
		}
		return ret
	}
	singleKey := secp256k1.GenPrivKey().PubKey()
	singleLevelMultiKey := kmultisig.NewLegacyAminoPubKey(4, genPubKeys(5))
	multiLevelSubKey1 := kmultisig.NewLegacyAminoPubKey(4, genPubKeys(5))
	multiLevelSubKey2 := kmultisig.NewLegacyAminoPubKey(4, genPubKeys(5))
	multiLevelMultiKey := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{
		multiLevelSubKey1, multiLevelSubKey2, secp256k1.GenPrivKey().PubKey()})
	type args struct {
		pub cryptotypes.PubKey
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
			require.Equal(t, tc.want, middleware.CountSubKeys(tc.args.pub))
		})
	}
}

func (s *MWTestSuite) TestTxHandlerSigLimitExceeded() {
	ctx := s.SetupTest(false) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	accounts := s.createTestAccounts(ctx, 8)
	var addrs []sdk.AccAddress
	var privs []cryptotypes.PrivKey
	for i := 0; i < 8; i++ {
		addrs = append(addrs, accounts[i].acc.GetAddress())
		privs = append(privs, accounts[i].priv)
	}
	msgs := []sdk.Msg{testdata.NewTestMsg(addrs...)}
	accNums, accSeqs := []uint64{0, 1, 2, 3, 4, 5, 6, 7}, []uint64{0, 0, 0, 0, 0, 0, 0, 0}
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	testCases := []TestCase{
		{
			"test rejection logic",
			func() {},
			false,
			false,
			sdkerrors.ErrTooManySignatures,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			s.runTestCase(ctx, txBuilder, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, ctx.ChainID(), tc)
		})
	}
}

// Test custom SignatureVerificationGasConsumer
func (s *MWTestSuite) TestCustomSignatureVerificationGasConsumer() {
	ctx := s.SetupTest(false) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	txHandler, err := middleware.NewDefaultTxHandler(
		middleware.TxHandlerOptions{
			AccountKeeper:   s.app.AccountKeeper,
			BankKeeper:      s.app.BankKeeper,
			FeegrantKeeper:  s.app.FeeGrantKeeper,
			SignModeHandler: s.clientCtx.TxConfig.SignModeHandler(),
			SigGasConsumer: func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error {
				switch pubkey := sig.PubKey.(type) {
				case *ed25519.PubKey:
					meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
					return nil
				default:
					return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
				}
			},
		},
	)
	s.Require().NoError(err)

	s.Require().NoError(err)

	// Same data for every test cases
	accounts := s.createTestAccounts(ctx, 1)
	txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())
	txBuilder.SetMsgs(testdata.NewTestMsg(accounts[0].acc.GetAddress()))

	// Variable data per test case
	var (
		accNums []uint64
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []TestCase{
		{
			"verify that an secp256k1 account gets rejected",
			func() {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrInvalidPubKey,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			tc.malleate()

			tx, txBytes, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
			s.Require().NoError(err)
			_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx, abci.RequestDeliverTx{Tx: txBytes})
			s.Require().Error(err)
			s.Require().True(errors.Is(err, tc.expErr))
		})
	}
}

func (s *MWTestSuite) TestTxHandlerReCheck() {
	ctx := s.SetupTest(false) // setup
	// Set recheck=true
	ctx = ctx.WithIsReCheckTx(true)
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	accounts := s.createTestAccounts(ctx, 1)

	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)

	msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
	msgs := []sdk.Msg{msg}
	s.Require().NoError(txBuilder.SetMsgs(msgs...))

	txBuilder.SetMemo("thisisatestmemo")

	// test that operations skipped on recheck do not run
	privs, accNums, accSeqs := []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
	tx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	s.Require().NoError(err)

	// make signature array empty which would normally cause ValidateBasicMiddleware and SigVerificationMiddleware fail
	// since these middlewares don't run on recheck, the tx should pass the middleware
	txBuilder, err = s.clientCtx.TxConfig.WrapTxBuilder(tx)
	s.Require().NoError(err)
	s.Require().NoError(txBuilder.SetSignatures())

	_, err = s.txHandler.CheckTx(sdk.WrapSDKContext(ctx), txBuilder.GetTx(), abci.RequestCheckTx{Type: abci.CheckTxType_Recheck})
	s.Require().Nil(err, "TxHandler errored on recheck unexpectedly: %v", err)

	tx, _, err = s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	s.Require().NoError(err)
	txBytes, err := json.Marshal(tx)
	s.Require().Nil(err, "Error marshalling tx: %v", err)
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
		s.app.AccountKeeper.SetParams(ctx, tc.params)

		_, err = s.txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx, abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_Recheck})

		s.Require().NotNil(err, "tx does not fail on recheck with updated params in test case: %s", tc.name)

		// reset parameters to default values
		s.app.AccountKeeper.SetParams(ctx, types.DefaultParams())
	}

	// require that local mempool fee check is still run on recheck since validator may change minFee between check and recheck
	// create new minimum gas price so txhandler fails on recheck
	ctx = ctx.WithMinGasPrices([]sdk.DecCoin{{
		Denom:  "dnecoin", // fee does not have this denom
		Amount: sdk.NewDec(5),
	}})
	_, err = s.txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx, abci.RequestCheckTx{})

	s.Require().NotNil(err, "txhandler on recheck did not fail when mingasPrice was changed")
	// reset min gasprice
	ctx = ctx.WithMinGasPrices(sdk.DecCoins{})

	// remove funds for account so txhandler fails on recheck
	s.app.AccountKeeper.SetAccount(ctx, accounts[0].acc)
	balances := s.app.BankKeeper.GetAllBalances(ctx, accounts[0].acc.GetAddress())
	err = s.app.BankKeeper.SendCoinsFromAccountToModule(ctx, accounts[0].acc.GetAddress(), minttypes.ModuleName, balances)
	s.Require().NoError(err)

	_, err = s.txHandler.CheckTx(sdk.WrapSDKContext(ctx), tx, abci.RequestCheckTx{})
	s.Require().NotNil(err, "txhandler on recheck did not fail once feePayer no longer has sufficient funds")
}
