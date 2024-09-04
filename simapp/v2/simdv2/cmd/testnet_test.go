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
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func TestInitTestFilesCmd(t *testing.T) {
	rootCmd := cmd.NewRootCmd[transaction.Tx]()
	rootCmd.SetArgs([]string{
		"testnet", // Test the testnet init-files command
		"init-files",
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest), // Set keyring-backend to test
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", simapp.DefaultNodeHome))
}
