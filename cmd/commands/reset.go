package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tmlibs/cli"
)

var UnsafeResetAllCmd = &cobra.Command{
	Use:   "unsafe_reset_all",
	Short: "Reset all blockchain data",
	RunE:  unsafeResetAllCmd,
}

func unsafeResetAllCmd(cmd *cobra.Command, args []string) error {
	rootDir := viper.GetString(cli.HomeFlag)
	// wipe out rootdir if it exists before recreating it
	os.RemoveAll(rootDir)
	config.EnsureRoot(rootDir)
	return nil
}
