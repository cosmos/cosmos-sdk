package commands

import (
	"github.com/spf13/cobra"

	tmcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	tmcfg "github.com/tendermint/tendermint/config/tendermint"
)

var UnsafeResetAllCmd = &cobra.Command{
	Use:   "unsafe_reset_all",
	Short: "Reset all blockchain data",
	RunE:  unsafeResetAllCmd,
}

func unsafeResetAllCmd(cmd *cobra.Command, args []string) error {
	basecoinDir := BasecoinRoot("")
	tmConfig := tmcfg.GetConfig(basecoinDir)
	tmcmd.ResetAll(tmConfig, log)
	return nil
}
