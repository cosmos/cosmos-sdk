package main

import (
	"os"

	"cosmossdk.io/log/v2"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/iavlx/internal"
)

// newImportCmd creates the offline import command for one-time migration from iavl/v1.
// It reads the v1 LevelDB database, exports each store's tree, and imports it into
// iavlx format using the Importer (see internal/importer.go).
func newImportCmd() *cobra.Command {
	var from, to, format string
	cmd := &cobra.Command{
		Use:     "import --from [from] --to [to] --format [v1-leveldb]",
		Short:   "Import an iavl/v1 LevelDB multi-store into iavlx format (one-time offline migration)",
		Args:    cobra.ExactArgs(0),
	}
	cmd.Flags().StringVar(&from, "from", "", "The source directory to import from")
	cmd.Flags().StringVar(&to, "to", "", "The destination directory to import to")
	cmd.Flags().StringVar(&format, "format", "v1-leveldb", "The format of the source data (currently only v1-leveldb is supported)")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if from == "" || to == "" || format == "" {
			return cmd.Help()
		}

		return internal.ImportIAVLV1MultiStore(from, to, log.NewLogger(os.Stdout))
	}
	return cmd
}
