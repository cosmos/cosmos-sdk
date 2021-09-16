package middleware_test

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (s *MWTestSuite) TestRejectExtensionOptionsMiddleware() {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

	txHandler := middleware.ComposeMiddlewares(noopTxHandler{}, middleware.RejectExtensionOptionsMiddleware)

	// no extension options should not trigger an error
	theTx := txBuilder.GetTx()
	_, err := txHandler.CheckTx(sdk.WrapSDKContext(ctx), theTx, abci.RequestCheckTx{})
	s.Require().NoError(err)

	extOptsTxBldr, ok := txBuilder.(tx.ExtensionOptionsTxBuilder)
	if !ok {
		// if we can't set extension options, this middleware doesn't apply and we're done
		return
	}

	// setting any extension option should cause an error
	any, err := types.NewAnyWithValue(testdata.NewTestMsg())
	s.Require().NoError(err)
	extOptsTxBldr.SetExtensionOptions(any)
	theTx = txBuilder.GetTx()
	_, err = txHandler.CheckTx(sdk.WrapSDKContext(ctx), theTx, abci.RequestCheckTx{})
	s.Require().EqualError(err, "unknown extension options")
}
