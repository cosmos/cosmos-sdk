package ante_test

import (
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegranttestutil "github.com/cosmos/cosmos-sdk/x/feegrant/testutil"
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

	atomPrice = sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(1).Quo(math.LegacyNewDec(100000)))
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

	dfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, nil, nil)
	antehandler := sdk.ChainAnteDecorators(dfd)
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sdkerrors.ErrInsufficientFunds)

	_, err = antehandler(s.ctx, tx, false)

	require.NotNil(t, err, "Tx did not error when fee payer had insufficient funds")

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	_, err = antehandler(s.ctx, tx, false)

	require.Nil(t, err, "Tx errored after account has been set with sufficient funds")
}

func TestDeductFees_WithFeeRecipientModule(t *testing.T) {
	tests := []struct {
		name            string
		recipientModule string
		expRecipient    string
	}{
		{
			name:            "default sends to fee_collector",
			recipientModule: "",
			expRecipient:    authtypes.FeeCollectorName,
		},
		{
			name:            "overridden module receives fees",
			recipientModule: "mint",
			expRecipient:    "mint",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := SetupTestSuite(t, false)
			s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

			accs := s.CreateTestAccounts(1)

			msg := testdata.NewTestMsg(accs[0].acc.GetAddress())
			feeAmount := testdata.NewTestFeeAmount()
			gasLimit := testdata.NewTestGasLimit()
			require.NoError(t, s.txBuilder.SetMsgs(msg))
			s.txBuilder.SetFeeAmount(feeAmount)
			s.txBuilder.SetGasLimit(gasLimit)

			privs, accNums, accSeqs := []cryptotypes.PrivKey{accs[0].priv}, []uint64{0}, []uint64{0}
			tx, err := s.CreateTestTx(s.ctx, privs, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
			require.NoError(t, err)

			dfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, nil, nil)
			if tc.recipientModule != "" {
				dfd = dfd.WithFeeRecipientModule(tc.recipientModule)
			}
			antehandler := sdk.ChainAnteDecorators(dfd)

			// Expect SendCoinsFromAccountToModule to be called with the correct recipient
			s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(
				gomock.Any(), accs[0].acc.GetAddress(), tc.expRecipient, feeAmount,
			).Return(nil)

			_, err = antehandler(s.ctx, tx, false)
			require.NoError(t, err)
		})
	}
}

// TestDeductFeeDecorator_FeegrantErrorIsBech32 runs the DeductFeeDecorator through the
// real x/feegrant keeper (not a mock) and asserts that when a fee grant does not allow
// paying the fee, the returned error contains both bech32 addresses and no raw address
// bytes / control characters. Formatting the raw []byte granter/payer with %s previously
// emitted control characters into the error and the tx log, breaking UTF-8/YAML clients.
// It fails on main (raw bytes) and passes with the fix.
func TestDeductFeeDecorator_FeegrantErrorIsBech32(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	testdata.RegisterInterfaces(encCfg.InterfaceRegistry)
	feegrant.RegisterInterfaces(encCfg.InterfaceRegistry)

	authKey := storetypes.NewKVStoreKey(authtypes.StoreKey)
	fgKey := storetypes.NewKVStoreKey(feegrant.StoreKey)
	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{authtypes.StoreKey: authKey, feegrant.StoreKey: fgKey},
		nil, nil,
	).WithBlockHeight(1)

	ak := authkeeper.NewAccountKeeper(
		encCfg.Codec, runtime.NewKVStoreService(authKey), authtypes.ProtoBaseAccount,
		map[string][]string{authtypes.FeeCollectorName: nil},
		authcodec.NewBech32Codec("cosmos"), sdk.Bech32MainPrefix, authtypes.NewModuleAddress("gov").String(),
	)
	require.NoError(t, ak.Params.Set(ctx, authtypes.DefaultParams()))
	ak.GetModuleAccount(ctx, authtypes.FeeCollectorName)

	// real feegrant keeper on a test store
	ctrl := gomock.NewController(t)
	bankKeeper := feegranttestutil.NewMockBankKeeper(ctrl)
	bankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()
	fgk := feegrantkeeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(fgKey), ak).SetBankKeeper(bankKeeper)

	// granter and payer accounts
	_, _, granter := testdata.KeyTestPubAddr()
	payerPriv, _, payer := testdata.KeyTestPubAddr()
	for i, addr := range []sdk.AccAddress{granter, payer} {
		acc := ak.NewAccountWithAddress(ctx, addr)
		require.NoError(t, acc.SetAccountNumber(uint64(i)))
		ak.SetAccount(ctx, acc)
	}

	txCfg := tx.NewTxConfig(codec.NewProtoCodec(encCfg.InterfaceRegistry), tx.DefaultSignModes)
	dfd := ante.NewDeductFeeDecorator(ak, nil, fgk, nil)
	anteHandler := sdk.ChainAnteDecorators(dfd)
	fee := sdk.NewCoins(sdk.NewInt64Coin("atom", 1000))

	buildTx := func(t *testing.T) sdk.Tx {
		acc := ak.GetAccount(ctx, payer)
		msgs := []sdk.Msg{testdata.NewTestMsg(payer)}
		txn, err := genTxWithFeeGranter(txCfg, msgs, fee, 10_000_000, ctx.ChainID(),
			[]uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, granter, payerPriv)
		require.NoError(t, err)
		return txn
	}

	assertCleanError := func(t *testing.T, err error) {
		require.Error(t, err)
		msg := err.Error()
		require.Truef(t, utf8.ValidString(msg), "error must be valid UTF-8, got %q", msg)
		for _, r := range msg {
			if r != '\n' && r != '\t' {
				require.Falsef(t, unicode.IsControl(r), "error contains control char %U: %q", r, msg)
			}
		}
		require.Contains(t, msg, granter.String())
		require.Contains(t, msg, payer.String())
		require.Contains(t, msg, "does not allow to pay fees for")
	}

	t.Run("no grant", func(t *testing.T) {
		_, err := anteHandler(ctx, buildTx(t), false)
		assertCleanError(t, err)
	})

	t.Run("grant spend limit below fee", func(t *testing.T) {
		require.NoError(t, fgk.GrantAllowance(ctx, granter, payer, &feegrant.BasicAllowance{
			SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 1)),
		}))
		_, err := anteHandler(ctx, buildTx(t), false)
		assertCleanError(t, err)
	})
}
