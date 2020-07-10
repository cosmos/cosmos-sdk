package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	gcutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group gov queries under a subcommand
	govQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the governance module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	govQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryProposal(),
		GetCmdQueryProposals(),
		GetCmdQueryVote(),
		GetCmdQueryVotes(),
		GetCmdQueryParam(),
		GetCmdQueryParams(),
		GetCmdQueryProposer(),
		GetCmdQueryDeposit(),
		GetCmdQueryDeposits(),
		GetCmdQueryTally(),
	)...)

	return govQueryCmd
}

// GetCmdQueryProposal implements the query proposal command.
func GetCmdQueryProposal() *cobra.Command {
	return &cobra.Command{
		Use:   "proposal [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query details of a single proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for a proposal. You can find the
proposal-id by running "%s query gov proposals".

Example:
$ %s query gov proposal 1
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// Query the proposal
			res, err := gcutils.QueryProposalByID(proposalID, clientCtx, types.QuerierRoute)
			if err != nil {
				return err
			}

			var proposal types.Proposal
			clientCtx.JSONMarshaler.MustUnmarshalJSON(res, &proposal)
			return clientCtx.PrintOutput(proposal) // nolint:errcheck
		},
	}
}

// GetCmdQueryProposals implements a query proposals command.
func GetCmdQueryProposals() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposals",
		Short: "Query proposals with optional filters",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for a all paginated proposals that match optional filters:

