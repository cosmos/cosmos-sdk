package middleware_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// testAccount represents an account used in the tests in x/auth/middleware.
type testAccount struct {
	acc  authtypes.AccountI
	priv cryptotypes.PrivKey
}

// MWTestSuite is a test suite to be used with middleware tests.
type MWTestSuite struct {
	suite.Suite

	app       *simapp.SimApp
	clientCtx client.Context
	txHandler txtypes.Handler
}

// returns context and app with params set on account keeper
func createTestApp(t *testing.T, isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(t, isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{}).WithBlockGasMeter(sdk.NewInfiniteGasMeter())
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	return app, ctx
}

// setupTest setups a new test, with new app and context.
func (s *MWTestSuite) SetupTest(isCheckTx bool) sdk.Context {
	var ctx sdk.Context
	s.app, ctx = createTestApp(s.T(), isCheckTx)
	ctx = ctx.WithBlockHeight(1)

	// Set up TxConfig.
	encodingConfig := simapp.MakeTestEncodingConfig()
	// We're using TestMsg encoding in some tests, so register it here.
	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	s.clientCtx = client.Context{}.
		WithTxConfig(encodingConfig.TxConfig).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithCodec(codec.NewAminoCodec(encodingConfig.Amino))

	// We don't use simapp's own txHandler. For more flexibility (i.e. around
	// using testdata), we create own own txHandler for this test suite.
	msr := middleware.NewMsgServiceRouter(encodingConfig.InterfaceRegistry)
	testdata.RegisterMsgServer(msr, testdata.MsgServerImpl{})
	legacyRouter := middleware.NewLegacyRouter()
	legacyRouter.AddRoute(sdk.NewRoute((&testdata.TestMsg{}).Route(), func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) { return &sdk.Result{}, nil }))
	txHandler, err := middleware.NewDefaultTxHandler(middleware.TxHandlerOptions{
		Debug:            s.app.Trace(),
		MsgServiceRouter: msr,
		LegacyRouter:     legacyRouter,
		AccountKeeper:    s.app.AccountKeeper,
		BankKeeper:       s.app.BankKeeper,
		FeegrantKeeper:   s.app.FeeGrantKeeper,
		SignModeHandler:  encodingConfig.TxConfig.SignModeHandler(),
		SigGasConsumer:   middleware.DefaultSigVerificationGasConsumer,
	})
	s.Require().NoError(err)
	s.txHandler = txHandler

	return ctx
}

// createTestAccounts creates `numAccs` accounts, and return all relevant
// information about them including their private keys.
func (s *MWTestSuite) createTestAccounts(ctx sdk.Context, numAccs int) []testAccount {
	var accounts []testAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		err := acc.SetAccountNumber(uint64(i))
		s.Require().NoError(err)
		s.app.AccountKeeper.SetAccount(ctx, acc)
		someCoins := sdk.Coins{
			sdk.NewInt64Coin("atom", 10000000),
		}
		err = s.app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, someCoins)
		s.Require().NoError(err)

		err = s.app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, someCoins)
		s.Require().NoError(err)

		accounts = append(accounts, testAccount{acc, priv})
	}

	return accounts
}

// createTestTx is a helper function to create a tx given multiple inputs.
func (s *MWTestSuite) createTestTx(txBuilder client.TxBuilder, privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, []byte, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  s.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			Address:       sdk.AccAddress(priv.PubKey().Address()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			SignerIndex:   i,
		}
		sigV2, err := tx.SignWithPrivKey(
			s.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			txBuilder, priv, s.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, nil, err
	}

	txBytes, err := s.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, nil, err
	}

	return txBuilder.GetTx(), txBytes, nil
}

func (s *MWTestSuite) runTestCase(ctx sdk.Context, txBuilder client.TxBuilder, privs []cryptotypes.PrivKey, msgs []sdk.Msg, feeAmount sdk.Coins, gasLimit uint64, accNums, accSeqs []uint64, chainID string, tc TestCase) {
	s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
		s.Require().NoError(txBuilder.SetMsgs(msgs...))
		txBuilder.SetFeeAmount(feeAmount)
		txBuilder.SetGasLimit(gasLimit)

		// Theoretically speaking, middleware unit tests should only test
		// middlewares, but here we sometimes also test the tx creation
		// process.
		tx, _, txErr := s.createTestTx(txBuilder, privs, accNums, accSeqs, chainID)
		newCtx, txHandlerErr := s.txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx, types.RequestDeliverTx{})

		if tc.expPass {
			s.Require().NoError(txErr)
			s.Require().NoError(txHandlerErr)
			s.Require().NotNil(newCtx)
		} else {
			switch {
			case txErr != nil:
				s.Require().Error(txErr)
				s.Require().True(errors.Is(txErr, tc.expErr))

			case txHandlerErr != nil:
				s.Require().Error(txHandlerErr)
				s.Require().True(errors.Is(txHandlerErr, tc.expErr))

			default:
				s.Fail("expected one of txErr,txHandlerErr to be an error")
			}
		}
	})
}

// TestCase represents a test case used in test tables.
type TestCase struct {
	desc     string
	malleate func()
	simulate bool
	expPass  bool
	expErr   error
}

func TestMWTestSuite(t *testing.T) {
	suite.Run(t, new(MWTestSuite))
}
