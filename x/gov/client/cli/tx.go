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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// Proposal flags
const (
	FlagTitle     = "title"
	FlagDeposit   = "deposit"
	FlagMetadata  = "metadata"
	FlagSummary   = "summary"
	FlagExpedited = "expedited"

	// Deprecated: only used for v1beta1 legacy proposals.
	FlagProposal = "proposal"
	// Deprecated: only used for v1beta1 legacy proposals.
	FlagDescription = "description"
	// Deprecated: only used for v1beta1 legacy proposals.
	FlagProposalType = "type"
)

// ProposalFlags defines the core required fields of a legacy proposal. It is used to
// verify that these values are not provided in conjunction with a JSON proposal
// file.
var ProposalFlags = []string{
	FlagTitle,
	FlagDescription,
	FlagProposalType,
	FlagDeposit,
}

// NewTxCmd returns the transaction commands for this module
// governance ModuleClient is slightly different from other ModuleClients in that
// it contains a slice of legacy "proposal" child commands. These commands are respective
// to the proposal type handlers that are implemented in other modules but are mounted
// under the governance CLI (eg. parameter change proposals).
func NewTxCmd(legacyPropCmds []*cobra.Command) *cobra.Command {
	govTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Governance transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmdSubmitLegacyProp := NewCmdSubmitLegacyProposal()
	for _, propCmd := range legacyPropCmds {
		flags.AddTxFlagsToCmd(propCmd)
		cmdSubmitLegacyProp.AddCommand(propCmd)
	}

	govTxCmd.AddCommand(
		NewCmdDeposit(),
		NewCmdVote(),
		NewCmdWeightedVote(),
		NewCmdSubmitProposal(),
		NewCmdDraftProposal(),
		NewCmdCancelProposal(),
		NewCmdGenerateConstitutionAmendment(),
		CreateGovernorCmd(),
		EditGovernorCmd(),
		UpdateGovernorStatusCmd(),
		DelegateGovernorCmd(),
		UndelegateGovernorCmd(),

		// Deprecated
		cmdSubmitLegacyProp,
	)

	return govTxCmd
}

// NewCmdSubmitProposal implements submitting a proposal transaction command.
func NewCmdSubmitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal [path/to/proposal.json]",
		Short: "Submit a proposal along with some messages, metadata and deposit",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a proposal along with some messages, metadata and deposit.
They should be defined in a JSON file.

Example:
$ %s tx gov submit-proposal path/to/proposal.json

Where proposal.json contains:

{
  // array of proto-JSON-encoded sdk.Msgs
  "messages": [
    {
      "@type": "/cosmos.bank.v1beta1.MsgSend",
      "from_address": "cosmos1...",
      "to_address": "cosmos1...",
      "amount":[{"denom": "stake","amount": "10"}]
    }
  ],
  // metadata can be any of base64 encoded, raw text, stringified json, IPFS link to json
  // see below for example metadata
  "metadata": "4pIMOgIGx1vZGU=",
  "deposit": "10stake",
  "title": "My proposal",
  "summary": "A short summary of my proposal",
  "expedited": false
}

