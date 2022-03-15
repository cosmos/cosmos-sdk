package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/tieredfee/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

var testCoins = sdk.Coins{sdk.NewInt64Coin("atom", 10000000)}

// testAccount represents an account used in the tests in x/auth/middleware.
type testAccount struct {
	acc    authtypes.AccountI
	priv   cryptotypes.PrivKey
	accNum uint64
}

// createTestAccounts creates `numAccs` accounts, and return all relevant
// information about them including their private keys.
func (suite *IntegrationTestSuite) createTestAccounts(numAccs int, coins sdk.Coins) []testAccount {
	var accounts []testAccount
	ctx := suite.ctx

	for i := 0; i < numAccs; i++ {
		priv, _, addr := testdata.KeyTestPubAddr()
		acc := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		accNum := uint64(i)
		err := acc.SetAccountNumber(accNum)
		suite.Require().NoError(err)
		suite.app.AccountKeeper.SetAccount(ctx, acc)
		err = suite.app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
		suite.Require().NoError(err)

		err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins)
		suite.Require().NoError(err)

		accounts = append(accounts, testAccount{acc, priv, accNum})
	}

	return accounts
}

// createTestTx is a helper function to create a tx given multiple inputs.
func (suite *IntegrationTestSuite) createTestTx(txBuilder client.TxBuilder, privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, []byte, error) {
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
			PubKey:        priv.PubKey(),
		}
		sigV2, err := tx.SignWithPrivKey(
			suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			txBuilder, priv, suite.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, nil, err
	}

	txBytes, err := suite.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, nil, err
	}

	return txBuilder.GetTx(), txBytes, nil
}

func (suite *IntegrationTestSuite) TestMiddleware() {
	suite.SetupTest()
	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
	accounts := suite.createTestAccounts(3, testCoins)
	privs := []cryptotypes.PrivKey{accounts[0].priv, accounts[1].priv, accounts[2].priv}
	accSeqs := []uint64{0, 0, 0}
	accNums := []uint64{0, 1, 2}
	msgs := []sdk.Msg{
		banktypes.NewMsgSend(accounts[0].acc.GetAddress(), accounts[1].acc.GetAddress(), sdk.Coins{sdk.NewInt64Coin("atom", 1)}),
		banktypes.NewMsgSend(accounts[1].acc.GetAddress(), accounts[1].acc.GetAddress(), sdk.Coins{sdk.NewInt64Coin("atom", 1)}),
		banktypes.NewMsgSend(accounts[2].acc.GetAddress(), accounts[1].acc.GetAddress(), sdk.Coins{sdk.NewInt64Coin("atom", 1)}),
	}
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		expErr   error
	}{
		{"default tier without extension option", func() {}, true, nil},
		{"specify tier with extension option", func() {
			builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
			suite.Require().True(ok)
			option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionTieredTx{
				FeeTier: 0,
			})
			suite.Require().NoError(err)
			builder.SetExtensionOptions(option)
		}, true, nil},
		{"specify invalid tier", func() {
			builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
			suite.Require().True(ok)
			option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionTieredTx{
				FeeTier: 1,
			})
			suite.Require().NoError(err)
			builder.SetExtensionOptions(option)
		}, false, nil},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			suite.Require().NoError(txBuilder.SetMsgs(msgs...))
			txBuilder.SetFeeAmount(feeAmount)
			txBuilder.SetGasLimit(gasLimit)
			_, txBytes, txErr := suite.createTestTx(txBuilder, privs, accNums, accSeqs, suite.ctx.ChainID())
			rsp := suite.app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
			if tc.expPass {
				suite.Require().NoError(txErr)
				suite.Require().Zero(rsp.Code, rsp.Log)
				for i, n := range accSeqs {
					accSeqs[i] = n + 1
				}
			} else {
				if txErr != nil {
					suite.Require().NoError(txErr)
				} else {
					suite.Require().Positive(rsp.Code)
				}
			}
		})
	}
}
