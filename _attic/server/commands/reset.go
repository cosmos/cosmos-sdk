package commands

import (
	"github.com/spf13/cobra"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
)

// UnsafeResetAllCmd - extension of the tendermint command, resets initialization
var UnsafeResetAllCmd = &cobra.Command{
	Use:   "unsafe_reset_all",
	Short: "Reset all blockchain data",
	RunE:  unsafeResetAllCmd,
}

func unsafeResetAllCmd(cmd *cobra.Command, args []string) error {
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}
	tcmd.ResetAll(cfg.DBDir(), cfg.PrivValidatorFile(), logger)
	return nil
}
