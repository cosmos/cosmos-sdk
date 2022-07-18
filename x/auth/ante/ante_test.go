package ante_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Test that simulate transaction accurately estimates gas cost
func TestSimulateGasCost(t *testing.T) {
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	testCases := []NewTestCase{
		{
			"tx with 150atom fee",
			func(suite *NewAnteTestSuite) {
				suite.txBuilder.SetFeeAmount(feeAmount)
				suite.txBuilder.SetGasLimit(gasLimit)
			},
			true,
			true,
			nil,
		},
		{
			"with previously estimated gas",
			func(suite *NewAnteTestSuite) {
				simulatedGas := suite.ctx.GasMeter().GasConsumed()
				suite.txBuilder.SetFeeAmount(feeAmount)
				suite.txBuilder.SetGasLimit(simulatedGas)
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			accounts := suite.CreateTestAccounts(3)

			suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accounts[0].acc.GetAddress(), authtypes.FeeCollectorName, feeAmount).Return(nil)

			msgs := []sdk.Msg{
				testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress()),
				testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress()),
				testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[2].acc.GetAddress()),
			}
			privs := []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}

			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate(suite)

			accSeqs := []uint64{0, 0, 0}
			accNums := []uint64{0, 1, 2}
			suite.RunTestCase(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
		})
	}
}

// Test various error cases in the AnteHandler control flow.
func TestAnteHandlerSigErrors(t *testing.T) {
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

	testCases := []NewTestCase{
		{
			"check no signatures fails",
			func(suite *NewAnteTestSuite) {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{}, []uint64{}, []uint64{}

				// Create tx manually to test the tx's signers
				require.NoError(t, suite.txBuilder.SetMsgs(msgs...))
				tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
				require.NoError(t, err)
				// tx.GetSigners returns addresses in correct order: addr1, addr2, addr3
				expectedSigners := []sdk.AccAddress{addr0, addr1, addr2}
				require.Equal(t, expectedSigners, tx.GetSigners())
			},
			false,
			false,
			sdkerrors.ErrNoSignatures,
		},
		{
			"num sigs dont match GetSigners",
			func(suite *NewAnteTestSuite) {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"unrecognized account",
			func(suite *NewAnteTestSuite) {
				privs, accNums, accSeqs = []cryptotypes.PrivKey{priv0, priv1, priv2}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
			},
			false,
			false,
			sdkerrors.ErrUnknownAddress,
		},
		{
			"save the first account, but second is still unrecognized",
			func(suite *NewAnteTestSuite) {
				acc1 := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr0)
				suite.accountKeeper.SetAccount(suite.ctx, acc1)
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrUnknownAddress,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)

			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate(suite)

			suite.RunTestCase(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
		})
	}
}

