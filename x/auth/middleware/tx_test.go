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

	testcases := []struct {
		name      string
		txHandler tx.Handler
		req       tx.Request
		expErr    bool
	}{
		{"empty tx bz", noopTxHandler, tx.Request{}, true},
		{
			"tx bz and tx populated",
			customTxHandler{func(c context.Context, r tx.Request) (tx.Response, error) {
				require.NotNil(r.Tx)
				require.Equal(sdkTx.GetMsgs()[0], r.Tx.GetMsgs()[0])
				return tx.Response{}, nil
			}},
			tx.Request{Tx: sdkTx, TxBytes: txBz},
			false,
		},
		{
			"tx bz populated only",
			customTxHandler{func(c context.Context, r tx.Request) (tx.Response, error) {
				require.NotNil(r.Tx)
				require.Equal(sdkTx.GetMsgs()[0], r.Tx.GetMsgs()[0])
				return tx.Response{}, nil
			}},
			tx.Request{TxBytes: txBz},
			false,
		},
	}
	for _, tc := range testcases {
		s.Run(tc.name, func() {
			txHandler := middleware.ComposeMiddlewares(
				tc.txHandler,
				middleware.NewTxDecoderMiddleware(s.clientCtx.TxConfig.TxDecoder()),
			)
			_, err := txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tc.req)
			if tc.expErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}
