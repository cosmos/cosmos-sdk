package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govClientUtils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GetCmdQueryProposal implements the query proposal command.
func GetCmdQueryProposal(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query details of a single proposal",
		Long: strings.TrimSpace(`
Query details for a proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli query gov proposal 1
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// Query the proposal
			res, err := queryProposal(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}

	return cmd
}

func queryProposal(proposalID uint64, cliCtx context.CLIContext, cdc *codec.Codec, queryRoute string) ([]byte, error) {
	// Construct query
	params := gov.NewQueryProposalParams(proposalID)
	bz, err := cdc.MarshalJSON(params)
	if err != nil {
		return nil, err
	}

	// Query store
	res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/proposal", queryRoute), bz)
	if err != nil {
		return nil, err
	}
	return res, err
}

// GetCmdQueryProposals implements a query proposals command.
func GetCmdQueryProposals(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposals",
		Short: "Query proposals with optional filters",
		Long: strings.TrimSpace(`
Query for a all proposals. You can filter the returns with the following flags:

$ gaiacli query gov proposals --depositor cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ gaiacli query gov proposals --voter cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ gaiacli query gov proposals --status (DepositPeriod|VotingPeriod|Passed|Rejected)
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			bechDepositorAddr := viper.GetString(flagDepositor)
			bechVoterAddr := viper.GetString(flagVoter)
			strProposalStatus := viper.GetString(flagStatus)
			numLimit := uint64(viper.GetInt64(flagNumLimit))

			var depositorAddr sdk.AccAddress
			var voterAddr sdk.AccAddress
			var proposalStatus gov.ProposalStatus

			params := gov.NewQueryProposalsParams(proposalStatus, numLimit, voterAddr, depositorAddr)

			if len(bechDepositorAddr) != 0 {
				depositorAddr, err := sdk.AccAddressFromBech32(bechDepositorAddr)
				if err != nil {
					return err
				}
				params.Depositor = depositorAddr
			}

			if len(bechVoterAddr) != 0 {
				voterAddr, err := sdk.AccAddressFromBech32(bechVoterAddr)
				if err != nil {
					return err
				}
				params.Voter = voterAddr
			}

			if len(strProposalStatus) != 0 {
				proposalStatus, err := gov.ProposalStatusFromString(govClientUtils.NormalizeProposalStatus(strProposalStatus))
				if err != nil {
					return err
				}
				params.ProposalStatus = proposalStatus
			}

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/proposals", queryRoute), bz)
			if err != nil {
				return err
			}

			var matchingProposals []gov.Proposal
			err = cdc.UnmarshalJSON(res, &matchingProposals)
			if err != nil {
				return err
			}

			if len(matchingProposals) == 0 {
				fmt.Println("No matching proposals found")
				return nil
			}

			for _, proposal := range matchingProposals {
				fmt.Printf("  %d - %s\n", proposal.GetProposalID(), proposal.GetTitle())
			}

			return nil
		},
	}

	cmd.Flags().String(flagNumLimit, "", "(optional) limit to latest [number] proposals. Defaults to all proposals")
	cmd.Flags().String(flagDepositor, "", "(optional) filter by proposals deposited on by depositor")
	cmd.Flags().String(flagVoter, "", "(optional) filter by proposals voted on by voted")
	cmd.Flags().String(flagStatus, "", "(optional) filter proposals by proposal status, status: deposit_period/voting_period/passed/rejected")

	return cmd
}

// Command to Get a Proposal Information
// GetCmdQueryVote implements the query proposal vote command.
func GetCmdQueryVote(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [voter-address]",
		Args:  cobra.ExactArgs(2),
		Short: "Query details of a single vote",
		Long: strings.TrimSpace(`
Query details for a single vote on a proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli query gov vote 1 cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = queryProposal(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return fmt.Errorf("Failed to fetch proposal-id %d: %s", proposalID, err)
			}

			// get voter address
			voterAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			// Construct query
			params := gov.NewQueryVoteParams(proposalID, voterAddr)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Query store
			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/vote", queryRoute), bz)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}

	return cmd
}

// GetCmdQueryVotes implements the command to query for proposal votes.
func GetCmdQueryVotes(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "votes [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query votes on a proposal",
		Long: strings.TrimSpace(`
Query vote details for a single proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli query gov votes 1
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = queryProposal(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return fmt.Errorf("Failed to fetch proposal-id %d: %s", proposalID, err)
			}

			// Construct query
			params := gov.NewQueryProposalParams(proposalID)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Query store
			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/votes", queryRoute), bz)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}

	return cmd
}

// Command to Get a specific Deposit Information
// GetCmdQueryDeposit implements the query proposal deposit command.
func GetCmdQueryDeposit(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [proposal-id] [depositer-address]",
		Args:  cobra.ExactArgs(2),
		Short: "Query details of a deposit",
		Long: strings.TrimSpace(`
Query details for a single proposal deposit on a proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli query gov deposit 1 cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = queryProposal(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return fmt.Errorf("Failed to fetch proposal-id %d: %s", proposalID, err)
			}

			// Get the depositer address
			depositorAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			// Construct query
			params := gov.NewQueryDepositParams(proposalID, depositorAddr)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Query store
			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/deposit", queryRoute), bz)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}

	return cmd
}

// GetCmdQueryDeposits implements the command to query for proposal deposits.
func GetCmdQueryDeposits(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query deposits on a proposal",
		Long: strings.TrimSpace(`
Query details for all deposits on a proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli query gov deposits 1
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = queryProposal(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return fmt.Errorf("Failed to fetch proposal-id %d: %s", proposalID, err)
			}

			// Construct query
			params := gov.NewQueryProposalParams(proposalID)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Query store
			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/deposits", queryRoute), bz)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}

	return cmd
}

// GetCmdQueryTally implements the command to query for proposal tally result.
func GetCmdQueryTally(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tally [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Get the tally of a proposal vote",
		Long: strings.TrimSpace(`
Query tally of votes on a proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli query gov tally 1
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = queryProposal(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return fmt.Errorf("Failed to fetch proposal-id %d: %s", proposalID, err)
			}

			// Construct query
			params := gov.NewQueryProposalParams(proposalID)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Query store
			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/tally", queryRoute), bz)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}

	return cmd
}

// GetCmdQueryProposal implements the query proposal command.
func GetCmdQueryParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "param [param-type]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the parameters (voting|tallying|deposit) of the governance process",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query store
			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/params/%s", queryRoute, args[0]), nil)
			if err != nil {
				return err
			}

			fmt.Println(string(res))
			return nil
		},
	}

	return cmd
}
