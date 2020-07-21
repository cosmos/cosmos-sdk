package cli

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestGetCommandEncode(t *testing.T) {
	encodingConfig := simappparams.MakeEncodingConfig()

	cmd := GetEncodeCommand()
	_ = testutil.ApplyMockIODiscardOutErr(cmd)

	authtypes.RegisterCodec(encodingConfig.Amino)
	sdk.RegisterCodec(encodingConfig.Amino)

	txGen := encodingConfig.TxConfig

	// Build a test transaction
	fee := authtypes.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	stdTx := authtypes.NewStdTx([]sdk.Msg{}, fee, []authtypes.StdSignature{}, "foomemo")
	JSONEncoded, err := txGen.TxJSONEncoder()(stdTx)
	require.NoError(t, err)

	txFile, cleanup := testutil.WriteToNewTempFile(t, string(JSONEncoded))
	txFileName := txFile.Name()
	t.Cleanup(cleanup)

	ctx := context.Background()
	clientCtx := client.Context{}.
		WithTxConfig(encodingConfig.TxConfig).
		WithJSONMarshaler(encodingConfig.Marshaler)
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	cmd.SetArgs([]string{txFileName})
	err = cmd.ExecuteContext(ctx)
	require.NoError(t, err)
}

func TestGetCommandDecode(t *testing.T) {
	encodingConfig := simappparams.MakeEncodingConfig()

	clientCtx := client.Context{}.
		WithTxConfig(encodingConfig.TxConfig).
		WithJSONMarshaler(encodingConfig.Marshaler)

	cmd := GetDecodeCommand()
	_ = testutil.ApplyMockIODiscardOutErr(cmd)

	sdk.RegisterCodec(encodingConfig.Amino)

	txGen := encodingConfig.TxConfig
	clientCtx = clientCtx.WithTxConfig(txGen)

	// Build a test transaction
	fee := authtypes.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	stdTx := authtypes.NewStdTx([]sdk.Msg{}, fee, []authtypes.StdSignature{}, "foomemo")

	// Encode transaction
	txBytes, err := clientCtx.TxConfig.TxEncoder()(stdTx)
	require.NoError(t, err)

	// Convert the transaction into base64 encoded string
	base64Encoded := base64.StdEncoding.EncodeToString(txBytes)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	// Execute the command
	cmd.SetArgs([]string{base64Encoded})
	require.NoError(t, cmd.ExecuteContext(ctx))
}