Example:
$ %s query gov proposals --depositor cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ %s query gov proposals --voter cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ %s query gov proposals --status (DepositPeriod|VotingPeriod|Passed|Rejected)
$ %s query gov proposals --page=2 --limit=100
`,
				version.AppName, version.AppName, version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			bechDepositorAddr, _ := cmd.Flags().GetString(flagDepositor)
			bechVoterAddr, _ := cmd.Flags().GetString(flagVoter)
			strProposalStatus, _ := cmd.Flags().GetString(flagStatus)
			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			var depositorAddr sdk.AccAddress
			var voterAddr sdk.AccAddress
			var proposalStatus types.ProposalStatus

			params := types.NewQueryProposalsParams(page, limit, proposalStatus, voterAddr, depositorAddr)

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
				proposalStatus, err := types.ProposalStatusFromString(gcutils.NormalizeProposalStatus(strProposalStatus))
				if err != nil {
					return err
				}
				params.ProposalStatus = proposalStatus
			}

			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			bz, err := clientCtx.JSONMarshaler.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/proposals", types.QuerierRoute), bz)
			if err != nil {
				return err
			}

			var matchingProposals types.Proposals
			err = clientCtx.JSONMarshaler.UnmarshalJSON(res, &matchingProposals)
			if err != nil {
				return err
			}

			if len(matchingProposals) == 0 {
				return fmt.Errorf("no matching proposals found")
			}

			return clientCtx.PrintOutput(matchingProposals) // nolint:errcheck
		},
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of proposals to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of proposals to query for")
	cmd.Flags().String(flagDepositor, "", "(optional) filter by proposals deposited on by depositor")
	cmd.Flags().String(flagVoter, "", "(optional) filter by proposals voted on by voted")
	cmd.Flags().String(flagStatus, "", "(optional) filter proposals by proposal status, status: deposit_period/voting_period/passed/rejected")

	return cmd
}

// Command to Get a Proposal Information
// GetCmdQueryVote implements the query proposal vote command.
func GetCmdQueryVote() *cobra.Command {
	return &cobra.Command{
		Use:   "vote [proposal-id] [voter-addr]",
		Args:  cobra.ExactArgs(2),
		Short: "Query details of a single vote",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for a single vote on a proposal given its identifier.

Example:
$ %s query gov vote 1 cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = gcutils.QueryProposalByID(proposalID, clientCtx, types.QuerierRoute)
			if err != nil {
				return fmt.Errorf("failed to fetch proposal-id %d: %s", proposalID, err)
			}

			voterAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			params := types.NewQueryVoteParams(proposalID, voterAddr)
			bz, err := clientCtx.JSONMarshaler.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/vote", types.QuerierRoute), bz)
			if err != nil {
				return err
			}

			var vote types.Vote

			// XXX: Allow the decoding to potentially fail as the vote may have been
			// pruned from state. If so, decoding will fail and so we need to check the
			// Empty() case. Consider updating Vote JSON decoding to not fail when empty.
			_ = clientCtx.JSONMarshaler.UnmarshalJSON(res, &vote)

			if vote.Empty() {
				res, err = gcutils.QueryVoteByTxQuery(clientCtx, params)
				if err != nil {
					return err
				}

				if err := clientCtx.JSONMarshaler.UnmarshalJSON(res, &vote); err != nil {
					return err
				}
			}

			return clientCtx.PrintOutput(vote)
		},
	}
}

// GetCmdQueryVotes implements the command to query for proposal votes.
func GetCmdQueryVotes() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "votes [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query votes on a proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query vote details for a single proposal by its identifier.

Example:
$ %[1]s query gov votes 1
$ %[1]s query gov votes 1 --page=2 --limit=100
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			params := types.NewQueryProposalVotesParams(proposalID, page, limit)
			bz, err := clientCtx.JSONMarshaler.MarshalJSON(params)
			if err != nil {
				return err
			}

			// check to see if the proposal is in the store
			res, err := gcutils.QueryProposalByID(proposalID, clientCtx, types.QuerierRoute)
			if err != nil {
				return fmt.Errorf("failed to fetch proposal-id %d: %s", proposalID, err)
			}

			var proposal types.Proposal
			clientCtx.JSONMarshaler.MustUnmarshalJSON(res, &proposal)

			propStatus := proposal.Status
			if !(propStatus == types.StatusVotingPeriod || propStatus == types.StatusDepositPeriod) {
				res, err = gcutils.QueryVotesByTxQuery(clientCtx, params)
			} else {
				res, _, err = clientCtx.QueryWithData(fmt.Sprintf("custom/%s/votes", types.QuerierRoute), bz)
			}

			if err != nil {
				return err
			}

			var votes types.Votes
			clientCtx.JSONMarshaler.MustUnmarshalJSON(res, &votes)
			return clientCtx.PrintOutput(votes)
		},
	}
	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of votes to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of votes to query for")
	return cmd
}

// Command to Get a specific Deposit Information
// GetCmdQueryDeposit implements the query proposal deposit command.
func GetCmdQueryDeposit() *cobra.Command {
	return &cobra.Command{
		Use:   "deposit [proposal-id] [depositer-addr]",
		Args:  cobra.ExactArgs(2),
		Short: "Query details of a deposit",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for a single proposal deposit on a proposal by its identifier.

Example:
$ %s query gov deposit 1 cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = gcutils.QueryProposalByID(proposalID, clientCtx, types.QuerierRoute)
			if err != nil {
				return fmt.Errorf("failed to fetch proposal-id %d: %s", proposalID, err)
			}

			depositorAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			params := types.NewQueryDepositParams(proposalID, depositorAddr)
			bz, err := clientCtx.JSONMarshaler.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/deposit", types.QuerierRoute), bz)
			if err != nil {
				return err
			}

			var deposit types.Deposit
			clientCtx.JSONMarshaler.MustUnmarshalJSON(res, &deposit)

			if deposit.Empty() {
				res, err = gcutils.QueryDepositByTxQuery(clientCtx, params)
				if err != nil {
					return err
				}
				clientCtx.JSONMarshaler.MustUnmarshalJSON(res, &deposit)
			}

			return clientCtx.PrintOutput(deposit)
		},
	}
}

// GetCmdQueryDeposits implements the command to query for proposal deposits.
func GetCmdQueryDeposits() *cobra.Command {
	return &cobra.Command{
		Use:   "deposits [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query deposits on a proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for all deposits on a proposal.
You can find the proposal-id by running "%s query gov proposals".

Example:
$ %s query gov deposits 1
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			params := types.NewQueryProposalParams(proposalID)
			bz, err := clientCtx.JSONMarshaler.MarshalJSON(params)
			if err != nil {
				return err
			}

			// check to see if the proposal is in the store
			res, err := gcutils.QueryProposalByID(proposalID, clientCtx, types.QuerierRoute)
			if err != nil {
				return fmt.Errorf("failed to fetch proposal with id %d: %s", proposalID, err)
			}

			var proposal types.Proposal
			clientCtx.JSONMarshaler.MustUnmarshalJSON(res, &proposal)

			propStatus := proposal.Status
			if !(propStatus == types.StatusVotingPeriod || propStatus == types.StatusDepositPeriod) {
				res, err = gcutils.QueryDepositsByTxQuery(clientCtx, params)
			} else {
				res, _, err = clientCtx.QueryWithData(fmt.Sprintf("custom/%s/deposits", types.QuerierRoute), bz)
			}

			if err != nil {
				return err
			}

			var dep types.Deposits
			clientCtx.JSONMarshaler.MustUnmarshalJSON(res, &dep)
			return clientCtx.PrintOutput(dep)
		},
	}
}

// GetCmdQueryTally implements the command to query for proposal tally result.
func GetCmdQueryTally() *cobra.Command {
	return &cobra.Command{
		Use:   "tally [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Get the tally of a proposal vote",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query tally of votes on a proposal. You can find
the proposal-id by running "%s query gov proposals".

Example:
$ %s query gov tally 1
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = gcutils.QueryProposalByID(proposalID, clientCtx, types.QuerierRoute)
			if err != nil {
				return fmt.Errorf("failed to fetch proposal-id %d: %s", proposalID, err)
			}

			// Construct query
			params := types.NewQueryProposalParams(proposalID)
			bz, err := clientCtx.JSONMarshaler.MarshalJSON(params)
			if err != nil {
				return err
			}

			// Query store
			res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/tally", types.QuerierRoute), bz)
			if err != nil {
				return err
			}

			var tally types.TallyResult
			clientCtx.JSONMarshaler.MustUnmarshalJSON(res, &tally)
			return clientCtx.PrintOutput(tally)
		},
	}
}

// GetCmdQueryProposal implements the query proposal command.
func GetCmdQueryParams() *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Query the parameters of the governance process",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the all the parameters for the governance process.

Example:
$ %s query gov params
`,
				version.AppName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			tp, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/params/tallying", types.QuerierRoute), nil)
			if err != nil {
				return err
			}
			dp, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/params/deposit", types.QuerierRoute), nil)
			if err != nil {
				return err
			}
			vp, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/params/voting", types.QuerierRoute), nil)
			if err != nil {
				return err
			}

			var tallyParams types.TallyParams
			clientCtx.JSONMarshaler.MustUnmarshalJSON(tp, &tallyParams)
			var depositParams types.DepositParams
			clientCtx.JSONMarshaler.MustUnmarshalJSON(dp, &depositParams)
			var votingParams types.VotingParams
			clientCtx.JSONMarshaler.MustUnmarshalJSON(vp, &votingParams)

			return clientCtx.PrintOutput(types.NewParams(votingParams, tallyParams, depositParams))
		},
	}
}

