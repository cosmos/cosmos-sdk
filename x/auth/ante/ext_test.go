package ante_test

import (
	codectypes "github.com/pointnetwork/cosmos-point-sdk/codec/types"
	"github.com/pointnetwork/cosmos-point-sdk/testutil/testdata"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	"github.com/pointnetwork/cosmos-point-sdk/x/auth/ante"
	"github.com/pointnetwork/cosmos-point-sdk/x/auth/tx"
)

func (suite *AnteTestSuite) TestRejectExtensionOptionsDecorator() {
	suite.SetupTest(true) // setup

	testCases := []struct {
		msg   string
		allow bool
	}{
		{"allow extension", true},
		{"reject extension", false},
	}
	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
			txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()

			reod := ante.NewExtensionOptionsDecorator(func(_ *codectypes.Any) bool {
				return tc.allow
			})
			antehandler := sdk.ChainAnteDecorators(reod)

			// no extension options should not trigger an error
			theTx := txBuilder.GetTx()
			_, err := antehandler(suite.ctx, theTx, false)
			suite.Require().NoError(err)

			extOptsTxBldr, ok := txBuilder.(tx.ExtensionOptionsTxBuilder)
			if !ok {
				// if we can't set extension options, this decorator doesn't apply and we're done
				return
			}

			// set an extension option and check
			any, err := codectypes.NewAnyWithValue(testdata.NewTestMsg())
			suite.Require().NoError(err)
			extOptsTxBldr.SetExtensionOptions(any)
			theTx = txBuilder.GetTx()
			_, err = antehandler(suite.ctx, theTx, false)
			if tc.allow {
				suite.Require().NoError(err)
			} else {
				suite.Require().EqualError(err, "unknown extension options")
			}
		})
	}
}
