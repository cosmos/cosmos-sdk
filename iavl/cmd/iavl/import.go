package main

import (
	"fmt"
	"os"

	"cosmossdk.io/log/v2"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

func newImportCmd() *cobra.Command {
	var from, to, format string
	cmd := &cobra.Command{
		Use:     "import --from [from] --to [to] --format [v1-leveldb]",
		Aliases: []string{"v"},
		Short:   "Interactively browse IAVL store data",
		Args:    cobra.ExactArgs(0),
	}
	cmd.Flags().StringVar(&from, "from", "", "The source directory to import from")
	cmd.Flags().StringVar(&to, "to", "", "The destination directory to import to")
	cmd.Flags().StringVar(&format, "format", "v1-leveldb", "The format of the source data (currently only v1-leveldb is supported)")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if from == "" || to == "" || format == "" {
			return cmd.Help()
		}

		db, err := dbm.NewGoLevelDB("", from, nil)
		if err != nil {
			return fmt.Errorf("failed to open source database: %w", err)
		}

		return internal.ImportIAVLV1MultiStore(db, to, log.NewLogger(os.Stdout))
	}
	return cmd
}
