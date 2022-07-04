package ante_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	_ "github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	antetestutil "github.com/cosmos/cosmos-sdk/x/auth/ante/testutil"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	txtestutil "github.com/cosmos/cosmos-sdk/x/auth/tx/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// TestAccount represents an account used in the tests in x/auth/ante.
type TestAccount struct {
	acc  sdk.AccountI
	priv cryptotypes.PrivKey
}

// AnteTestSuite is a test suite to be used with ante handler tests.
type AnteTestSuite struct {
	env             appmodule.Environment
	anteHandler     sdk.AnteHandler
	ctx             sdk.Context
	clientCtx       client.Context
	txBuilder       client.TxBuilder
	accountKeeper   keeper.AccountKeeper
	bankKeeper      *authtestutil.MockBankKeeper
	consensusKeeper *antetestutil.MockConsensusKeeper
	acctsModKeeper  *authtestutil.MockAccountsModKeeper
	txBankKeeper    *txtestutil.MockBankKeeper
	feeGrantKeeper  *antetestutil.MockFeegrantKeeper
	encCfg          moduletestutil.TestEncodingConfig
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func SetupTestSuite(t *testing.T, isCheckTx bool) *AnteTestSuite {
	t.Helper()
	suite := &AnteTestSuite{}
	// gomock initializations
	ctrl := gomock.NewController(t)
	suite.bankKeeper = authtestutil.NewMockBankKeeper(ctrl)
	suite.txBankKeeper = txtestutil.NewMockBankKeeper(ctrl)
	suite.feeGrantKeeper = antetestutil.NewMockFeegrantKeeper(ctrl)
	suite.acctsModKeeper = authtestutil.NewMockAccountsModKeeper(ctrl)

	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithIsCheckTx(isCheckTx).WithBlockHeight(1).WithHeaderInfo(header.Info{Height: 1, ChainID: testCtx.Ctx.ChainID()})
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{})

	accNum := uint64(0)
	suite.acctsModKeeper.EXPECT().NextAccountNumber(gomock.Any()).AnyTimes().DoAndReturn(func(ctx context.Context) (uint64, error) {
		currNum := accNum
		accNum++
		return currNum, nil
	})
	maccPerms := map[string][]string{
		"fee_collector":          nil,
		"mint":                   {"minter"},
		"bonded_tokens_pool":     {"burner", "staking"},
		"not_bonded_tokens_pool": {"burner", "staking"},
		"multiPerm":              {"burner", "minter", "staking"},
		"random":                 {"random"},
	}

	msgRouter := baseapp.NewMsgServiceRouter()
	grpcQueryRouter := baseapp.NewGRPCQueryRouter()
	grpcQueryRouter.SetInterfaceRegistry(suite.encCfg.InterfaceRegistry)

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (s *AnteTestSuite) SetupTest(isCheckTx bool) {
	s.app, s.ctx = createTestApp(s.T(), isCheckTx)
	s.ctx = s.ctx.WithBlockHeight(1)

	// We're using TestMsg encoding in some tests, so register it here.
	suite.encCfg.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg")
	testdata.RegisterInterfaces(suite.encCfg.InterfaceRegistry)

	s.clientCtx = client.Context{}.
		WithTxConfig(encodingConfig.TxConfig)

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   s.app.AccountKeeper,
			BankKeeper:      s.app.BankKeeper,
			FeegrantKeeper:  s.app.FeeGrantKeeper,
			SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			Environment:     suite.env,
		},
	)

	s.Require().NoError(err)
	s.anteHandler = anteHandler
}

// CreateTestAccounts creates `numAccs` accounts, and return all relevant
// information about them including their private keys.
func (s *AnteTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
		err := acc.SetAccountNumber(uint64(i))
		s.Require().NoError(err)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		someCoins := sdk.Coins{
			sdk.NewInt64Coin("atom", 10000000),
		}
		err = s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, someCoins)
		s.Require().NoError(err)

		err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, addr, someCoins)
		s.Require().NoError(err)

		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

// TestCase represents a test case used in test tables.
type TestCase struct {
	desc     string
	malleate func(*AnteTestSuite) TestCaseArgs
	simulate bool
	expPass  bool
	expErr   error
}

type TestCaseArgs struct {
	chainID   string
	accNums   []uint64
	accSeqs   []uint64
	feeAmount sdk.Coins
	gasLimit  uint64
	msgs      []sdk.Msg
	privs     []cryptotypes.PrivKey
}

