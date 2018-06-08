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
	cmd := &cobra.Command{
		Use:   "simplegov",
		Short: "Define new proposals and vote by staking in existing ones",
		// Args:  cobra.MinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	proposeCmd := &cobra.Command{
		Use:   "propose",
		Short: "Submit a new proposal",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO
			// run tests N
		},
	}
	proposeCmd.Flags().AddFlagSet(fsDetails)
	proposeCmd.Flags().AddFlagSet(fsAmount)
	proposeCmd.Flags().AddFlagSet(fsProposer)
	proposeCmd.MarkFlagRequired("title")
	proposeCmd.MarkFlagRequired("description")
	proposeCmd.MarkFlagRequired("deposit")
	proposeCmd.MarkFlagRequired("address-proposer")
	cmd.AddCommand(proposeCmd)

	voteCmd := &cobra.Command{
		Use:   "vote",
		Short: "Vote on a existing open proposal",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO
			// run tests N
		},
	}
	proposeCmd.Flags().AddFlagSet(fsVote)
	voteCmd.MarkFlagRequired("proposal-id")
	voteCmd.MarkFlagRequired("description")
	voteCmd.MarkFlagRequired("address-proposer")
	cmd.AddCommand(voteCmd)
	return cmd
}