// GetCmdQueryProposal implements the query proposal command.
func GetCmdQueryParam() *cobra.Command {
	return &cobra.Command{
		Use:   "param [param-type]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the parameters (voting|tallying|deposit) of the governance process",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the all the parameters for the governance process.

Example:
$ %s query gov param voting
$ %s query gov param tallying
$ %s query gov param deposit
`,
				version.AppName, version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			// Query store
			res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/params/%s", types.QuerierRoute, args[0]), nil)
			if err != nil {
				return err
			}
			var out fmt.Stringer
			switch args[0] {
			case "voting":
				var param types.VotingParams
				clientCtx.JSONMarshaler.MustUnmarshalJSON(res, &param)
				out = param
			case "tallying":
				var param types.TallyParams
				clientCtx.JSONMarshaler.MustUnmarshalJSON(res, &param)
				out = param
			case "deposit":
				var param types.DepositParams
				clientCtx.JSONMarshaler.MustUnmarshalJSON(res, &param)
				out = param
			default:
				return fmt.Errorf("argument must be one of (voting|tallying|deposit), was %s", args[0])
			}

			return clientCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryProposer implements the query proposer command.
func GetCmdQueryProposer() *cobra.Command {
	return &cobra.Command{
		Use:   "proposer [proposal-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the proposer of a governance proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query which address proposed a proposal with a given ID.

Example:
$ %s query gov proposer 1
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			// validate that the proposalID is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s is not a valid uint", args[0])
			}

			prop, err := gcutils.QueryProposerByTxQuery(clientCtx, proposalID)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(prop)
		},
	}
}

// DONTCOVER