func (t TestCaseArgs) WithAccountsInfo(accs []TestAccount) TestCaseArgs {
	newT := t
	for _, acc := range accs {
		newT.accNums = append(newT.accNums, acc.acc.GetAccountNumber())
		newT.accSeqs = append(newT.accSeqs, acc.acc.GetSequence())
		newT.privs = append(newT.privs, acc.priv)
	}
	return newT
}

// DeliverMsgs constructs a tx and runs it through the ante handler. This is used to set the context for a test case, for
// example to test for replay protection.
func (suite *AnteTestSuite) DeliverMsgs(t *testing.T, privs []cryptotypes.PrivKey, msgs []sdk.Msg, feeAmount sdk.Coins, gasLimit uint64, accNums, accSeqs []uint64, chainID string, simulate bool) (sdk.Context, error) {
	t.Helper()
	require.NoError(t, suite.txBuilder.SetMsgs(msgs...))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	tx, txErr := suite.CreateTestTx(suite.ctx, privs, accNums, accSeqs, chainID, signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, txErr)
	txBytes, err := suite.clientCtx.TxConfig.TxEncoder()(tx)
	bytesCtx := suite.ctx.WithTxBytes(txBytes)
	require.NoError(t, err)
	return suite.anteHandler(bytesCtx, tx, simulate)
}

func (suite *AnteTestSuite) RunTestCase(t *testing.T, tc TestCase, args TestCaseArgs) {
	t.Helper()
	require.NoError(t, suite.txBuilder.SetMsgs(args.msgs...))
	suite.txBuilder.SetFeeAmount(args.feeAmount)
	suite.txBuilder.SetGasLimit(args.gasLimit)

	// Theoretically speaking, ante handler unit tests should only test
	// ante handlers, but here we sometimes also test the tx creation
	// process.
	tx, txErr := suite.CreateTestTx(suite.ctx, args.privs, args.accNums, args.accSeqs, args.chainID, signing.SignMode_SIGN_MODE_DIRECT)
	txBytes, err := suite.clientCtx.TxConfig.TxEncoder()(tx)
	require.NoError(t, err)
	bytesCtx := suite.ctx.WithTxBytes(txBytes)
	newCtx, anteErr := suite.anteHandler(bytesCtx, tx, tc.simulate)

	if tc.expPass {
		require.NoError(t, txErr)
		require.NoError(t, anteErr)
		require.NotNil(t, newCtx)

		suite.ctx = newCtx
	} else {
		switch {
		case txErr != nil:
			require.Error(t, txErr)
			require.ErrorIs(t, txErr, tc.expErr)

		case anteErr != nil:
			require.Error(t, anteErr)
			require.ErrorIs(t, anteErr, tc.expErr)

		default:
			t.Fatal("expected one of txErr, anteErr to be an error")
		}
	}
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (s *AnteTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
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
	err := s.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			Address:       sdk.AccAddress(priv.PubKey().Address()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        priv.PubKey(),
		}
		sigV2, err := tx.SignWithPrivKey(
			s.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			s.txBuilder, priv, s.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = s.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return s.txBuilder.GetTx(), nil
}

// TestCase represents a test case used in test tables.
type TestCase struct {
	desc     string
	malleate func()
	simulate bool
	expPass  bool
	expErr   error
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (s *AnteTestSuite) RunTestCase(privs []cryptotypes.PrivKey, msgs []sdk.Msg, feeAmount sdk.Coins, gasLimit uint64, accNums, accSeqs []uint64, chainID string, tc TestCase) {
	s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
		s.Require().NoError(s.txBuilder.SetMsgs(msgs...))
		s.txBuilder.SetFeeAmount(feeAmount)
		s.txBuilder.SetGasLimit(gasLimit)

		// Theoretically speaking, ante handler unit tests should only test
		// ante handlers, but here we sometimes also test the tx creation
		// process.
		tx, txErr := s.CreateTestTx(privs, accNums, accSeqs, chainID)
		newCtx, anteErr := s.anteHandler(s.ctx, tx, tc.simulate)

		if tc.expPass {
			s.Require().NoError(txErr)
			s.Require().NoError(anteErr)
			s.Require().NotNil(newCtx)

			s.ctx = newCtx
		} else {
			switch {
			case txErr != nil:
				s.Require().Error(txErr)
				s.Require().True(errors.Is(txErr, tc.expErr))

			case anteErr != nil:
				s.Require().Error(anteErr)
				s.Require().True(errors.Is(anteErr, tc.expErr))

			default:
				s.Fail("expected one of txErr,anteErr to be an error")
			}
		}
	})
}
