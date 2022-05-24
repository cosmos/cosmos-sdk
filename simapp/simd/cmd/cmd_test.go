package cmd_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	svrcmd "github.com/Stride-Labs/cosmos-sdk/server/cmd"
	"github.com/Stride-Labs/cosmos-sdk/simapp"
	"github.com/Stride-Labs/cosmos-sdk/simapp/simd/cmd"
	"github.com/Stride-Labs/cosmos-sdk/x/genutil/client/cli"
)

func TestInitCmd(t *testing.T) {
	rootCmd, _ := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{
		"init",        // Test the init cmd
		"simapp-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
	})

	require.NoError(t, svrcmd.Execute(rootCmd, simapp.DefaultNodeHome))
}
