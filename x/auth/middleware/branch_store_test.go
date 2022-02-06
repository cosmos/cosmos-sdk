package middleware_test

import (
	"context"
	"fmt"
	"math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

var blockMaxGas = uint64(simapp.DefaultConsensusParams.Block.MaxGas)

func (s *MWTestSuite) TestBranchStore() {
	testcases := []struct {
		name         string
		gasToConsume uint64 // gas to consume in the msg execution
		panicTx      bool   // panic explicitly in tx execution
		expErr       bool
	}{
		{"less than block gas meter", 10, false, false},
		{"more than block gas meter", blockMaxGas, false, true},
		{"more than block gas meter", uint64(float64(blockMaxGas) * 1.2), false, true},
		{"consume MaxUint64", math.MaxUint64, false, true},
		{"consume block gas when paniced", 10, true, true},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			ctx := s.SetupTest(true).WithBlockGasMeter(sdk.NewGasMeter(blockMaxGas)) // setup
			txBuilder := s.clientCtx.TxConfig.NewTxBuilder()

			// tx fee
			feeCoin := sdk.NewCoin("atom", sdk.NewInt(150))
			feeAmount := sdk.NewCoins(feeCoin)

			// test account and fund
			priv1, _, addr1 := testdata.KeyTestPubAddr()
			err := s.app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, feeAmount)
			s.Require().NoError(err)
			err = s.app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr1, feeAmount)
			s.Require().NoError(err)
			s.Require().Equal(feeCoin.Amount, s.app.BankKeeper.GetBalance(ctx, addr1, feeCoin.Denom).Amount)
			seq, _ := s.app.AccountKeeper.GetSequence(ctx, addr1)
			s.Require().Equal(uint64(0), seq)

			// testMsgTxHandler is a test txHandler that handles one single TestMsg,
			// consumes the given `tc.gasToConsume`, and sets the bank store "ok" key to "ok".
			var testMsgTxHandler = customTxHandler{func(ctx context.Context, req tx.Request) (tx.Response, error) {
				msg, ok := req.Tx.GetMsgs()[0].(*testdata.TestMsg)
				if !ok {
					return tx.Response{}, fmt.Errorf("Wrong Msg type, expected %T, got %T", (*testdata.TestMsg)(nil), msg)
				}

				sdkCtx := sdk.UnwrapSDKContext(ctx)
				sdkCtx.KVStore(s.app.GetKey("bank")).Set([]byte("ok"), []byte("ok"))
				sdkCtx.GasMeter().ConsumeGas(tc.gasToConsume, "TestMsg")
				if tc.panicTx {
					panic("panic in tx execution")
				}
				return tx.Response{}, nil
			}}

			txHandler := middleware.ComposeMiddlewares(
				testMsgTxHandler,
				middleware.NewTxDecoderMiddleware(s.clientCtx.TxConfig.TxDecoder()),
				middleware.GasTxMiddleware,
				middleware.RecoveryTxMiddleware,
				middleware.DeductFeeMiddleware(s.app.AccountKeeper, s.app.BankKeeper, s.app.FeeGrantKeeper),
				middleware.IncrementSequenceMiddleware(s.app.AccountKeeper),
				middleware.WithBranchedStore,
				middleware.ConsumeBlockGasMiddleware,
			)

			// msg and signatures
			msg := testdata.NewTestMsg(addr1)
			var gasLimit uint64 = math.MaxUint64 // no limit on sdk.GasMeter
			s.Require().NoError(txBuilder.SetMsgs(msg))
			txBuilder.SetFeeAmount(feeAmount)
			txBuilder.SetGasLimit(gasLimit)

			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
			testTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
			s.Require().NoError(err)

			_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx})

			bankStore := ctx.KVStore(s.app.GetKey("bank"))
			okValue := bankStore.Get([]byte("ok"))

			if tc.expErr {
				s.Require().Error(err)
				if tc.panicTx {
					s.Require().True(sdkerrors.IsOf(err, sdkerrors.ErrPanic))
				} else {
					s.Require().True(sdkerrors.IsOf(err, sdkerrors.ErrOutOfGas))
				}
				s.Require().Empty(okValue)
			} else {
				s.Require().NoError(err)
				s.Require().Equal([]byte("ok"), okValue)
			}
			// block gas is always consumed
			baseGas := uint64(24564) // baseGas is the gas consumed by middlewares
			expGasConsumed := addUint64Saturating(tc.gasToConsume, baseGas)
			s.Require().Equal(expGasConsumed, ctx.BlockGasMeter().GasConsumed())
			// tx fee is always deducted
			s.Require().Equal(int64(0), s.app.BankKeeper.GetBalance(ctx, addr1, feeCoin.Denom).Amount.Int64())
			// sender's sequence is always increased
			seq, err = s.app.AccountKeeper.GetSequence(ctx, addr1)
			s.Require().NoError(err)
			s.Require().Equal(uint64(1), seq)
		})
	}
}

func addUint64Saturating(a, b uint64) uint64 {
	if math.MaxUint64-a < b {
		return math.MaxUint64
	}

	return a + b
}