// Test logic around account number checking with one signer and many signers.
func TestAnteHandlerAccountNumbers(t *testing.T) {
	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []NewTestCase{
		{
			"good tx from one signer",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accounts[0].acc.GetAddress(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			true,
			nil,
		},
		{
			"new tx from wrong account number",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{1}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accounts[0].acc.GetAddress(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx from correct account number",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accounts[0].acc.GetAddress(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			true,
			nil,
		},
		{
			"new tx with another signer and incorrect account numbers",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(2)
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{2, 0}, []uint64{0, 0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accounts[0].acc.GetAddress(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx with correct account numbers",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(2)
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{0, 1}, []uint64{0, 0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			tc.malleate(suite)

			feeAmount := testdata.NewTestFeeAmount()
			gasLimit := testdata.NewTestGasLimit()
			suite.RunTestCase(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
		})
	}
}

// Test logic around account number checking with many signers when BlockHeight is 0.
func TestAnteHandlerAccountNumbersAtBlockHeightZero(t *testing.T) {
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []NewTestCase{
		{
			"good tx from one signer",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)

				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			true,
			nil,
		},
		{
			"new tx from wrong account number",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)

				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{1}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx from correct account number",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)

				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			true,
			nil,
		},
		{
			"new tx with another signer and incorrect account numbers",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(2)

				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{1, 0}, []uint64{0, 0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accounts[0].acc.GetAddress(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"new tx with another signer and correct account numbers",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(2)

				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2}

				// Note that accNums is [0,0] at block 0.
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{0, 0}, []uint64{0, 0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accounts[0].acc.GetAddress(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			suite.ctx = suite.ctx.WithBlockHeight(0)
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate(suite)

			suite.RunTestCase(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
		})
	}
}

// Test logic around sequence checking with one signer and many signers.
func TestAnteHandlerSequences(t *testing.T) {
	// Same data for every test cases
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []NewTestCase{
		{
			"good tx from one signer",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			true,
			nil,
		},
		{
			"test sending it again fails (replay protection)",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}

				// This will be called only once given that the second tx will fail
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

				// Send the same tx before running the test case, to trigger replay protection.
				var err error
				suite.ctx, err = suite.DeliverMsgs(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), false)
				require.NoError(t, err)
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
		{
			"fix sequence, should pass",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

				// Send the same tx before running the test case, then change the sequence to a valid one.
				var err error
				suite.ctx, err = suite.DeliverMsgs(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), false)
				require.NoError(t, err)

				// +1 the account sequence
				accSeqs = []uint64{1}
			},
			false,
			true,
			nil,
		},
		{
			"new tx with another signer and correct sequences",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(3)
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			true,
			nil,
		},
		{
			"replay fails",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(3)
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

				// Send the same tx before running the test case, to trigger replay protection.
				var err error
				suite.ctx, err = suite.DeliverMsgs(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), false)
				require.NoError(t, err)
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
		{
			"tx from just second signer with incorrect sequence fails",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(3)
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

				// Send the same tx before running the test case, to trigger replay protection.
				var err error
				suite.ctx, err = suite.DeliverMsgs(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), false)
				require.NoError(t, err)

				// Send a message using the second signer, this will fail given that the second signer already sent a TX,
				// thus the sequence (0) is incorrect.
				msg1 = testdata.NewTestMsg(accounts[1].acc.GetAddress())
				msgs = []sdk.Msg{msg1}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv}, []uint64{1}, []uint64{0}
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
		{
			"fix the sequence and it passes",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(3)
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

				// Send the same tx before running the test case, to trigger replay protection.
				var err error
				suite.ctx, err = suite.DeliverMsgs(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), false)
				require.NoError(t, err)

				// Send a message using the second signer, this will now pass given that the second signer already sent a TX
				// and the sequence was fixed (1).
				msg1 = testdata.NewTestMsg(accounts[1].acc.GetAddress())
				msgs = []sdk.Msg{msg1}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv}, []uint64{1}, []uint64{1}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate(suite)

			suite.RunTestCase(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
		})
	}
}

// Test logic around fee deduction.
func TestAnteHandlerFees(t *testing.T) {
	// Same data for every test cases
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []NewTestCase{
		{
			"signer has no funds",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accounts[0].acc.GetAddress(), gomock.Any(), feeAmount).Return(sdkerrors.ErrInsufficientFunds)
			},
			false,
			false,
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"signer does not have enough funds to pay the fee",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accounts[0].acc.GetAddress(), gomock.Any(), feeAmount).Return(sdkerrors.ErrInsufficientFunds)
			},
			false,
			false,
			sdkerrors.ErrInsufficientFunds,
		},
		{
			"signer as enough funds, should pass",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accounts[0].acc.GetAddress(), gomock.Any(), feeAmount).Return(nil)
			},
			false,
			true,
			nil,
		},
		{
			"signer doesn't have any more funds",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accounts[0].acc.GetAddress(), gomock.Any(), feeAmount).Return(sdkerrors.ErrInsufficientFunds)
			},
			false,
			false,
			sdkerrors.ErrInsufficientFunds,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate(suite)

			suite.RunTestCase(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
		})
	}
}

