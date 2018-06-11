package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/gamarin/cosmos-sdk/client/context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// nolint
const (
	FlagDeposit     = "deposit"
	FlagTitle       = "title"
	FlagDescription = "description"

	FlagProposalID = "proposal-id"
	FlagOption     = "option"
)

// // GetSimpleGovCmdDefault gets the cmd for the simpleGov. type
// func GetSimpleGovCmdDefault(storeName string, cdc *wire.Codec) *cobra.Command {
// 	return SimpleGovCmd(storeName, cdc)
// }

// ProposeCmd is the command to create proposals
func ProposeCmd(storeName string, cdc *wire.Codec) *cobra.Command {
	proposeCmd := &cobra.Command{
		Use:   "propose",
		Short: "Submit a new proposal",
		Run: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
			proposer, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			title := viper.GetString(FlagTitle)
			description := viper.GetString(FlagDescription)
			coins := viper.GetString(FlagDeposit)
			deposit, err := sdk.ParseCoins(coins)
			if err != nil {
				return err
			}

			// TODO build msg

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
	proposeCmd.Flags().String(FlagTitle, "", "Title of the proposal")
	proposeCmd.Flags().String(FlagDescription, "", "Description of the proposal")
	proposeCmd.Flags().String(FlagDeposit, "1steak", "Amount of coins to deposit on the proposal")
	proposeCmd.MarkFlagRequired(FlagTitle)
	proposeCmd.MarkFlagRequired(FlagDescription)
	proposeCmd.MarkFlagRequired(FlagDeposit)

	return proposeCmd
}

// VoteCmd is the command to vote on proposals
func VoteCmd(storeName string, cdc *wire.Codec) *cobra.Command {
	voteCmd := &cobra.Command{
		Use:   "vote",
		Short: "Vote on a existing open proposal",
		Run: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
			voter, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			option := viper.GetString(FlagOption)
			proposalID := viper.GetInt64(FlagProposalID)

			// TODO build msg

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
	voteCmd.Flags().Int(FlagProposalID, 1, "Id of the proposal")
	voteCmd.Flags().String(FlagOption, "Yes/No/Abstain", "Vote options")
	voteCmd.MarkFlagRequired(FlagProposalID)
	voteCmd.MarkFlagRequired(FlagOption)
	return voteCmd
}
