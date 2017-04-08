package commands

import (
	"path"

	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/cmd/tendermint/commands"
	tmcfg "github.com/tendermint/tendermint/config/tendermint"
)

var UnsafeResetAllCmd = &cobra.Command{
	Use:   "unsafe_reset_all",
	Short: "Reset all blockchain data",
	Run:   unsafeResetAllCmd,
}

func unsafeResetAllCmd(cmd *cobra.Command, args []string) {
	basecoinDir := BasecoinRoot("")
	tmDir := path.Join(basecoinDir)
	tmConfig := tmcfg.GetConfig(tmDir)

	commands.ResetAll(tmConfig, log)
}