metadata example:
{
	"title": "",
	"authors": [""],
	"summary": "",
	"details": "",
	"proposal_forum_url": "",
	"vote_option_context": "",
}
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, msgs, deposit, err := parseSubmitProposal(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			msg, err := v1.NewMsgSubmitProposal(msgs, deposit, clientCtx.GetFromAddress().String(), proposal.Metadata, proposal.Title, proposal.Summary)
			if err != nil {
				return fmt.Errorf("invalid message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCmdCancelProposal implements submitting a cancel proposal transaction command.
func NewCmdCancelProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel-proposal [proposal-id]",
		Short:   "Cancel governance proposal before the voting period ends. Must be signed by the proposal creator.",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf(`$ %s tx gov cancel-proposal 1 --from mykey`, version.AppName),
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

			// Get proposer address
			from := clientCtx.GetFromAddress()
			msg := v1.NewMsgCancelProposal(proposalID, from.String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCmdSubmitLegacyProposal implements submitting a proposal transaction command.
// Deprecated: please use NewCmdSubmitProposal instead.
func NewCmdSubmitLegacyProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-legacy-proposal",
		Short: "Submit a legacy proposal along with an initial deposit",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a legacy proposal along with an initial deposit.
Proposal title, description, type and deposit can be given directly or through a proposal JSON file.

Example:
$ %s tx gov submit-legacy-proposal --proposal="path/to/proposal.json" --from mykey

Where proposal.json contains:

{
  "title": "Test Proposal",
  "description": "My awesome proposal",
  "type": "Text",
  "deposit": "10test"
}

Which is equivalent to:

$ %s tx gov submit-legacy-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --deposit="10test" --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := parseSubmitLegacyProposal(cmd.Flags())
			if err != nil {
				return fmt.Errorf("failed to parse proposal: %w", err)
			}

			amount, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			content, ok := v1beta1.ContentFromProposalType(proposal.Title, proposal.Description, proposal.Type)
			if !ok {
				return fmt.Errorf("failed to create proposal content: unknown proposal type %s", proposal.Type)
			}

			msg, err := v1beta1.NewMsgSubmitProposal(content, amount, clientCtx.GetFromAddress())
			if err != nil {
				return fmt.Errorf("invalid message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagTitle, "", "The proposal title")
	cmd.Flags().String(FlagDescription, "", "The proposal description")
	cmd.Flags().String(FlagProposalType, "", "The proposal Type")
	cmd.Flags().String(FlagDeposit, "", "The proposal deposit")
	cmd.Flags().String(FlagProposal, "", "Proposal file path (if this path is given, other proposal flags are ignored)")
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

			msg := v1.NewMsgDeposit(from, proposalID, amount)

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
		Short: "Vote for an active proposal, options: yes/no/abstain",
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
			byteVoteOption, err := v1.VoteOptionFromString(govutils.NormalizeVoteOption(args[1]))
			if err != nil {
				return err
			}

			metadata, err := cmd.Flags().GetString(FlagMetadata)
			if err != nil {
				return err
			}

			// Build vote message and run basic validation
			msg := v1.NewMsgVote(from, proposalID, byteVoteOption, metadata)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagMetadata, "", "Specify metadata of the vote")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCmdWeightedVote implements creating a new weighted vote command.
func NewCmdWeightedVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "weighted-vote [proposal-id] [weighted-options]",
		Args:  cobra.ExactArgs(2),
		Short: "Vote for an active proposal, options: yes/no/abstain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a vote for an active proposal. You can
find the proposal-id by running "%s query gov proposals".

Example:
$ %s tx gov weighted-vote 1 yes=0.6,no=0.3,abstain=0.05 --from mykey
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
			options, err := v1.WeightedVoteOptionsFromString(govutils.NormalizeWeightedVoteOptions(args[1]))
			if err != nil {
				return err
			}

			metadata, err := cmd.Flags().GetString(FlagMetadata)
			if err != nil {
				return err
			}

			// Build vote message and run basic validation
			msg := v1.NewMsgVoteWeighted(from, proposalID, options, metadata)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagMetadata, "", "Specify metadata of the weighted vote")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCmdConstitutionAmendmentMsg returns the command to generate the sdk.Msg
// required for a constitution amendment proposal generating the unified diff
// between the current constitution (queried) and the updated constitution
// from the provided markdown file.
func NewCmdGenerateConstitutionAmendment() *cobra.Command {
	flagCurrentConstitution := "current-constitution"

	cmd := &cobra.Command{
		Use:   "generate-constitution-amendment [path/to/updated/constitution.md]",
		Args:  cobra.ExactArgs(1),
		Short: "Generate a constitution amendment proposal message",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Generate a constitution amendment proposal message from the current
constitution and the provided updated constitution.
Queries the current constitution from the node (unless --current-constitution is used)
and generates a valid constitution amendment proposal message containing the unified diff
between the current constitution and the updated constitution provided
in a markdown file.

NOTE: this is just a utility command, it is not able to generate or
submit a valid Tx. Use the 'tx gov submit-proposal' command in
conjunction with the result of this one to submit the proposal.
See also 'tx gov draft-proposal' for a more general proposal drafting tool.

Example:
$ %s tx gov generate-constitution-amendment path/to/updated/constitution.md
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read the updated constitution from the provided markdown file
			updatedConstitution, err := readFromMarkdownFile(args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var currentConstitution string
			currentConstitutionPath, err := cmd.Flags().GetString(flagCurrentConstitution)
			if err != nil {
				return err
			}

			if currentConstitutionPath != "" {
				// Read the current constitution from the provided file
				currentConstitution, err = readFromMarkdownFile(currentConstitutionPath)
				if err != nil {
					return err
				}
			} else {
				// Query the current constitution from the node
				queryClient := v1.NewQueryClient(clientCtx)
				resp, err := queryClient.Constitution(cmd.Context(), &v1.QueryConstitutionRequest{})
				if err != nil {
					return err
				}
				currentConstitution = resp.Constitution
			}

			// Generate the unified diff between the current and updated constitutions
			diff, err := govutils.GenerateUnifiedDiff(currentConstitution, updatedConstitution)
			if err != nil {
				return err
			}

			msg := v1.NewMsgProposeConstitutionAmendment(authtypes.NewModuleAddress(types.ModuleName), diff)
			return clientCtx.PrintProto(msg)
		},
	}

	// This is not a tx command (but a utility for the proposal tx), so we don't need to add tx flags.
	// It might actually be confusing, so we just add the query flags.
	flags.AddQueryFlagsToCmd(cmd)

	// query commands have the FlagOutput default to "text", but we want to override it to "json"
	// in this case.
	cmd.Flags().Lookup(flags.FlagOutput).DefValue = flags.OutputFormatJSON
	err := cmd.Flags().Set(flags.FlagOutput, flags.OutputFormatJSON)
	if err != nil {
		panic(err)
	}

	// add flag to pass input constitution file instead of querying node
	// for the current constitution
	cmd.Flags().String(flagCurrentConstitution, "", "Path to the current constitution markdown file (optional, if not provided, the current constitution will be queried from the node)")

	return cmd
}

// CreateGovernorCmd creates a new Governor
func CreateGovernorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-governor [address] [moniker] [identity] [website] [security-contact] [details]",
		Short: "Create a new Governor",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			description := v1.GovernorDescription{
				Moniker:         args[1],
				Identity:        args[2],
				Website:         args[3],
				SecurityContact: args[4],
				Details:         args[5],
			}

			msg := v1.NewMsgCreateGovernor(address, description)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// EditGovernorCmd edits a Governor
func EditGovernorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-governor [address] [moniker] [identity] [website] [security-contact] [details]",
		Short: "Edit a Governor.",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			description := v1.GovernorDescription{
				Moniker:         args[1],
				Identity:        args[2],
				Website:         args[3],
				SecurityContact: args[4],
				Details:         args[5],
			}

			msg := v1.NewMsgEditGovernor(address, description)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// UpdateGovernorStatusCmd updates the status of a Governor
func UpdateGovernorStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-governor-status [address] [status]",
		Short: "Update the status of a Governor",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			status, err := v1.GovernorStatusFromString(args[1])
			if err != nil {
				return err
			}

			msg := v1.NewMsgUpdateGovernorStatus(address, status)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// DelegateGovernorCmd delegates or redelegates to a Governor
func DelegateGovernorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate-governor [delegator-address] [governor-address]",
		Short: "Delegate governance power to a Governor. Triggers a redelegation if a governance delegation already exists",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			delegatorAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			governorAddress, err := types.GovernorAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := v1.NewMsgDelegateGovernor(delegatorAddress, governorAddress)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// UndelegateGovernorCmd undelegates from a Governor
func UndelegateGovernorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelegate-governor [delegator-address]",
		Short: "Undelegate tokens from a Governor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			delegatorAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := v1.NewMsgUndelegateGovernor(delegatorAddress)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
