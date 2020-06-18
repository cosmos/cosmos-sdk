package types_test

import (
	"testing"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func testCodec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	cdc.RegisterConcrete(sdk.TestMsg{}, "cosmos-sdk/Test", nil)
	return cdc
}

func TestStdTxGenerator(t *testing.T) {
	cdc := testCodec()
	txGen := types.StdTxGenerator{Cdc: cdc}
	suite.Run(t, tx.NewTxGeneratorTestSuite(txGen))
}

func TestTxEncoder(t *testing.T) {
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithTxGenerator(simappparams.MakeEncodingConfig().TxGenerator)

	encodingConfig := simappparams.MakeEncodingConfig()
	sdk.RegisterCodec(encodingConfig.Amino)

	txGen := encodingConfig.TxGenerator
	clientCtx = clientCtx.WithTxGenerator(txGen)

	// Build a test transaction
	fee := types.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	stdTx := types.NewStdTx([]sdk.Msg{nil}, fee, []types.StdSignature{}, "foomemo")

	// Encode transaction
	txBytes, err := clientCtx.TxGenerator.TxEncoder()(stdTx)
	require.NoError(t, err)
	require.NotNil(t, txBytes)

	tx, err := clientCtx.TxGenerator.TxDecoder()(txBytes)
	require.NoError(t, err)
	require.Equal(t, []sdk.Msg{nil}, tx.GetMsgs())
}
