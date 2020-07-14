package ante_test

import (
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// TestAccount represents an account used in the tests below. It's simply an
// AccountI, where we have access to the PrivKey.
type TestAccount struct {
	acc  types.AccountI
	priv crypto.PrivKey
}

// AnteTestSuite is a test suite to be used with ante handler tests.
type AnteTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	anteHandler sdk.AnteHandler
	ctx         sdk.Context
	clientCtx   client.Context
	txBuilder   client.TxBuilder
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (suite *AnteTestSuite) SetupTest(isCheckTx bool) {
	suite.app, suite.ctx = createTestApp(isCheckTx)
	suite.ctx = suite.ctx.WithBlockHeight(1)
	suite.anteHandler = ante.NewAnteHandler(suite.app.AccountKeeper, suite.app.BankKeeper, *suite.app.IBCKeeper, ante.DefaultSigVerificationGasConsumer, types.LegacyAminoJSONHandler{})

	// set up the TxBuilder
	encodingConfig := simappparams.MakeEncodingConfig()
	suite.clientCtx = client.Context{}.
		WithTxGenerator(encodingConfig.TxGenerator).
		WithJSONMarshaler(encodingConfig.Marshaler)
}

// CreateTestAccounts creates `numAccs` accounts, and return all relevant
// information about them including their private keys.
func (suite *AnteTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount

	for i := 0; i < numAccs; i++ {
		priv, _, addr := types.KeyTestPubAddr()
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr)
		err := acc.SetAccountNumber(uint64(i))
		suite.Require().NoError(err)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		suite.app.BankKeeper.SetBalances(suite.ctx, addr, sdk.Coins{
			sdk.NewInt64Coin("atom", 10000000),
		})

		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (suite *AnteTestSuite) CreateTestTx(privs []crypto.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) xauthsigning.SigFeeMemoTx {
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2, err := tx.SignWithPrivKey(suite.clientCtx.TxGenerator.SignModeHandler().DefaultMode(), priv, accNums[i], accSeqs[i], chainID, suite.clientCtx.TxGenerator, suite.txBuilder)
		suite.Require().NoError(err)

		sigsV2 = append(sigsV2, sigV2)
	}
	suite.txBuilder.SetSignatures(sigsV2...)

	return suite.txBuilder.GetTx()
}
