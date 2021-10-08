package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	govutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Proposal flags
const (
	FlagDeposit   = "deposit"
	flagVoter     = "voter"
	flagDepositor = "depositor"
	flagStatus    = "status"
)

type proposal struct {
	Title       string
	Description string
	Type        string
	Deposit     string
}

// ProposalFlags defines the core required fields of a proposal. It is used to
// verify that these values are not provided in conjunction with a JSON proposal
// file.
var ProposalFlags = []string{
	FlagDeposit,
}

// NewTxCmd returns the transaction commands for this module
// governance ModuleClient is slightly different from other ModuleClients in that
// it contains a slice of "proposal" child commands. These commands are respective
// to proposal type handlers that are implemented in other modules but are mounted
// under the governance CLI (eg. parameter change proposals).
func NewTxCmd(propCmds []*cobra.Command) *cobra.Command {
	govTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Governance transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmdSubmitProp := NewCmdSubmitProposal()
	for _, propCmd := range propCmds {
		flags.AddTxFlagsToCmd(propCmd)
		cmdSubmitProp.AddCommand(propCmd)
	}

	govTxCmd.AddCommand(
		NewCmdDeposit(),
		NewCmdVote(),
		NewCmdWeightedVote(),
		cmdSubmitProp,
	)

	return govTxCmd
}

// NewCmdSubmitProposal implements submitting a proposal transaction command.
func NewCmdSubmitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal [path/to/messages.json]",
		Short: "Submit a proposal along with an initial deposit",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a proposal with an array of messages along with an initial deposit.

Example:
$ %s tx gov submit-proposal path/to/messages.json --from mykey --deposit="10test"
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalMsgs, err := authclient.ReadTxFromFile(clientCtx, args[0])
			if err != nil {
				return err
			}

			depositString, err := cmd.Flags().GetString(FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositString)
			if err != nil {
				return err
			}

			msg, err := types.NewMsgSubmitProposal(proposalMsgs.GetMsgs(), deposit, clientCtx.GetFromAddress())
			if err != nil {
				return fmt.Errorf("invalid message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagDeposit, "", "The proposal deposit")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCmdDeposit implements depositing tokens for an active proposal.
func NewCmdDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [proposal-id] [deposit]",
		Args:  cobra.ExactArgs(2),
		Short: "Deposit tokens for an active proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a deposit for an active proposal. You can
find the proposal-id by running "%s query gov proposals".

Example:
$ %s tx gov deposit 1 10stake --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// Get depositor address
			from := clientCtx.GetFromAddress()

			// Get amount of coins
			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgDeposit(from, proposalID, amount)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCmdVote implements creating a new vote command.
func NewCmdVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [option]",
		Args:  cobra.ExactArgs(2),
		Short: "Vote for an active proposal, options: yes/no/no_with_veto/abstain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a vote for an active proposal. You can
find the proposal-id by running "%s query gov proposals".

Example:
$ %s tx gov vote 1 yes --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// Get voting address
			from := clientCtx.GetFromAddress()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// Find out which vote option user chose
			byteVoteOption, err := types.VoteOptionFromString(govutils.NormalizeVoteOption(args[1]))
			if err != nil {
				return err
			}

			// Build vote message and run basic validation
			msg := types.NewMsgVote(from, proposalID, byteVoteOption)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCmdWeightedVote implements creating a new weighted vote command.
func NewCmdWeightedVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "weighted-vote [proposal-id] [weighted-options]",
		Args:  cobra.ExactArgs(2),
		Short: "Vote for an active proposal, options: yes/no/no_with_veto/abstain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a vote for an active proposal. You can
find the proposal-id by running "%s query gov proposals".

Example:
$ %s tx gov weighted-vote 1 yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05 --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Get voter address
			from := clientCtx.GetFromAddress()

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// Figure out which vote options user chose
			options, err := types.WeightedVoteOptionsFromString(govutils.NormalizeWeightedVoteOptions(args[1]))
			if err != nil {
				return err
			}

			// Build vote message and run basic validation
			msg := types.NewMsgVoteWeighted(from, proposalID, options)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