// Test logic around memo gas consumption.
func TestAnteHandlerMemoGas(t *testing.T) {
	var (
		accNums   []uint64
		msgs      []sdk.Msg
		privs     []cryptotypes.PrivKey
		accSeqs   []uint64
		feeAmount sdk.Coins
		gasLimit  uint64
	)

	testCases := []NewTestCase{
		{
			"tx does not have enough gas",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				feeAmount = sdk.NewCoins(sdk.NewInt64Coin("atom", 0))
				gasLimit = 0
			},
			false,
			false,
			sdkerrors.ErrOutOfGas,
		},
		{
			"tx with memo doesn't have enough gas",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				feeAmount = sdk.NewCoins(sdk.NewInt64Coin("atom", 0))
				gasLimit = 801
				suite.txBuilder.SetMemo("abcininasidniandsinasindiansdiansdinaisndiasndiadninsd")
			},
			false,
			false,
			sdkerrors.ErrOutOfGas,
		},
		{
			"memo too large",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				feeAmount = sdk.NewCoins(sdk.NewInt64Coin("atom", 0))
				gasLimit = 50000
				suite.txBuilder.SetMemo(strings.Repeat("01234567890", 500))
			},
			false,
			false,
			sdkerrors.ErrMemoTooLarge,
		},
		{
			"tx with memo has enough gas",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				feeAmount = sdk.NewCoins(sdk.NewInt64Coin("atom", 0))
				gasLimit = 60000
				suite.txBuilder.SetMemo(strings.Repeat("0123456789", 10))
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate(suite)

			suite.RunTestCase(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
		})
	}
}

func TestAnteHandlerMultiSigner(t *testing.T) {
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []NewTestCase{
		{
			"signers in order",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(3)
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress())
				msg3 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[2].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2, msg3}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
				suite.txBuilder.SetMemo("Check signers are in expected order and different account numbers works")
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			true,
			nil,
		},
		{
			"change sequence numbers (only accounts 0 and 1 sign)",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(3)
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress())
				msg3 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[2].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2, msg3}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{0, 0, 0}

				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

				var err error
				suite.ctx, err = suite.DeliverMsgs(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), false)
				require.NoError(t, err)

				msgs = []sdk.Msg{msg1}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv}, []uint64{0, 1}, []uint64{1, 1}
			},
			false,
			true,
			nil,
		},
		{
			"change sequence numbers (only accounts 1 and 2 sign)",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(3)
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress())
				msg3 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[2].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2, msg3}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{0, 0, 0}

				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
				var err error
				suite.ctx, err = suite.DeliverMsgs(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), false)
				require.NoError(t, err)

				msgs = []sdk.Msg{msg3}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv, accounts[2].priv}, []uint64{1, 2}, []uint64{1, 1}
			},
			false,
			true,
			nil,
		},
		{
			"everyone signs again",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(3)
				msg1 := testdata.NewTestMsg(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress())
				msg2 := testdata.NewTestMsg(accounts[2].acc.GetAddress(), accounts[0].acc.GetAddress())
				msg3 := testdata.NewTestMsg(accounts[1].acc.GetAddress(), accounts[2].acc.GetAddress())
				msgs = []sdk.Msg{msg1, msg2, msg3}

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{0, 0, 0}

				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
				var err error
				suite.ctx, err = suite.DeliverMsgs(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), false)
				require.NoError(t, err)

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}, []uint64{0, 1, 2}, []uint64{1, 1, 1}
			},
			false,
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate(suite)

			suite.RunTestCase(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
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
				chainID = suite.ctx.ChainID()
				feeAmount = testdata.NewTestFeeAmount()
				gasLimit = testdata.NewTestGasLimit()
				msgs = []sdk.Msg{msg0}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
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
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"test wrong accSeqs",
			func() {
				chainID = suite.ctx.ChainID() // Back to correct chainID
				accSeqs = []uint64{2}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
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
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrUnauthorized,
		},
		{
			"test wrong msg",
			func() {
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[1].acc.GetAddress())}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
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
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
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
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
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
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
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
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrInvalidPubKey,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate()

			suite.RunTestCase(privs, msgs, feeAmount, gasLimit, accNums, accSeqs, chainID, tc)
		})
	}
}

