package benchmark_test

import (
	"bytes"
	"context"
	"testing"

	"gotest.tools/v3/assert"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	signing "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type TxBenchmarkSuite struct {
	cfg         network.Config
	network     *network.Network
	txHeight    int64
	queryClient tx.ServiceClient
}

func BenchmarkTx(b *testing.B) {
	s := NewTxBenchmarkSuite(b)
	b.Cleanup(s.Close)

	val := s.network.Validators[0]
	txBuilder := mkTxBuilder(b, s)

	txBytes, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	assert.NilError(b, err)

	testCases := []struct {
		name string
		req  *tx.SimulateRequest
	}{
		{"valid request with tx_bytes", &tx.SimulateRequest{TxBytes: txBytes}},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			res, err := s.queryClient.Simulate(context.Background(), tc.req)
			assert.NilError(b, err)
			assert.Assert(b, len(res.GetResult().GetEvents()) >= 10)
			assert.Assert(b, res.GetGasInfo().GetGasUsed() > 0)
		}
	}
}

func NewTxBenchmarkSuite(tb testing.TB) *TxBenchmarkSuite {
	tb.Helper()

	s := new(TxBenchmarkSuite)

	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	s.cfg = cfg

	var err error
	s.network, err = network.New(tb, tb.TempDir(), s.cfg)
	assert.NilError(tb, err)

	val := s.network.Validators[0]
	assert.NilError(tb, s.network.WaitForNextBlock())

	s.queryClient = tx.NewServiceClient(val.ClientCtx)

	msgSend := &banktypes.MsgSend{
		FromAddress: val.Address.String(),
		ToAddress:   val.Address.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdkmath.NewInt(10))),
	}

	out, err := SubmitTestTx(val.ClientCtx, msgSend, val.Moniker, TestTxConfig{Memo: "foobar"})
	assert.NilError(tb, err)

	var txRes sdk.TxResponse
	assert.NilError(tb, val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txRes))
	assert.Equal(tb, uint32(0), txRes.Code)

	resp, err := cli.GetTxResponse(s.network, val.ClientCtx, txRes.TxHash)
	assert.NilError(tb, err)
	s.txHeight = resp.Height

	assert.NilError(tb, err)
	s.txHeight = resp.Height

	return s
}

func (s *TxBenchmarkSuite) Close() {
	s.network.Cleanup()
}

func mkTxBuilder(tb testing.TB, s *TxBenchmarkSuite) client.TxBuilder {
	tb.Helper()

	val := s.network.Validators[0]
	assert.NilError(tb, s.network.WaitForNextBlock())

	txBuilder := val.ClientCtx.TxConfig.NewTxBuilder()
	feeAmount := sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 10)}
	gasLimit := testdata.NewTestGasLimit()

	assert.NilError(tb,
		txBuilder.SetMsgs(&banktypes.MsgSend{
			FromAddress: val.Address.String(),
			ToAddress:   val.Address.String(),
			Amount:      sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 10)},
		}),
	)

	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetMemo("foobar")

	txFactory := clienttx.Factory{}.
		WithChainID(val.ClientCtx.ChainID).
		WithKeybase(val.ClientCtx.Keyring).
		WithTxConfig(val.ClientCtx.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	err := authclient.SignTx(txFactory, val.ClientCtx, val.Moniker, txBuilder, false, true)
	assert.NilError(tb, err)

	return txBuilder
}

type TestTxConfig struct {
	Memo    string
	Offline bool
	AccNum  uint64
	Seq     uint64
}

func SubmitTestTx(
	clientCtx client.Context,
	msg sdk.Msg,
	signerName string,
	cfg TestTxConfig,
) (*bytes.Buffer, error) {
	txBuilder := clientCtx.TxConfig.NewTxBuilder()

	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, err
	}

	txBuilder.SetMemo(cfg.Memo)
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("stake", 10)))

	txFactory := clienttx.Factory{}.
		WithTxConfig(clientCtx.TxConfig).
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithGas(testdata.NewTestGasLimit()).
		WithFees("10stake").
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	if cfg.Offline {
		txFactory = txFactory.
			WithAccountNumber(cfg.AccNum).
			WithSequence(cfg.Seq)
	}

	if err := authclient.SignTx(txFactory, clientCtx, signerName, txBuilder, false, true); err != nil {
		return nil, err
	}

	return BroadcastTx(clientCtx, txBuilder.GetTx())
}

func BroadcastTx(clientCtx client.Context, tx sdk.Tx) (*bytes.Buffer, error) {
	txBytes, err := clientCtx.TxConfig.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}

	clientCtx = clientCtx.WithBroadcastMode("sync")

	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	bz, err := clientCtx.Codec.MarshalJSON(res)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(bz), nil
}
