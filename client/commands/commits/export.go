package commits

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/certifiers/files"

	"github.com/cosmos/cosmos-sdk/client/commands"
)

var exportCmd = &cobra.Command{
	Use:   "export <file>",
	Short: "Export selected commits to given file",
	Long: `Exports the most recent commit to a binary file.
If desired, you can select by an older height or validator hash.
`,
	RunE:         commands.RequireInit(exportCommit),
	SilenceUsage: true,
}

func init() {
	exportCmd.Flags().Int(heightFlag, 0, "Show the commit with closest height to this")
	exportCmd.Flags().String(hashFlag, "", "Show the commit matching the validator hash")
	RootCmd.AddCommand(exportCmd)
}

func exportCommit(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide a filepath to output")
	}
	path := args[0]

	// load the seed as specified
	trust, _ := commands.GetProviders()
	h := viper.GetInt(heightFlag)
	hash := viper.GetString(hashFlag)
	fc, err := loadCommit(trust, h, hash, "")
	if err != nil {
		return err
	}

	// now get the output file and write it
	return files.SaveFullCommitJSON(fc, path)
}
