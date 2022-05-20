package ante_test

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

func (suite *AnteTestSuite) TestRejectExtensionOptionsDecorator() {
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	reod := ante.NewRejectExtensionOptionsDecorator()
	antehandler := sdk.ChainAnteDecorators(reod)

	// no extension options should not trigger an error
	theTx := suite.txBuilder.GetTx()
	_, err := antehandler(suite.ctx, theTx, false)
	suite.Require().NoError(err)

	extOptsTxBldr, ok := suite.txBuilder.(tx.ExtensionOptionsTxBuilder)
	if !ok {
		// if we can't set extension options, this decorator doesn't apply and we're done
		return
	}

	// setting any extension option should cause an error
	any, err := types.NewAnyWithValue(testdata.NewTestMsg())
	suite.Require().NoError(err)
	extOptsTxBldr.SetExtensionOptions(any)
	theTx = suite.txBuilder.GetTx()
	_, err = antehandler(suite.ctx, theTx, false)
	suite.Require().EqualError(err, "unknown extension options")
}
