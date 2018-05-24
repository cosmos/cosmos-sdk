package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/client/context"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

const (
	flagTitle          = "title"
	flagDescription    = "description"
	flagProposalType   = "type"
	flagInitialDeposit = "deposit"
	flagproposer = "proposer"
	flagdepositer = "depositer"
	flagproposalID = "proposalID"
	flagamount = "amount"
	flagvoter = "voter"
	flagoption = "option"
)

// submit a proposal tx
func SubmitProposalCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submitproposal",
		Short: "Submit a proposal along with an initial deposit",
		RunE: func(cmd *cobra.Command, args []string) error {
			title := viper.GetString(flagTitle)
			description := viper.GetString(flagDescription)
			proposalType := viper.GetString(flagProposalType)
			initialDeposit := viper.GetString(flagInitialDeposit)

			// get the from address from the name flag
			from, err := sdk.GetAddress(viper.GetString(flagproposer))
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoins(initialDeposit)
			if err != nil {
				return err
			}

			// create the message
			msg := gov.NewMsgSubmitProposal(title, description, proposalType, from, amount)
			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().String(flagTitle,"","")
	cmd.Flags().String(flagDescription,"","")
	cmd.Flags().String(flagProposalType,"","")
	cmd.Flags().String(flagInitialDeposit,"","")
	cmd.Flags().String(flagproposer,"","")


	return cmd
}

// set a new cool trend transaction
func DepositCmd(cdc *wire.Codec) *cobra.Command {
	cmd :=  &cobra.Command{
		Use:   "deposit",
		Short: "You're so cool, tell us what is cool!",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get the from address from the name flag
			from, err := sdk.GetAddress(viper.GetString(flagdepositer))
			if err != nil {
				return err
			}

			proposalID := viper.GetInt64(flagproposalID)

			amount, err := sdk.ParseCoins(viper.GetString(flagamount))
			if err != nil {
				return err
			}


			// create the message
			msg := gov.NewMsgDeposit(proposalID,from,amount)
			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}
			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().String(flagdepositer,"","")
	cmd.Flags().Int64(flagproposalID,0,"")
	cmd.Flags().String(flagamount,"","")
	return cmd
}

// set a new cool trend transaction
func VoteCmd(cdc *wire.Codec) *cobra.Command {
	cmd :=  &cobra.Command{
		Use:   "vote",
		Short: "You're so cool, tell us what is cool!",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get the from address from the name flag
			voter, err := sdk.GetAddress(viper.GetString(flagvoter))
			if err != nil {
				return err
			}

			proposalID := viper.GetInt64(flagproposalID)

			option := viper.GetString(flagoption)
			if err != nil {
				return err
			}
			// create the message
			msg := gov.NewMsgVote(voter,proposalID,option)
			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}
			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().String(flagvoter,"","")
	cmd.Flags().Int64(flagproposalID,0,"")
	cmd.Flags().String(flagoption,"","")

	return cmd
}
