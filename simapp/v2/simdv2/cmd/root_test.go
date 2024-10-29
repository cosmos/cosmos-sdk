package cmd_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/simapp/v2/simdv2/cmd"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

func TestInitCmd(t *testing.T) {
	args := []string{
		"init",        // Test the init cmd
		"simapp-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
	}
	rootCmd, err := cmd.NewRootCmd[transaction.Tx](args...)
	require.NoError(t, err)
	rootCmd.SetArgs(args)
	require.NoError(t, rootCmd.Execute())
}

func TestHomeFlagRegistration(t *testing.T) {
	homeDir := "/tmp/foo"
	args := []string{
		"query",
		fmt.Sprintf("--%s", flags.FlagHome),
		homeDir,
	}

	rootCmd, err := cmd.NewRootCmd[transaction.Tx](args...)
	require.NoError(t, err)
	rootCmd.SetArgs(args)
	require.NoError(t, rootCmd.Execute())

	result, err := rootCmd.Flags().GetString(flags.FlagHome)
	require.NoError(t, err)
	require.Equal(t, result, homeDir)
}
