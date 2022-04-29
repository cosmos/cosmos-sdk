package cmd

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand_Error(t *testing.T) {
	output, err := cli.ExecTestCLICmd(client.Context{}, rootCmd, []string{"version"})
	require.Error(t, err)

	require.Contains(t, output.String(), "DAEMON_NAME is not set")
}
