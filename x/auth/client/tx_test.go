package client_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	priv = ed25519.GenPrivKey()
	addr = sdk.AccAddress(priv.PubKey().Address())
)

func TestParseQueryResponse(t *testing.T) {
	simRes := &sdk.SimulationResponse{
		GasInfo: sdk.GasInfo{GasUsed: 10, GasWanted: 20},
		Result:  &sdk.Result{Data: []byte("tx data"), Log: "log"},
	}

	bz, err := codec.ProtoMarshalJSON(simRes)
	require.NoError(t, err)

	res, err := authclient.ParseQueryResponse(bz)
	require.NoError(t, err)
	require.Equal(t, 10, int(res.GasInfo.GasUsed))
	require.NotNil(t, res.Result)

	res, err = authclient.ParseQueryResponse([]byte("fuzzy"))
	require.Error(t, err)
}

func TestCalculateGas(t *testing.T) {
	cdc := makeCodec()
	makeQueryFunc := func(gasUsed uint64, wantErr bool) func(string, []byte) ([]byte, int64, error) {
		return func(string, []byte) ([]byte, int64, error) {
			if wantErr {
				return nil, 0, errors.New("query failed")
			}
			simRes := &sdk.SimulationResponse{
				GasInfo: sdk.GasInfo{GasUsed: gasUsed, GasWanted: gasUsed},
				Result:  &sdk.Result{Data: []byte("tx data"), Log: "log"},
			}

			bz, _ := codec.ProtoMarshalJSON(simRes)
			return bz, 0, nil
		}
	}

	type args struct {
		queryFuncGasUsed uint64
		queryFuncWantErr bool
		adjustment       float64
	}

	tests := []struct {
		name         string
		args         args
		wantEstimate uint64
		wantAdjusted uint64
		expPass      bool
	}{
		{"error", args{0, true, 1.2}, 0, 0, false},
		{"adjusted gas", args{10, false, 1.2}, 10, 12, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			queryFunc := makeQueryFunc(tt.args.queryFuncGasUsed, tt.args.queryFuncWantErr)
			simRes, gotAdjusted, err := authclient.CalculateGas(queryFunc, cdc, []byte(""), tt.args.adjustment)
			if tt.expPass {
				require.NoError(t, err)
				require.Equal(t, simRes.GasInfo.GasUsed, tt.wantEstimate)
				require.Equal(t, gotAdjusted, tt.wantAdjusted)
				require.NotNil(t, simRes.Result)
			} else {
				require.Error(t, err)
				require.Nil(t, simRes.Result)
			}
		})
	}
}

func TestDefaultTxEncoder(t *testing.T) {
	cdc := makeCodec()

	defaultEncoder := authtypes.DefaultTxEncoder(cdc)
	encoder := authclient.GetTxEncoder(cdc)

	compareEncoders(t, defaultEncoder, encoder)
}

func TestReadTxFromFile(t *testing.T) {
	t.Parallel()
	encodingConfig := simapp.MakeEncodingConfig()

	txCfg := encodingConfig.TxConfig
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithInterfaceRegistry(encodingConfig.InterfaceRegistry)
	clientCtx = clientCtx.WithTxConfig(txCfg)

	feeAmount := sdk.Coins{sdk.NewInt64Coin("atom", 150)}
	gasLimit := uint64(50000)
	memo := "foomemo"

	txBuilder := txCfg.NewTxBuilder()
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetMemo(memo)

	// Write it to the file
	encodedTx, err := txCfg.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	jsonTxFile, cleanup := testutil.WriteToNewTempFile(t, string(encodedTx))
	t.Cleanup(cleanup)

	// Read it back
	decodedTx, err := authclient.ReadTxFromFile(clientCtx, jsonTxFile.Name())
	require.NoError(t, err)
	txBldr, err := txCfg.WrapTxBuilder(decodedTx)
	require.NoError(t, err)
	t.Log(txBuilder.GetTx())
	t.Log(txBldr.GetTx())
	require.Equal(t, txBuilder.GetTx().GetMemo(), txBldr.GetTx().GetMemo())
	require.Equal(t, txBuilder.GetTx().GetFee(), txBldr.GetTx().GetFee())
}

func TestBatchScanner_Scan(t *testing.T) {
	t.Parallel()
	encodingConfig := simappparams.MakeEncodingConfig()
	std.RegisterCodec(encodingConfig.Amino)

	txGen := encodingConfig.TxConfig
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithTxConfig(txGen)

	batch1 := `{"msg":[],"fee":{"amount":[{"denom":"atom","amount":"150"}],"gas":"50000"},"signatures":[],"memo":"foomemo"}
{"msg":[],"fee":{"amount":[{"denom":"atom","amount":"150"}],"gas":"10000"},"signatures":[],"memo":"foomemo"}
{"msg":[],"fee":{"amount":[{"denom":"atom","amount":"1"}],"gas":"10000"},"signatures":[],"memo":"foomemo"}
`
	batch2 := `{"msg":[],"fee":{"amount":[{"denom":"atom","amount":"150"}],"gas":"50000"},"signatures":[],"memo":"foomemo"}
malformed
{"msg":[],"fee":{"amount":[{"denom":"atom","amount":"1"}],"gas":"10000"},"signatures":[],"memo":"foomemo"}
`
	batch3 := `{"msg":[],"fee":{"amount":[{"denom":"atom","amount":"150"}],"gas":"50000"},"signatures":[],"memo":"foomemo"}
{"msg":[],"fee":{"amount":[{"denom":"atom","amount":"1"}],"gas":"10000"},"signatures":[],"memo":"foomemo"}`
	batch4 := `{"msg":[],"fee":{"amount":[{"denom":"atom","amount":"150"}],"gas":"50000"},"signatures":[],"memo":"foomemo"}

{"msg":[],"fee":{"amount":[{"denom":"atom","amount":"1"}],"gas":"10000"},"signatures":[],"memo":"foomemo"}
`
	tests := []struct {
		name               string
		batch              string
		wantScannerError   bool
		wantUnmarshalError bool
		numTxs             int
	}{
		{"good batch", batch1, false, false, 3},
		{"malformed", batch2, false, true, 1},
		{"missing trailing newline", batch3, false, false, 2},
		{"empty line", batch4, false, true, 1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			scanner, i := authclient.NewBatchScanner(clientCtx.TxConfig, strings.NewReader(tt.batch)), 0
			for scanner.Scan() {
				_ = scanner.Tx()
				i++
			}
			require.Equal(t, tt.wantScannerError, scanner.Err() != nil)
			require.Equal(t, tt.wantUnmarshalError, scanner.UnmarshalErr() != nil)
			require.Equal(t, tt.numTxs, i)
		})
	}
}

func compareEncoders(t *testing.T, expected sdk.TxEncoder, actual sdk.TxEncoder) {
	msgs := []sdk.Msg{testdata.NewTestMsg(addr)}
	tx := authtypes.NewStdTx(msgs, authtypes.StdFee{}, []authtypes.StdSignature{}, "")

	defaultEncoderBytes, err := expected(tx)
	require.NoError(t, err)
	encoderBytes, err := actual(tx)
	require.NoError(t, err)
	require.Equal(t, defaultEncoderBytes, encoderBytes)
}

func makeCodec() *codec.Codec {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)
	authtypes.RegisterCodec(cdc)
	cdc.RegisterConcrete(testdata.TestMsg{}, "cosmos-sdk/Test", nil)
	return cdc
}
