package cli_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	evidcli "github.com/cosmos/cosmos-sdk/x/evidence/client/cli"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
)

func findSubCommand(cmd *cobra.Command, use string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if c.Use == use {
			return c
		}
	}
	return nil
}

func TestGetTxCmdRegistersSubmitAndChildCommands(t *testing.T) {
	child := &cobra.Command{Use: "dummy-evidence"}

	root := evidcli.GetTxCmd([]*cobra.Command{child})
	require.NotNil(t, root)
	require.Equal(t, evidencetypes.ModuleName, root.Use)

	submit := findSubCommand(root, "submit")
	require.NotNil(t, submit, "submit subcommand should be registered under evidence tx root")

	dummy := findSubCommand(submit, "dummy-evidence")
	require.NotNil(t, dummy, "child evidence command should be registered under submit subcommand")
}
