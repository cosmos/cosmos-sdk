package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"encoding/hex"
	"github.com/pkg/errors"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"strconv"
)

const (
	flagTitle          = "title"
	flagDescription    = "description"
	flagProposalType   = "type"
	flagInitialDeposit = "deposit"
	flagproposer       = "proposer"
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

			fmt.Printf("Committed at block:%d. Hash:%s.Response:%+v \n", res.Height, res.Hash.String(), res.DeliverTx)
			return nil
		},
	}

	cmd.Flags().String(flagTitle, "", "title of proposal")
	cmd.Flags().String(flagDescription, "", "description of proposal")
	cmd.Flags().String(flagProposalType, "", "proposalType of proposal")
	cmd.Flags().String(flagInitialDeposit, "", "deposit of proposal")
	cmd.Flags().String(flagproposer, "", "proposer of proposal")

	return cmd
}

// set a new Deposit transaction
func DepositCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [depositer] [proposalID] [amount]",
		Short: "deposit your token [steak] for activing proposalI",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// get the from address from the name flag
			depositer, err := sdk.GetAddress(args[0])
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}

			// create the message
			msg := gov.NewMsgDeposit(proposalID, depositer, amount)
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
	return cmd
}

// set a new Vote transaction
func VoteCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [voter] [proposalID] [option]",
		Short: "vote for current actived proposal,option:Yes/NO/NoWithVeto/Abstain",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			voter, err := sdk.GetAddress(args[0])
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			option := args[2]
			// create the message
			msg := gov.NewMsgVote(voter, proposalID, option)

			fmt.Printf("Vote[Voter:%s,ProposalID:%d,Option:%s]", hex.EncodeToString(msg.Voter), msg.ProposalID, msg.Option)

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
	return cmd
}

func GetProposalCmd(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal [proposalID]",
		Short: "query proposal details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			ctx := context.NewCoreContextFromViper()

			key, _ := cdc.MarshalBinary(proposalID)
			res, err := ctx.Query(key, storeName)
			if len(res) == 0 || err != nil {
				return  errors.Errorf("proposalID [%d] is not existed", proposalID)
			}

			proposal := new(gov.Proposal)
			cdc.MustUnmarshalBinary(res, proposal)
			output, err := wire.MarshalJSONIndent(cdc, proposal)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			return nil
		},
	}
	return cmd
}
