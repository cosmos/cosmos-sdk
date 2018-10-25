package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/examples/simpleGov/x/simple_governance"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// –––––––––––– Flags ––––––––––––––––

// nolint
const (
	FlagDeposit     = "deposit"
	FlagTitle       = "title"
	FlagDescription = "description"

	FlagProposalID = "proposal-id"
	FlagOption     = "option"

	FlagActiveProposal = "active"
)

// –––––––––––– GET commands ––––––––––––––––

// GetCmdQueryProposal gets the command to get a single proposal
func GetCmdQueryProposal(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal [proposal-id]",
		Short: "Query a proposal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			proprosalID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			ctx := context.NewCLIContext()
			key := simpleGovernance.GenerateProposalKey(int64(proprosalID))
			res, err := ctx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the proposal
			proposal := new(simpleGovernance.Proposal)
			cdc.MustUnmarshalBinary(res, proposal)
			output, err := wire.MarshalJSONIndent(cdc, proposal)
			fmt.Println(string(output))

			return nil
		},
	}
	return cmd
}

// GetCmdQueryProposals gets the command to get all proposals
func GetCmdQueryProposals(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposals",
		Short: "Query all proposals",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext()

			resKVs, err := ctx.QuerySubspace([]byte("proposals"), storeName)
			if err != nil {
				return err
			}

			// parse out the proposals
			var proposals []simpleGovernance.Proposal
			for _, KV := range resKVs {
				var proposal simpleGovernance.Proposal
				cdc.MustUnmarshalBinary(KV.Value, &proposal)
				//candidates = append(proposals, proposal)
			}

			output, err := wire.MarshalJSONIndent(cdc, proposals)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	cmd.Flags().Bool(FlagActiveProposal, false, "Query only open proposals (default: true)")
	return cmd
}

// GetCmdQueryProposalVotes gets the command to get all the votes from a proposal
func GetCmdQueryProposalVotes(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal-votes [proposal-id]",
		Short: "Query all the votes from a proposal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			proprosalID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			ctx := context.NewCLIContext()
			key := simpleGovernance.GenerateProposalVotesKey(int64(proprosalID))
			resKVs, err := ctx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the proposal votes
			var votes []string
			for _, KV := range resKVs {
				var vote string
				cdc.MustUnmarshalBinary(KV.Value, &vote)
				votes = append(votes, vote)
			}

			output, err := wire.MarshalJSONIndent(cdc, votes)
			if err != nil {
				return err
			}
			fmt.Println(string(output))

			return nil
		},
	}
	return cmd
}

// GetCmdQueryProposalVote gets the command to get a single vote from a proposal
func GetCmdQueryProposalVote(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal-vote [proposal-id] [voter-addr]",
		Short: "Query a proposal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			voterAddr, err := sdk.AccAddressFromHex(args[1])
			if err != nil {
				return err
			}
			proprosalID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			ctx := context.NewCLIContext()

			key := simpleGovernance.GenerateProposalVoteKey(int64(proprosalID), voterAddr)

			res, err := ctx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validator
			var vote string
			cdc.MustUnmarshalBinary(res, vote)
			output, err := wire.MarshalJSONIndent(cdc, vote)
			fmt.Println(string(output))

			return nil

		},
	}
	return cmd
}

// –––––––––––– POST commands ––––––––––––––––

// PostCmdPropose gets the command to create proposals
func PostCmdPropose(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose",
		Short: "Submit a new proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			ctx := context.NewCLIContext().WithAccountDecoder(authcmd.GetAccountDecoder(cdc))
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

			msg := simpleGovernance.NewSubmitProposalMsg(title, description, deposit, proposer)
			return utils.SendTx(txCtx, ctx, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(FlagTitle, "", "Title of the proposal")
	cmd.Flags().String(FlagDescription, "", "Description of the proposal")
	cmd.Flags().String(FlagDeposit, "1steak", "Amount of coins to deposit on the proposal")
	cmd.MarkFlagRequired(FlagTitle)
	cmd.MarkFlagRequired(FlagDescription)
	cmd.MarkFlagRequired(FlagDeposit)

	return cmd
}

// PostCmdVote gets the command to vote on proposals
func PostCmdVote(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "Vote on a existing open proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			ctx := context.NewCLIContext().WithAccountDecoder(authcmd.GetAccountDecoder(cdc))
			voter, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			option := viper.GetString(FlagOption)
			proposalID := viper.GetInt64(FlagProposalID)

			msg := simpleGovernance.NewVoteMsg(proposalID, option, voter)
			return utils.SendTx(txCtx, ctx, []sdk.Msg{msg})
		},
	}
	cmd.Flags().Int(FlagProposalID, 1, "Id of the proposal")
	cmd.Flags().String(FlagOption, "Yes/No/Abstain", "Vote options")
	cmd.MarkFlagRequired(FlagProposalID)
	cmd.MarkFlagRequired(FlagOption)
	return cmd
}
