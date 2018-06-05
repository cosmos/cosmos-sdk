package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/wire"
)

// // GetSimpleGovCmdDefault gets the cmd for the simpleGov. type
// func GetSimpleGovCmdDefault(storeName string, cdc *wire.Codec) *cobra.Command {
// 	return SimpleGovCmd(storeName, cdc)
// }

// SimpleGovCmd is the command to create proposals and vote on them
func SimpleGovCmd(storeName string, cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "simplegov",
		Short: "Define new proposals and vote by staking in existing ones",
		// Args:  cobra.MinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
