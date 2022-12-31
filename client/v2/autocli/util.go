package autocli

import (
	"strings"

	"github.com/spf13/cobra"
)

// findSubCommand finds a sub-command of the provided command whose Use
// string is or begins with the provided subCmdName.
func findSubCommand(cmd *cobra.Command, subCmdName string) *cobra.Command {
	for _, subCmd := range cmd.Commands() {
		use := subCmd.Use
		if use == subCmdName || strings.HasPrefix(use, subCmdName+" ") {
			return subCmd
		}
	}
	return nil
}
