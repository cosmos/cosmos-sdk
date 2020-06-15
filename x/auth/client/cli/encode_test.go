package cli

import (
	"encoding/base64"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
)

func TestGetCommandEncode(t *testing.T) {
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithTxGenerator(simappparams.MakeEncodingConfig().TxGenerator)

	cmd := GetEncodeCommand(clientCtx)
	decodeCmd := GetDecodeCommand(clientCtx)
	encodingConfig := simappparams.MakeEncodingConfig()
	encodingConfig.Amino.RegisterConcrete(authtypes.StdTx{}, "cosmos-sdk/StdTx", nil)
	sdk.RegisterCodec(encodingConfig.Amino)

	txGen := encodingConfig.TxGenerator

	// Build a test transaction
	fee := authtypes.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	stdTx := authtypes.NewStdTx([]sdk.Msg{}, fee, []authtypes.StdSignature{}, "foomemo")
	JSONEncoded, err := txGen.TxJSONEncoder()(stdTx)
	require.NoError(t, err)

	txFile, cleanup := tests.WriteToNewTempFile(t, string(JSONEncoded))
	txFileName := txFile.Name()
	t.Cleanup(cleanup)

	err = cmd.RunE(cmd, []string{txFileName})
	require.NoError(t, err)

	clientCtx = clientCtx.WithTxGenerator(txGen)

	// Encode transaction
	txBytes, err := clientCtx.TxGenerator.TxEncoder()(stdTx)
	require.NoError(t, err)

	// Convert the transaction into base64 encoded string
	base64Encoded := base64.StdEncoding.EncodeToString(txBytes)

	// Execute the command
	err = runDecodeTxString(clientCtx)(decodeCmd, []string{base64Encoded})
	require.NoError(t, err)
}
