package cmd_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/transaction"
	svrcmd "cosmossdk.io/server/v2"
	"cosmossdk.io/simapp/v2"
	"cosmossdk.io/simapp/v2/simdv2/cmd"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

func TestInitCmd(t *testing.T) {
	rootCmd := cmd.NewRootCmd[transaction.Tx]()
	rootCmd.SetArgs([]string{
		"init",        // Test the init cmd
		"simapp-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", simapp.DefaultNodeHome))
}

func TestHomeFlagRegistration(t *testing.T) {
	homeDir := "/tmp/foo"

	rootCmd := cmd.NewRootCmd[transaction.Tx]()
	rootCmd.SetArgs([]string{
		"query",
		fmt.Sprintf("--%s", flags.FlagHome),
		homeDir,
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", simapp.DefaultNodeHome))

	result, err := rootCmd.Flags().GetString(flags.FlagHome)
	require.NoError(t, err)
	require.Equal(t, result, homeDir)
}
