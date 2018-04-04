package server

import (
	"github.com/spf13/cobra"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
)

// UnsafeResetAllCmd - extension of the tendermint command, resets initialization
func UnsafeResetAllCmd(ctx *Context) *cobra.Command {
	cmd := resetAll{ctx}
	return &cobra.Command{
		Use:   "unsafe_reset_all",
		Short: "Reset all blockchain data",
		RunE:  cmd.run,
	}
}

type resetAll struct {
	context *Context
}

func (r resetAll) run(cmd *cobra.Command, args []string) error {
	cfg := r.context.Config
	tcmd.ResetAll(cfg.DBDir(), cfg.PrivValidatorFile(), r.context.Logger)
	return nil
}
