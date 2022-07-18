package ante_test

import (
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func (s *AnteTestSuite) TestDeductFeeDecorator_ZeroGas() {
	s.SetupTest(true) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	mfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, s.feeGrantKeeper, nil)
	antehandler := sdk.ChainAnteDecorators(mfd)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	// coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(300)))
	// testutil.FundAccount(s.bankKeeper, s.ctx, addr1, coins)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	s.Require().NoError(s.txBuilder.SetMsgs(msg))

	// set zero gas
	s.txBuilder.SetGasLimit(0)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	_, err = antehandler(s.ctx, tx, false)
	s.Require().Error(err)

	// zero gas is accepted in simulation mode
	_, err = antehandler(s.ctx, tx, true)
	s.Require().NoError(err)
}

func TestEnsureMempoolFees(t *testing.T) {
	s := SetupTestSuite(t, true) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	mfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, s.feeGrantKeeper, nil)
	antehandler := sdk.ChainAnteDecorators(mfd)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	s.accountKeeper.SetAccount(s.ctx, s.accountKeeper.NewAccountWithAddress(s.ctx, addr1))
	// coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(300)))
	// testutil.FundAccount(s.bankKeeper, s.ctx, addr1, coins)

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	require.NoError(t, s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), addr1, authtypes.FeeCollectorName, feeAmount).Return(nil)
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), addr1, authtypes.FeeCollectorName, feeAmount).Return(nil)
	s.accountKeeper.SetAccount(s.ctx, s.accountKeeper.NewAccountWithAddress(s.ctx, addr1))

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	require.NoError(t, err)

	// Set high gas price so standard test fee fails
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(200).Quo(sdk.NewDec(100000)))
	highGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	require.NotNil(t, err, "Decorator should have errored on too low fee for local gasPrice")

	// antehandler should not error since we do not check minGasPrice in simulation mode
	cacheCtx, _ := s.ctx.CacheContext()
	_, err = antehandler(cacheCtx, tx, true)
	s.Require().Nil(err, "Decorator should not have errored in simulation mode")

	// Set IsCheckTx to false
	s.ctx = s.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = antehandler(s.ctx, tx, false)
	require.Nil(t, err, "MempoolFeeDecorator returned error in DeliverTx")

	// Set IsCheckTx back to true for testing sufficient mempool fee
	s.ctx = s.ctx.WithIsCheckTx(true)

	atomPrice = sdk.NewDecCoinFromDec("atom", sdk.NewDec(0).Quo(sdk.NewDec(100000)))
	lowGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(lowGasPrice)

	newCtx, err := antehandler(s.ctx, tx, false)
	require.Nil(t, err, "Decorator should not have errored on fee higher than local gasPrice")
	// Priority is the smallest amount in any denom. Since we have only 1 fee
	// of 150atom, the priority here is 150.
	require.Equal(t, feeAmount.AmountOf("atom").Int64(), newCtx.Priority())
}

func (s *AnteTestSuite) TestDeductFees() {
	s.SetupTest(false) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set account with insufficient funds
	acc := s.accountKeeper.NewAccountWithAddress(s.ctx, addr1)
	s.accountKeeper.SetAccount(s.ctx, acc)
	// coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(10)))
	// err = testutil.FundAccount(s.bankKeeper, s.ctx, addr1, coins)
	s.Require().NoError(err)

	dfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, nil, nil)
	antehandler := sdk.ChainAnteDecorators(dfd)
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sdkerrors.ErrInsufficientFunds)

	_, err = antehandler(s.ctx, tx, false)

	s.Require().NotNil(err, "Tx did not error when fee payer had insufficient funds")

	// Set account with sufficient funds
	s.accountKeeper.SetAccount(s.ctx, acc)
	// err = testutil.FundAccount(s.bankKeeper, s.ctx, addr1, sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200))))
	// s.Require().NoError(err)

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	_, err = antehandler(s.ctx, tx, false)

	s.Require().Nil(err, "Tx errored after account has been set with sufficient funds")
}
