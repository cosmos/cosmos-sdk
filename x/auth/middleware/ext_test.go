package middleware_test

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typestx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

func (s *MWTestSuite) TestExtensionOptionsMiddleware() {
	testCases := []struct {
		msg   string
		allow bool
	}{
		{"allow extension", true},
		{"reject extension", false},
	}
	for _, tc := range testCases {
		s.Run(tc.msg, func() {
			ctx := s.SetupTest(true) // setup
			txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

			txHandler := middleware.ComposeMiddlewares(noopTxHandler, middleware.NewExtensionOptionsMiddleware(func(_ *codectypes.Any) bool {
				return tc.allow
			}))

			// no extension options should not trigger an error
			theTx := txBuilder.GetTx()
			_, _, err := txHandler.CheckTx(sdk.WrapSDKContext(ctx), typestx.Request{Tx: theTx}, typestx.RequestCheckTx{})
			s.Require().NoError(err)

			extOptsTxBldr, ok := txBuilder.(tx.ExtensionOptionsTxBuilder)
			if !ok {
				// if we can't set extension options, this middleware doesn't apply and we're done
				return
			}

			// set an extension option and check
			any, err := types.NewAnyWithValue(testdata.NewTestMsg())
			s.Require().NoError(err)
			extOptsTxBldr.SetExtensionOptions(any)
			theTx = txBuilder.GetTx()
			_, _, err = txHandler.CheckTx(sdk.WrapSDKContext(ctx), typestx.Request{Tx: theTx}, typestx.RequestCheckTx{})
			if tc.allow {
				s.Require().NoError(err)
			} else {
				s.Require().EqualError(err, "unknown extension options")
			}
		})
	}
}
