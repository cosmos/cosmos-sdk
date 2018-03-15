package commands

import (
	"github.com/spf13/cobra"

	wire "github.com/tendermint/go-amino"
)

func AddCommands(cmd *cobra.Command) {
	cdc := wire.NewCodec()

	cmd.AddCommand(
		IBCTransferCmd(cdc),
		IBCRelayCmd(cdc),
	)
}
