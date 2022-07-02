package ante_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/testutil"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
)

// TestAccount represents an account used in the tests in x/auth/ante.
type TestAccount struct {
	acc  types.AccountI
	priv cryptotypes.PrivKey
}

// AnteTestSuite is a test suite to be used with ante handler tests.
type AnteTestSuite struct {
	suite.Suite

	interfaceRegistry codectypes.InterfaceRegistry
	anteHandler       sdk.AnteHandler
	ctx               sdk.Context
	clientCtx         client.Context
	txBuilder         client.TxBuilder
	accountKeeper     keeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	feeGrantKeeper    feegrantkeeper.Keeper
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (suite *AnteTestSuite) SetupTest(isCheckTx bool) {
	var (
		txConfig    client.TxConfig
		legacyAmino *codec.LegacyAmino
	)

	app, err := simtestutil.Setup(
		testutil.AppConfig,
		&suite.accountKeeper,
		&suite.bankKeeper,
		&suite.feeGrantKeeper,
		&suite.interfaceRegistry,
		&txConfig,
		&legacyAmino,
	)
	suite.Require().NoError(err)

	suite.ctx = app.BaseApp.NewContext(isCheckTx, tmproto.Header{}).WithBlockHeight(1)
	suite.accountKeeper.SetParams(suite.ctx, authtypes.DefaultParams())

	// We're using TestMsg encoding in some tests, so register it here.
	legacyAmino.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(suite.interfaceRegistry)

	suite.clientCtx = client.Context{}.
		WithTxConfig(txConfig)

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   suite.accountKeeper,
			BankKeeper:      suite.bankKeeper,
			FeegrantKeeper:  suite.feeGrantKeeper,
			SignModeHandler: txConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
	)

	suite.Require().NoError(err)
	suite.anteHandler = anteHandler
}

// CreateTestAccounts creates `numAccs` accounts, and return all relevant
// information about them including their private keys.
func (suite *AnteTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
		err := acc.SetAccountNumber(uint64(i))
		suite.Require().NoError(err)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		someCoins := sdk.Coins{
			sdk.NewInt64Coin("atom", 10000000),
		}
		err = suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, someCoins)
		suite.Require().NoError(err)

		err = suite.bankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr, someCoins)
		suite.Require().NoError(err)

		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (suite *AnteTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			suite.txBuilder, priv, suite.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return suite.txBuilder.GetTx(), nil
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
func (suite *AnteTestSuite) RunTestCase(privs []cryptotypes.PrivKey, msgs []sdk.Msg, feeAmount sdk.Coins, gasLimit uint64, accNums, accSeqs []uint64, chainID string, tc TestCase) {
	suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
		suite.Require().NoError(suite.txBuilder.SetMsgs(msgs...))
		suite.txBuilder.SetFeeAmount(feeAmount)
		suite.txBuilder.SetGasLimit(gasLimit)

		// Theoretically speaking, ante handler unit tests should only test
		// ante handlers, but here we sometimes also test the tx creation
		// process.
		tx, txErr := suite.CreateTestTx(privs, accNums, accSeqs, chainID)
		newCtx, anteErr := suite.anteHandler(suite.ctx, tx, tc.simulate)

		if tc.expPass {
			suite.Require().NoError(txErr)
			suite.Require().NoError(anteErr)
			suite.Require().NotNil(newCtx)

			suite.ctx = newCtx
		} else {
			switch {
			case txErr != nil:
				suite.Require().Error(txErr)
				suite.Require().True(errors.Is(txErr, tc.expErr))

			case anteErr != nil:
				suite.Require().Error(anteErr)
				suite.Require().True(errors.Is(anteErr, tc.expErr))

			default:
				suite.Fail("expected one of txErr,anteErr to be an error")
			}
		}
	})
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}
