package client

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/cli_test"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func TestReadStdTxFromFile(t *testing.T) {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)

	// Build a test transaction
	fee := auth.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	stdTx := auth.NewStdTx([]sdk.Msg{}, fee, []auth.StdSignature{}, "foomemo")

	// Write it to the file
	encodedTx, _ := cdc.MarshalJSON(stdTx)
	jsonTxFile := clitest.WriteToNewTempFile(t, string(encodedTx))
	defer os.Remove(jsonTxFile.Name())

	// Read it back
	decodedTx, err := ReadStdTxFromFile(cdc, jsonTxFile.Name())
	require.Nil(t, err)
	require.Equal(t, decodedTx.Memo, "foomemo")
}
