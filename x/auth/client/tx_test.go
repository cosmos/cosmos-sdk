package client

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/testdata"
	"github.com/cosmos/cosmos-sdk/testutil"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"

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

	res, err := parseQueryResponse(bz)
	require.NoError(t, err)
	require.Equal(t, 10, int(res.GasInfo.GasUsed))
	require.NotNil(t, res.Result)

	res, err = parseQueryResponse([]byte("fuzzy"))
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
			simRes, gotAdjusted, err := CalculateGas(queryFunc, cdc, []byte(""), tt.args.adjustment)
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
	encoder := GetTxEncoder(cdc)

	compareEncoders(t, defaultEncoder, encoder)
}

func TestConfiguredTxEncoder(t *testing.T) {
	cdc := makeCodec()

	customEncoder := func(tx sdk.Tx) ([]byte, error) {
		return json.Marshal(tx)
	}

	config := sdk.GetConfig()
	config.SetTxEncoder(customEncoder)

	encoder := GetTxEncoder(cdc)

	compareEncoders(t, customEncoder, encoder)
}

func TestReadStdTxFromFile(t *testing.T) {
	t.Parallel()

	encodingConfig := simappparams.MakeEncodingConfig()
	sdk.RegisterCodec(encodingConfig.Amino)

	txGen := encodingConfig.TxConfig
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithTxConfig(txGen)

	// Build a test transaction
	fee := authtypes.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	stdTx := authtypes.NewStdTx([]sdk.Msg{}, fee, []authtypes.StdSignature{}, "foomemo")

	// Write it to the file
	encodedTx, err := txGen.TxJSONEncoder()(stdTx)
	require.NoError(t, err)
	jsonTxFile, cleanup := testutil.WriteToNewTempFile(t, string(encodedTx))
	t.Cleanup(cleanup)

	// Read it back
	decodedTx, err := ReadTxFromFile(clientCtx, jsonTxFile.Name())
	require.NoError(t, err)
	require.Equal(t, decodedTx.(authtypes.StdTx).Memo, "foomemo")
}

func TestBatchScanner_Scan(t *testing.T) {
	t.Parallel()
	cdc := codec.New()
	sdk.RegisterCodec(cdc)

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
			scanner, i := NewBatchScanner(cdc, strings.NewReader(tt.batch)), 0
			for scanner.Scan() {
				_ = scanner.StdTx()
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

func TestPrepareTxBuilder(t *testing.T) {
	cdc := makeCodec()

	encodingConfig := simappparams.MakeEncodingConfig()
	sdk.RegisterCodec(encodingConfig.Amino)

	fromAddr := sdk.AccAddress("test-addr0000000000")
	fromAddrStr := fromAddr.String()

	var accNum uint64 = 10
	var accSeq uint64 = 17

	txGen := encodingConfig.TxConfig
	clientCtx := client.Context{}
	clientCtx = clientCtx.
		WithTxConfig(txGen).
		WithJSONMarshaler(encodingConfig.Marshaler).
		WithAccountRetriever(client.TestAccountRetriever{Accounts: map[string]struct {
			Address sdk.AccAddress
			Num     uint64
			Seq     uint64
		}{
			fromAddrStr: {
				Address: fromAddr,
				Num:     accNum,
				Seq:     accSeq,
			},
		}}).
		WithFromAddress(fromAddr)

	bldr := authtypes.NewTxBuilder(
		authtypes.DefaultTxEncoder(cdc), 0, 0,
		200000, 1.1, false, "test-chain",
		"test-builder", sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))),
		sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10000, sdk.Precision))})

	bldr, err := PrepareTxBuilder(bldr, clientCtx)
	require.NoError(t, err)
	require.Equal(t, accNum, bldr.AccountNumber())
	require.Equal(t, accSeq, bldr.Sequence())
}

func makeCodec() *codec.Codec {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)
	authtypes.RegisterCodec(cdc)
	cdc.RegisterConcrete(testdata.TestMsg{}, "cosmos-sdk/Test", nil)
	return cdc
}
