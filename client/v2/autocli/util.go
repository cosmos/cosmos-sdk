package autocli

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/client/v2/internal/strcase"
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

// topLevelCmd creates a new top-level command with the provided name and
// description. The command will have DisableFlagParsing set to false and
// SuggestionsMinimumDistance set to 2.
func topLevelCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:                        use,
		Short:                      short,
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       validateCmd,
	}
}

func protoNameToCliName(name protoreflect.Name) string {
	return strcase.ToKebab(string(name))
}
