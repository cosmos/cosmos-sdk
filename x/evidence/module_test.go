package evidence

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	eviclient "github.com/cosmos/cosmos-sdk/x/evidence/client"
)

func findSubCommand(cmd *cobra.Command, use string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if c.Use == use {
			return c
		}
	}
	return nil
}

func TestAppModuleBasicGetTxCmdReturnsNilWhenNoHandlers(t *testing.T) {
	var m AppModuleBasic

	require.Nil(t, m.GetTxCmd())
}

func TestAppModuleBasicGetTxCmdBuildsSubmitTreeFromHandlers(t *testing.T) {
	handler := eviclient.NewEvidenceHandler(func() *cobra.Command {
		return &cobra.Command{Use: "dummy-evidence"}
	})

	m := AppModuleBasic{
		evidenceHandlers: []eviclient.EvidenceHandler{handler},
	}

	root := m.GetTxCmd()
	require.NotNil(t, root)

	submit := findSubCommand(root, "submit")
	require.NotNil(t, submit, "submit subcommand should be registered when handlers are provided")

	dummy := findSubCommand(submit, "dummy-evidence")
	require.NotNil(t, dummy, "evidence handler CLI command should be registered under submit subcommand")
}
