package auto

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AutoCompleteCmd - command to generate bash autocompletions
var AutoCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "generate bash autocompletions",
	RunE:  doAutoComplete,
}

// nolint - flags
const (
	FlagOutput = "file"
)

func init() {
	AutoCompleteCmd.Flags().String(FlagOutput, "", "file to output bash autocompletion")
	AutoCompleteCmd.MarkFlagFilename(FlagOutput)
}

func doAutoComplete(cmd *cobra.Command, args []string) error {
	output := viper.GetString(FlagOutput)
	if output == "" {
		return cmd.Root().GenBashCompletion(os.Stdout)
	}
	return cmd.Root().GenBashCompletionFile(output)
}
