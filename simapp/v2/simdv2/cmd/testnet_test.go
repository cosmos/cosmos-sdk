package cmd_test

import (
	"testing"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/simapp/v2"
	"cosmossdk.io/simapp/v2/simdv2/cmd"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/stretchr/testify/require"
)

func TestInitTestFilesCmd(t *testing.T) {
	rootCmd := cmd.NewRootCmd[transaction.Tx]()
	rootCmd.SetArgs([]string{
		"testnet", // Test the testnet init-files command
		"init-files",
		"--keyring-backend", // Set keyring-backend to test
		"test",
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", simapp.DefaultNodeHome))
}
