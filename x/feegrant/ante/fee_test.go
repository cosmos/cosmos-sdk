package ante_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authnante "github.com/cosmos/cosmos-sdk/x/authn/ante"
	authnsign "github.com/cosmos/cosmos-sdk/x/authn/signing"
	"github.com/cosmos/cosmos-sdk/x/authn/tx"
	authntypes "github.com/cosmos/cosmos-sdk/x/authn/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/ante"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

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

	// Set up TxConfig.
	encodingConfig := simapp.MakeTestEncodingConfig()
	// We're using TestMsg encoding in some tests, so register it here.
	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	suite.clientCtx = client.Context{}.
		WithTxConfig(encodingConfig.TxConfig)

	suite.anteHandler = ante.NewAnteHandler(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.FeeGrantKeeper, authnante.DefaultSigVerificationGasConsumer, encodingConfig.TxConfig.SignModeHandler())
}

func (suite *AnteTestSuite) TestDeductFeesNoDelegation() {
	suite.SetupTest(true)
	// setup
	app, ctx := suite.app, suite.ctx

	protoTxCfg := tx.NewTxConfig(codec.NewProtoCodec(app.InterfaceRegistry()), tx.DefaultSignModes)

	// this just tests our handler
	dfd := ante.NewDeductGrantedFeeDecorator(app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper)
	ourAnteHandler := sdk.ChainAnteDecorators(dfd)

	// this tests the whole stack
	anteHandlerStack := suite.anteHandler

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	priv3, _, addr3 := testdata.KeyTestPubAddr()
	priv4, _, addr4 := testdata.KeyTestPubAddr()
	priv5, _, addr5 := testdata.KeyTestPubAddr()

	// Set addr1 with insufficient funds
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	app.BankKeeper.SetBalances(ctx, addr1, []sdk.Coin{sdk.NewCoin("atom", sdk.NewInt(10))})

	// Set addr2 with more funds
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	app.AccountKeeper.SetAccount(ctx, acc2)
	app.BankKeeper.SetBalances(ctx, addr2, []sdk.Coin{sdk.NewCoin("atom", sdk.NewInt(99999))})

	// grant fee allowance from `addr2` to `addr3` (plenty to pay)
	err := app.FeeGrantKeeper.GrantFeeAllowance(ctx, addr2, addr3, &types.BasicFeeAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 500)),
	})
	suite.Require().NoError(err)

	// grant low fee allowance (20atom), to check the tx requesting more than allowed.
	err = app.FeeGrantKeeper.GrantFeeAllowance(ctx, addr2, addr4, &types.BasicFeeAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 20)),
	})
	suite.Require().NoError(err)

	cases := map[string]struct {
		signerKey     cryptotypes.PrivKey
		signer        sdk.AccAddress
		feeAccount    sdk.AccAddress
		feeAccountKey cryptotypes.PrivKey
		handler       sdk.AnteHandler
		fee           int64
		valid         bool
	}{
		"paying with low funds (only ours)": {
			signerKey: priv1,
			signer:    addr1,
			fee:       50,
			handler:   ourAnteHandler,
			valid:     false,
		},
		"paying with good funds (only ours)": {
			signerKey: priv2,
			signer:    addr2,
			fee:       50,
			handler:   ourAnteHandler,
			valid:     true,
		},
		"paying with no account (only ours)": {
			signerKey: priv3,
			signer:    addr3,
			fee:       1,
			handler:   ourAnteHandler,
			valid:     false,
		},
		"no fee with real account (only ours)": {
			signerKey: priv1,
			signer:    addr1,
			fee:       0,
			handler:   ourAnteHandler,
			valid:     true,
		},
		"no fee with no account (only ours)": {
			signerKey: priv5,
			signer:    addr5,
			fee:       0,
			handler:   ourAnteHandler,
			valid:     false,
		},
		"valid fee grant without account (only ours)": {
			signerKey:  priv3,
			signer:     addr3,
			feeAccount: addr2,
			fee:        50,
			handler:    ourAnteHandler,
			valid:      true,
		},
		"no fee grant (only ours)": {
			signerKey:  priv3,
			signer:     addr3,
			feeAccount: addr1,
			fee:        2,
			handler:    ourAnteHandler,
			valid:      false,
		},
		"allowance smaller than requested fee (only ours)": {
			signerKey:  priv4,
			signer:     addr4,
			feeAccount: addr2,
			fee:        50,
			handler:    ourAnteHandler,
			valid:      false,
		},
		"granter cannot cover allowed fee grant (only ours)": {
			signerKey:  priv4,
			signer:     addr4,
			feeAccount: addr1,
			fee:        50,
			handler:    ourAnteHandler,
			valid:      false,
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		suite.T().Run(name, func(t *testing.T) {
			fee := sdk.NewCoins(sdk.NewInt64Coin("atom", tc.fee))
			msgs := []sdk.Msg{testdata.NewTestMsg(tc.signer)}

			acc := app.AccountKeeper.GetAccount(ctx, tc.signer)
			privs, accNums, seqs := []cryptotypes.PrivKey{tc.signerKey}, []uint64{0}, []uint64{0}
			if acc != nil {
				accNums, seqs = []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}
			}

			tx, err := genTxWithFeeGranter(protoTxCfg, msgs, fee, helpers.DefaultGenTxGas, ctx.ChainID(), accNums, seqs, tc.feeAccount, privs...)
			suite.Require().NoError(err)
			_, err = ourAnteHandler(ctx, tx, false)
			if tc.valid {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

			_, err = anteHandlerStack(ctx, tx, false)
			if tc.valid {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.AccountKeeper.SetParams(ctx, authntypes.DefaultParams())

	return app, ctx
}

// don't consume any gas
func SigGasNoConsumer(meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params authntypes.Params) error {
	return nil
}

func genTxWithFeeGranter(gen client.TxConfig, msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string, accNums,
	accSeqs []uint64, feeGranter sdk.AccAddress, priv ...cryptotypes.PrivKey) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(priv))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	signMode := gen.SignModeHandler().DefaultMode()

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range priv {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	tx := gen.NewTxBuilder()
	err := tx.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = tx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	tx.SetMemo(memo)
	tx.SetFeeAmount(feeAmt)
	tx.SetGasLimit(gas)
	tx.SetFeeGranter(feeGranter)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := authnsign.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		signBytes, err := gen.SignModeHandler().GetSignBytes(signMode, signerData, tx.GetTx())
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = tx.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return tx.GetTx(), nil
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}
