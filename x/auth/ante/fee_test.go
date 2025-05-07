package ante_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestDeductFeeDecorator_ZeroGas(t *testing.T) {
	s := SetupTestSuite(t, true)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	mfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, s.feeGrantKeeper, nil)
	antehandler := sdk.ChainAnteDecorators(mfd)

	// keys and addresses
	accs := s.CreateTestAccounts(1)

	// msg and signatures
	msg := testdata.NewTestMsg(accs[0].acc.GetAddress())
	require.NoError(t, s.txBuilder.SetMsgs(msg))

	// set zero gas
	s.txBuilder.SetGasLimit(0)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{accs[0].priv}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(s.ctx, privs, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	// Get the signer's address from the transaction
	signerAddr := tx.GetSigners()[0]
	fmt.Printf("Signer address: %s\n", signerAddr.String())

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	_, err = antehandler(s.ctx, tx, false)
	require.Error(t, err)

	// zero gas is accepted in simulation mode
	_, err = antehandler(s.ctx, tx, true)
	require.NoError(t, err)
}

func TestEnsureMempoolFees(t *testing.T) {
	s := SetupTestSuite(t, true) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	mfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, s.feeGrantKeeper, nil)
	antehandler := sdk.ChainAnteDecorators(mfd)

	// keys and addresses
	accs := s.CreateTestAccounts(1)

	// msg and signatures
	msg := testdata.NewTestMsg(accs[0].acc.GetAddress())
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := uint64(15)
	require.NoError(t, s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accs[0].acc.GetAddress(), authtypes.FeeCollectorName, feeAmount).Return(nil).Times(3)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{accs[0].priv}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(s.ctx, privs, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	// Get the signer's address from the transaction
	signerAddr := tx.GetSigners()[0]
	fmt.Printf("Signer address: %s\n", signerAddr.String())

	// Set high gas price so standard test fee fails
	atomPrice := sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(20))
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
	require.Nil(t, err, "Decorator should not have errored in simulation mode")

	// Set IsCheckTx to false
	s.ctx = s.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = antehandler(s.ctx, tx, false)
	require.Nil(t, err, "MempoolFeeDecorator returned error in DeliverTx")

	// Set IsCheckTx back to true for testing sufficient mempool fee
	s.ctx = s.ctx.WithIsCheckTx(true)

	atomPrice = sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(0).Quo(math.LegacyNewDec(100000)))
	lowGasPrice := []sdk.DecCoin{atomPrice}
	s.ctx = s.ctx.WithMinGasPrices(lowGasPrice)

	newCtx, err := antehandler(s.ctx, tx, false)
	require.Nil(t, err, "Decorator should not have errored on fee higher than local gasPrice")
	// Priority is the smallest gas price amount in any denom. Since we have only 1 gas price
	// of 10atom, the priority here is 10.
	require.Equal(t, int64(10), newCtx.Priority())
}

func TestDeductFees(t *testing.T) {
	s := SetupTestSuite(t, false)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	accs := s.CreateTestAccounts(1)

	// msg and signatures
	msg := testdata.NewTestMsg(accs[0].acc.GetAddress())
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	require.NoError(t, s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{accs[0].priv}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(s.ctx, privs, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	// Get the signer's address from the transaction
	signerAddr := tx.GetSigners()[0]
	fmt.Printf("Signer address: %s\n", signerAddr.String())

	dfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, nil, nil)
	antehandler := sdk.ChainAnteDecorators(dfd)
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sdkerrors.ErrInsufficientFunds)

	_, err = antehandler(s.ctx, tx, false)

	require.NotNil(t, err, "Tx did not error when fee payer had insufficient funds")

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	_, err = antehandler(s.ctx, tx, false)

	require.Nil(t, err, "Tx errored after account has been set with sufficient funds")
}

func TestDeductFees_WithName(t *testing.T) {
	s := SetupTestSuite(t, false)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	require.NoError(t, s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	require.NoError(t, err)

	// Get the signer's address from the transaction
	signerAddr := tx.GetSigners()[0]
	fmt.Printf("Signer address: %s\n", signerAddr.String())

	s.accountKeeper.SetAccount(s.ctx, authtypes.NewBaseAccountWithAddress(signerAddr))

	// Set up initial account with coins
	coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200)))

	// Using mock bank keeper
	s.bankKeeper.EXPECT().MintCoins(s.ctx, types.FeeCollectorName, coins).Return(nil)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, types.FeeCollectorName, addr1, coins).Return(nil)

	err = s.bankKeeper.MintCoins(s.ctx, types.FeeCollectorName, coins)
	require.NoError(t, err)
	err = s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.FeeCollectorName, addr1, coins)
	require.NoError(t, err)

	altCollectorName := distrtypes.ModuleName

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), addr1, altCollectorName, gomock.Any()).Return(nil)
	dfd := ante.NewDeductFeeDecoratorWithName(s.accountKeeper, s.bankKeeper, nil, nil, altCollectorName)
	antehandler := sdk.ChainAnteDecorators(dfd)
	_, err = antehandler(s.ctx, tx, false)
	require.NoError(t, err)
}

func TestDeductFees_WithName_Table(t *testing.T) {
	testCases := []struct {
		name             string
		altCollectorName string
		expectedError    error
	}{
		{
			name:             "distribution module collector",
			altCollectorName: distrtypes.ModuleName,
			expectedError:    nil,
		},
		{
			name:             "fee collector module",
			altCollectorName: types.FeeCollectorName,
			expectedError:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := SetupTestSuite(t, false)
			s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

			// keys and addresses
			priv1, _, addr1 := testdata.KeyTestPubAddr()

			// msg and signatures
			msg := testdata.NewTestMsg(addr1)
			feeAmount := testdata.NewTestFeeAmount()
			gasLimit := testdata.NewTestGasLimit()
			require.NoError(t, s.txBuilder.SetMsgs(msg))
			s.txBuilder.SetFeeAmount(feeAmount)
			s.txBuilder.SetGasLimit(gasLimit)

			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
			tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
			require.NoError(t, err)

			// Get the signer's address from the transaction
			signerAddr := tx.GetSigners()[0]
			fmt.Printf("Signer address: %s\n", signerAddr.String())

			s.accountKeeper.SetAccount(s.ctx, authtypes.NewBaseAccountWithAddress(signerAddr))

			// Set up initial account with coins
			coins := sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(200)))

			// Using mock bank keeper
			s.bankKeeper.EXPECT().MintCoins(s.ctx, types.FeeCollectorName, coins).Return(nil)
			s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, types.FeeCollectorName, addr1, coins).Return(nil)

			err = s.bankKeeper.MintCoins(s.ctx, types.FeeCollectorName, coins)
			require.NoError(t, err)
			err = s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.FeeCollectorName, addr1, coins)
			require.NoError(t, err)

			s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), addr1, tc.altCollectorName, gomock.Any()).Return(nil)
			dfd := ante.NewDeductFeeDecoratorWithName(s.accountKeeper, s.bankKeeper, nil, nil, tc.altCollectorName)
			antehandler := sdk.ChainAnteDecorators(dfd)
			_, err = antehandler(s.ctx, tx, false)
			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