func TestAnteHandlerSetPubKey(t *testing.T) {
	// Same data for every test cases
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []NewTestCase{
		{
			"test good tx",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			true,
			nil,
		},
		{
			"make sure public key has been set (tx itself should fail because of replay protection)",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(1)

				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				var err error
				suite.ctx, err = suite.DeliverMsgs(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), false)
				require.NoError(t, err)

				// Make sure public key has been set from previous test.
				acc0 := suite.accountKeeper.GetAccount(suite.ctx, accounts[0].acc.GetAddress())
				require.Equal(t, acc0.GetPubKey(), accounts[0].priv.PubKey())
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
		{
			"test public key not found",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(2)
				// See above, `privs` still holds the private key of accounts[0].
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[1].acc.GetAddress())}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrInvalidPubKey,
		},
		{
			"make sure public key is not set, when tx has no pubkey or signature",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(2)
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				// Make sure public key has not been set from previous test.
				acc1 := suite.accountKeeper.GetAccount(suite.ctx, accounts[1].acc.GetAddress())
				require.Nil(t, acc1.GetPubKey())

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv}, []uint64{1}, []uint64{0}
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[1].acc.GetAddress())}
				suite.txBuilder.SetMsgs(msgs...)
				suite.txBuilder.SetFeeAmount(feeAmount)
				suite.txBuilder.SetGasLimit(gasLimit)

				// Manually create tx, and remove signature.
				tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
				require.NoError(t, err)
				txBuilder, err := suite.clientCtx.TxConfig.WrapTxBuilder(tx)
				require.NoError(t, err)
				require.NoError(t, txBuilder.SetSignatures())

				// Run anteHandler manually, expect ErrNoSignatures.
				_, err = suite.anteHandler(suite.ctx, txBuilder.GetTx(), false)
				require.Error(t, err)
				require.True(t, errors.Is(err, sdkerrors.ErrNoSignatures))

				// Make sure public key has not been set.
				acc1 = suite.accountKeeper.GetAccount(suite.ctx, accounts[1].acc.GetAddress())
				require.Nil(t, acc1.GetPubKey())

				// Set incorrect accSeq, to generate incorrect signature.
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv}, []uint64{1}, []uint64{1}
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
		{
			"make sure previous public key has been set after wrong signature",
			func(suite *NewAnteTestSuite) {
				accounts := suite.CreateTestAccounts(2)
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				// Make sure public key has not been set from previous test.
				acc1 := suite.accountKeeper.GetAccount(suite.ctx, accounts[1].acc.GetAddress())
				require.Nil(t, acc1.GetPubKey())

				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv}, []uint64{1}, []uint64{0}
				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[1].acc.GetAddress())}
				suite.txBuilder.SetMsgs(msgs...)
				suite.txBuilder.SetFeeAmount(feeAmount)
				suite.txBuilder.SetGasLimit(gasLimit)

				// Manually create tx, and remove signature.
				tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
				require.NoError(t, err)
				txBuilder, err := suite.clientCtx.TxConfig.WrapTxBuilder(tx)
				require.NoError(t, err)
				require.NoError(t, txBuilder.SetSignatures())

				// Run anteHandler manually, expect ErrNoSignatures.
				_, err = suite.anteHandler(suite.ctx, txBuilder.GetTx(), false)
				require.Error(t, err)
				require.True(t, errors.Is(err, sdkerrors.ErrNoSignatures))

				// Make sure public key has not been set.
				acc1 = suite.accountKeeper.GetAccount(suite.ctx, accounts[1].acc.GetAddress())
				require.Nil(t, acc1.GetPubKey())

				// Set incorrect accSeq, to generate incorrect signature.
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[1].priv}, []uint64{1}, []uint64{1}

				suite.ctx, err = suite.DeliverMsgs(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), false)
				require.Error(t, err)

				// Make sure public key has been set, as SetPubKeyDecorator
				// is called before all signature verification decorators.
				acc1 = suite.accountKeeper.GetAccount(suite.ctx, accounts[1].acc.GetAddress())
				require.Equal(t, acc1.GetPubKey(), accounts[1].priv.PubKey())
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrWrongSequence,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate(suite)

			suite.RunTestCase(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
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
		multiLevelSubKey1, multiLevelSubKey2, secp256k1.GenPrivKey().PubKey(),
	})
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
			require.Equal(t, tc.want, ante.CountSubKeys(tc.args.pub))
		})
	}
}

