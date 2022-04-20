package middleware_test

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
)

func (s *MWTestSuite) TestTxDecoderMiddleware() {
	ctx := s.SetupTest(true) // setup
	require := s.Require()

	// Create a tx.
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(testdata.NewTestMsg(addr1))
	require.NoError(err)
	sdkTx, txBz, err := s.createTestTx(txBuilder, []cryptotypes.PrivKey{priv1}, []uint64{1}, []uint64{0}, ctx.ChainID())
	require.NoError(err)

	// Create a custom tx.Handler that checks that the req.Tx field is
	// correctly populated.
	txReqChecker := customTxHandler{func(c context.Context, r tx.Request) (tx.Response, error) {
		require.NotNil(r.Tx)
		require.Equal(sdkTx.GetMsgs()[0], r.Tx.GetMsgs()[0])
		return tx.Response{}, nil
	}}

	testcases := []struct {
		name   string
		req    tx.Request
		expErr bool
	}{
		{"empty tx bz", tx.Request{}, true},
		{"tx bz and tx both given as inputs", tx.Request{Tx: sdkTx, TxBytes: txBz}, false},
		{"tx bz only given as input", tx.Request{TxBytes: txBz}, false},
		{"tx only given as input", tx.Request{Tx: sdkTx}, false},
	}
	for _, tc := range testcases {
		s.Run(tc.name, func() {
			txHandler := middleware.ComposeMiddlewares(
				txReqChecker,
				middleware.NewTxDecoderMiddleware(s.clientCtx.TxConfig.TxDecoder()),
			)

			// DeliverTx
			_, err := txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tc.req)

			// SimulateTx
			_, simErr := txHandler.SimulateTx(sdk.WrapSDKContext(ctx), tc.req)
			if tc.expErr {
				require.Error(err)
				require.Error(simErr)
			} else {
				require.NoError(err)
				require.NoError(simErr)
			}
		})
	}
}
