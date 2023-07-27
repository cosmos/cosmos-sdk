package ante_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

func (suite *AnteTestSuite) TestDeductFeesNoDelegation() {
	suite.SetupTest(false)
	// setup
	app, ctx := suite.app, suite.ctx

	protoTxCfg := tx.NewTxConfig(codec.NewProtoCodec(app.InterfaceRegistry()), tx.DefaultSignModes)

	// this just tests our handler
	dfd := ante.NewDeductFeeDecorator(app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper)
	feeAnteHandler := sdk.ChainAnteDecorators(dfd)

	// this tests the whole stack
	anteHandlerStack := suite.anteHandler

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()
	priv3, _, addr3 := testdata.KeyTestPubAddr()
	priv4, _, addr4 := testdata.KeyTestPubAddr()
	priv5, _, addr5 := testdata.KeyTestPubAddr()

	// Set addr1 with insufficient funds
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, []sdk.Coin{sdk.NewCoin("atom", sdk.NewInt(10))})
	suite.Require().NoError(err)

	// Set addr2 with more funds
	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr2, []sdk.Coin{sdk.NewCoin("atom", sdk.NewInt(99999))})
	suite.Require().NoError(err)

	// grant fee allowance from `addr2` to `addr3` (plenty to pay)
	err = app.FeeGrantKeeper.GrantAllowance(ctx, addr2, addr3, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 500)),
	})
	suite.Require().NoError(err)

	// grant low fee allowance (20atom), to check the tx requesting more than allowed.
	err = app.FeeGrantKeeper.GrantAllowance(ctx, addr2, addr4, &feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 20)),
	})
	suite.Require().NoError(err)

	cases := map[string]struct {
		signerKey  cryptotypes.PrivKey
		signer     sdk.AccAddress
		feeAccount sdk.AccAddress
		fee        int64
		valid      bool
	}{
		"paying with low funds": {
			signerKey: priv1,
			signer:    addr1,
			fee:       50,
			valid:     false,
		},
		"paying with good funds": {
			signerKey: priv2,
			signer:    addr2,
			fee:       50,
			valid:     true,
		},
		"paying with no account": {
			signerKey: priv3,
			signer:    addr3,
			fee:       1,
			valid:     false,
		},
		"no fee with real account": {
			signerKey: priv1,
			signer:    addr1,
			fee:       0,
			valid:     true,
		},
		"no fee with no account": {
			signerKey: priv5,
			signer:    addr5,
			fee:       0,
			valid:     false,
		},
		"valid fee grant without account": {
			signerKey:  priv3,
			signer:     addr3,
			feeAccount: addr2,
			fee:        50,
			valid:      true,
		},
		"no fee grant": {
			signerKey:  priv3,
			signer:     addr3,
			feeAccount: addr1,
			fee:        2,
			valid:      false,
		},
		"allowance smaller than requested fee": {
			signerKey:  priv4,
			signer:     addr4,
			feeAccount: addr2,
			fee:        50,
			valid:      false,
		},
		"granter cannot cover allowed fee grant": {
			signerKey:  priv4,
			signer:     addr4,
			feeAccount: addr1,
			fee:        50,
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
			_, err = feeAnteHandler(ctx, tx, false) // tests only feegrant ante
			if tc.valid {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

			_, err = anteHandlerStack(ctx, tx, false) // tests while stack
			if tc.valid {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// don't consume any gas
func SigGasNoConsumer(meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params authtypes.Params) error {
	return nil
}

func genTxWithFeeGranter(gen client.TxConfig, msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string, accNums,
	accSeqs []uint64, feeGranter sdk.AccAddress, priv ...cryptotypes.PrivKey,
) (sdk.Tx, error) {
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
		signerData := authsign.SignerData{
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
