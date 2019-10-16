package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
)

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "ibcmockbank",
		Short: "IBC mockbank module transaction subcommands",
		// RunE:  client.ValidateCmd,
	}
	txCmd.AddCommand(
	// TransferTxCmd(cdc),
	)

	return txCmd
}
