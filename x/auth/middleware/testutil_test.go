package middleware_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
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
}

// returns context and app with params set on account keeper
func createTestApp(t *testing.T, isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(t, isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
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
		WithTxConfig(encodingConfig.TxConfig)

	return ctx
}

// CreatetestAccounts creates `numAccs` accounts, and return all relevant
// information about them including their private keys.
func (s *MWTestSuite) CreatetestAccounts(ctx sdk.Context, numAccs int) []testAccount {
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
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
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

func TestMWTestSuite(t *testing.T) {
	suite.Run(t, new(MWTestSuite))
}
