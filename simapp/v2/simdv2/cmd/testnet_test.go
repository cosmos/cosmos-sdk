package cmd_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/simapp/v2"
	"cosmossdk.io/simapp/v2/simdv2/cmd"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/stretchr/testify/require"
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