func TestAnteHandlerSigLimitExceeded(t *testing.T) {
	// Same data for every test cases
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []NewTestCase{
		{
			"test rejection logic",
			func(suite *NewAnteTestSuite) {
				// Same data for every test cases
				accounts := suite.CreateTestAccounts(8)
				var addrs []sdk.AccAddress
				for i := 0; i < 8; i++ {
					addrs = append(addrs, accounts[i].acc.GetAddress())
					privs = append(privs, accounts[i].priv)
				}
				msgs = []sdk.Msg{testdata.NewTestMsg(addrs...)}
				accNums, accSeqs = []uint64{0, 1, 2, 3, 4, 5, 6, 7}, []uint64{0, 0, 0, 0, 0, 0, 0, 0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrTooManySignatures,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate(suite)

			suite.RunTestCase(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
		})
	}
}

// Test custom SignatureVerificationGasConsumer
func TestCustomSignatureVerificationGasConsumer(t *testing.T) {

	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	// Variable data per test case
	var (
		accNums []uint64
		msgs    []sdk.Msg
		privs   []cryptotypes.PrivKey
		accSeqs []uint64
	)

	testCases := []NewTestCase{
		{
			"verify that an secp256k1 account gets rejected",
			func(suite *NewAnteTestSuite) {
				// setup an ante handler that only accepts PubKeyEd25519
				anteHandler, err := ante.NewAnteHandler(
					ante.HandlerOptions{
						AccountKeeper:   suite.accountKeeper,
						BankKeeper:      suite.bankKeeper,
						FeegrantKeeper:  suite.feeGrantKeeper,
						SignModeHandler: suite.clientCtx.TxConfig.SignModeHandler(),
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
				require.NoError(t, err)
				suite.anteHandler = anteHandler

				// Same data for every test cases
				accounts := suite.CreateTestAccounts(1)

				msgs = []sdk.Msg{testdata.NewTestMsg(accounts[0].acc.GetAddress())}
				privs, accNums, accSeqs = []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
				suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			false,
			false,
			sdkerrors.ErrInvalidPubKey,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			suite := SetupTestSuite(t, false)
			suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
			tc.malleate(suite)

			suite.RunTestCase(t, privs, msgs, feeAmount, gasLimit, accNums, accSeqs, suite.ctx.ChainID(), tc)
		})
	}
}

func (suite *AnteTestSuite) TestAnteHandlerReCheck() {
	suite.SetupTest(false) // setup
	// Set recheck=true
	suite.ctx = suite.ctx.WithIsReCheckTx(true)
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// Same data for every test cases
	accounts := suite.CreateTestAccounts(1)

	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	msg := testdata.NewTestMsg(accounts[0].acc.GetAddress())
	msgs := []sdk.Msg{msg}
	suite.Require().NoError(suite.txBuilder.SetMsgs(msgs...))

	suite.txBuilder.SetMemo("thisisatestmemo")

	// test that operations skipped on recheck do not run
	privs, accNums, accSeqs := []cryptotypes.PrivKey{accounts[0].priv}, []uint64{0}, []uint64{0}
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	// make signature array empty which would normally cause ValidateBasicDecorator and SigVerificationDecorator fail
	// since these decorators don't run on recheck, the tx should pass the antehandler
	txBuilder, err := suite.clientCtx.TxConfig.WrapTxBuilder(tx)
	suite.Require().NoError(err)

	suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
	_, err = suite.anteHandler(suite.ctx, txBuilder.GetTx(), false)
	suite.Require().Nil(err, "AnteHandler errored on recheck unexpectedly: %v", err)

	tx, err = suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)
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
		err := suite.accountKeeper.SetParams(suite.ctx, tc.params)
		suite.Require().NoError(err)

		_, err = suite.anteHandler(suite.ctx, tx, false)

		suite.Require().NotNil(err, "tx does not fail on recheck with updated params in test case: %s", tc.name)

		// reset parameters to default values
		err = suite.accountKeeper.SetParams(suite.ctx, types.DefaultParams())
		suite.Require().NoError(err)
	}

	suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sdkerrors.ErrInsufficientFee)
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
	suite.accountKeeper.SetAccount(suite.ctx, accounts[0].acc)
	// balances := suite.bankKeeper.GetAllBalances(suite.ctx, accounts[0].acc.GetAddress())
	// err = suite.bankKeeper.SendCoinsFromAccountToModule(suite.ctx, accounts[0].acc.GetAddress(), minttypes.ModuleName, balances)
	// suite.Require().NoError(err)

	// suite.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sdkerrors.ErrInsufficientFunds)
	_, err = suite.anteHandler(suite.ctx, tx, false)
	suite.Require().NotNil(err, "antehandler on recheck did not fail once feePayer no longer has sufficient funds")
}
