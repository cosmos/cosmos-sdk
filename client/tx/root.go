package tx

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/wire"
)

// type used to pass around the provided cdc
type commander struct {
	cdc *wire.Codec
}

// AddCommands adds a number of tx-query related subcommands
func AddCommands(cmd *cobra.Command, cdc *wire.Codec) {
	cmdr := commander{cdc}
	cmd.AddCommand(
		SearchTxCmd(cmdr),
		QueryTxCmd(cmdr),
	)
}
