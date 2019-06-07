package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	priv = ed25519.GenPrivKey()
	addr = sdk.AccAddress(priv.PubKey().Address())
)

func TestParseQueryResponse(t *testing.T) {
	cdc := makeCodec()
	sdkResBytes := cdc.MustMarshalBinaryLengthPrefixed(sdk.Result{GasUsed: 10})
	gas, err := parseQueryResponse(cdc, sdkResBytes)
	assert.Equal(t, gas, uint64(10))
	assert.Nil(t, err)
	gas, err = parseQueryResponse(cdc, []byte("fuzzy"))
	assert.Equal(t, gas, uint64(0))
	assert.Error(t, err)
}

func TestCalculateGas(t *testing.T) {
	cdc := makeCodec()
	makeQueryFunc := func(gasUsed uint64, wantErr bool) func(string, common.HexBytes) ([]byte, error) {
		return func(string, common.HexBytes) ([]byte, error) {
			if wantErr {
				return nil, errors.New("")
			}
			return cdc.MustMarshalBinaryLengthPrefixed(sdk.Result{GasUsed: gasUsed}), nil
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
		wantErr      bool
	}{
		{"error", args{0, true, 1.2}, 0, 0, true},
		{"adjusted gas", args{10, false, 1.2}, 10, 12, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryFunc := makeQueryFunc(tt.args.queryFuncGasUsed, tt.args.queryFuncWantErr)
			gotEstimate, gotAdjusted, err := CalculateGas(queryFunc, cdc, []byte(""), tt.args.adjustment)
			assert.Equal(t, err != nil, tt.wantErr)
			assert.Equal(t, gotEstimate, tt.wantEstimate)
			assert.Equal(t, gotAdjusted, tt.wantAdjusted)
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
	cdc := makeCodec()

	// Build a test transaction
	fee := authtypes.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	stdTx := authtypes.NewStdTx([]sdk.Msg{}, fee, nil, "foomemo")

	// Write it to the file
	encodedTx, err := cdc.MarshalJSON(stdTx)
	require.NoError(t, err)
	jsonTxFile := writeToNewTempFile(t, string(encodedTx))
	defer os.Remove(jsonTxFile.Name())

	// Read it back
	decodedTx, err := ReadStdTxFromFile(cdc, jsonTxFile.Name())
	require.NoError(t, err)
	require.Equal(t, decodedTx.Memo, "foomemo")
}

func TestValidateCmd(t *testing.T) {
	// Setup root and subcommands
	rootCmd := &cobra.Command{
		Use: "root",
	}
	queryCmd := &cobra.Command{
		Use: "query",
	}
	rootCmd.AddCommand(queryCmd)

	// Command being tested
	distCmd := &cobra.Command{
		Use:                        "distr",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}
	queryCmd.AddCommand(distCmd)

	commissionCmd := &cobra.Command{
		Use: "commission",
	}
	distCmd.AddCommand(commissionCmd)

	tests := []struct {
		reason  string
		args    []string
		wantErr bool
	}{
		{"misspelled command", []string{"comission"}, true},
		{"no command provided", []string{}, false},
		{"help flag", []string{"commission", "--help"}, false},
		{"shorthand help flag", []string{"commission", "-h"}, false},
	}

	for _, tt := range tests {
		err := ValidateCmd(distCmd, tt.args)
		assert.Equal(t, tt.wantErr, err != nil, tt.reason)
	}

}

// aux functions

func compareEncoders(t *testing.T, expected sdk.TxEncoder, actual sdk.TxEncoder) {
	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	tx := authtypes.NewStdTx(msgs, sdk.Fee(nil), []sdk.Signature{}, "")

	defaultEncoderBytes, err := expected(tx)
	require.NoError(t, err)
	encoderBytes, err := actual(tx)
	require.NoError(t, err)
	require.Equal(t, defaultEncoderBytes, encoderBytes)
}

func writeToNewTempFile(t *testing.T, data string) *os.File {
	fp, err := ioutil.TempFile(os.TempDir(), "client_tx_test")
	require.NoError(t, err)

	_, err = fp.WriteString(data)
	require.NoError(t, err)

	return fp
}

func makeCodec() *codec.Codec {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	authtypes.RegisterCodec(cdc)
	cdc.RegisterConcrete(sdk.TestMsg{}, "cosmos-sdk/Test", nil)
	return cdc
}
