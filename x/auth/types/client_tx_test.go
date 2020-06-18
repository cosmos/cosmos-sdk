package types_test

import (
	"github.com/cosmos/cosmos-sdk/client"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTxEncoder(t *testing.T) {
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithTxGenerator(simappparams.MakeEncodingConfig().TxGenerator)

	encodingConfig := simappparams.MakeEncodingConfig()
	sdk.RegisterCodec(encodingConfig.Amino)

	txGen := encodingConfig.TxGenerator
	clientCtx = clientCtx.WithTxGenerator(txGen)

	// Build a test transaction
	fee := authtypes.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	stdTx := authtypes.NewStdTx([]sdk.Msg{nil}, fee, []authtypes.StdSignature{}, "foomemo")

	// Encode transaction
	txBytes, err := clientCtx.TxGenerator.TxEncoder()(stdTx)
	require.NoError(t, err)
	require.NotNil(t, txBytes)

	tx, err := clientCtx.TxGenerator.TxDecoder()(txBytes)
	require.NoError(t, err)
	require.Equal(t, []sdk.Msg{nil}, tx.GetMsgs())
}
