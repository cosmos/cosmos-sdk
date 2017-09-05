package seeds

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/tendermint/light-client/certifiers"
)

var exportCmd = &cobra.Command{
	Use:   "export <file>",
	Short: "Export selected seeds to given file",
	Long: `Exports the most recent seed to a binary file.
If desired, you can select by an older height or validator hash.
`,
	RunE:         commands.RequireInit(exportSeed),
	SilenceUsage: true,
}

func init() {
	exportCmd.Flags().Int(heightFlag, 0, "Show the seed with closest height to this")
	exportCmd.Flags().String(hashFlag, "", "Show the seed matching the validator hash")
	RootCmd.AddCommand(exportCmd)
}

func exportSeed(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide a filepath to output")
	}
	path := args[0]

	// load the seed as specified
	trust, _ := commands.GetProviders()
	h := viper.GetInt(heightFlag)
	hash := viper.GetString(hashFlag)
	seed, err := loadSeed(trust, h, hash, "")
	if err != nil {
		return err
	}

	// now get the output file and write it
	return writeSeed(seed, path)
}

// TODO use certifiers function
func writeSeed(seed certifiers.Seed, path string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	stream := json.NewEncoder(f)
	err = stream.Encode(seed)
	return errors.WithStack(err)
}
