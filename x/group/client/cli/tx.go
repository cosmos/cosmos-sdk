package cli

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/group"
)

const (
	FlagExec               = "exec"
	ExecTry                = "try"
	FlagGroupPolicyAsAdmin = "group-policy-as-admin"
)

// TxCmd returns a root CLI command handler for all x/group transaction commands.
func TxCmd(name string) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        name,
		Short:                      "Group transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		MsgCreateGroupCmd(),
		MsgUpdateGroupAdminCmd(),
		MsgUpdateGroupMetadataCmd(),
		MsgUpdateGroupMembersCmd(),
		MsgCreateGroupWithPolicyCmd(),
		MsgCreateGroupPolicyCmd(),
		MsgUpdateGroupPolicyAdminCmd(),
		MsgUpdateGroupPolicyDecisionPolicyCmd(),
		MsgUpdateGroupPolicyMetadataCmd(),
		MsgSubmitProposalCmd(),
		MsgVoteCmd(),
		MsgExecCmd(),
	)

	return txCmd
}

// MsgCreateGroupCmd creates a CLI command for Msg/CreateGroup.
func MsgCreateGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create-group [admin] [metadata] [members-json-file]",
		Short: "Create a group which is an aggregation " +
			"of member accounts with associated weights and " +
			"an administrator account. Note, the '--from' flag is " +
			"ignored as it is implied from [admin].",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a group which is an aggregation of member accounts with associated weights and
an administrator account. Note, the '--from' flag is ignored as it is implied from [admin].
Members accounts can be given through a members JSON file that contains an array of members.

Example:
$ %s tx group create-group [admin] [metadata] [members-json-file]

Where members.json contains:

{
	"members": [
		{
			"address": "addr1",
			"weight": "1",
			"metadata": "some metadata"
		},
		{
			"address": "addr2",
			"weight": "1",
			"metadata": "some metadata"
		}
	]
}
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			members, err := parseMembers(clientCtx, args[2])
			if err != nil {
				return err
			}

			metadata, err := base64.StdEncoding.DecodeString(args[1])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg := &group.MsgCreateGroup{
				Admin:    clientCtx.GetFromAddress().String(),
				Members:  members,
				Metadata: metadata,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupMembersCmd creates a CLI command for Msg/UpdateGroupMembers.
func MsgUpdateGroupMembersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-members [admin] [group-id] [members-json-file]",
		Short: "Update a group's members. Set a member's weight to \"0\" to delete it.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update a group's members

Example:
$ %s tx group update-group-members [admin] [group-id] [members-json-file]

Where members.json contains:

{
	"members": [
		{
			"address": "addr1",
			"weight": "1",
			"metadata": "some new metadata"
		},
		{
			"address": "addr2",
			"weight": "0",
			"metadata": "some metadata"
		}
	]
}

Set a member's weight to "0" to delete it.
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			members, err := parseMembers(clientCtx, args[2])
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := &group.MsgUpdateGroupMembers{
				Admin:         clientCtx.GetFromAddress().String(),
				MemberUpdates: members,
				GroupId:       groupID,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupAdminCmd creates a CLI command for Msg/UpdateGroupAdmin.
func MsgUpdateGroupAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-admin [admin] [group-id] [new-admin]",
		Short: "Update a group's admin",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := &group.MsgUpdateGroupAdmin{
				Admin:    clientCtx.GetFromAddress().String(),
				NewAdmin: args[2],
				GroupId:  groupID,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupMetadataCmd creates a CLI command for Msg/UpdateGroupMetadata.
func MsgUpdateGroupMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-metadata [admin] [group-id] [metadata]",
		Short: "Update a group's metadata",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg := &group.MsgUpdateGroupMetadata{
				Admin:    clientCtx.GetFromAddress().String(),
				Metadata: b,
				GroupId:  groupID,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgCreateGroupWithPolicyCmd creates a CLI command for Msg/CreateGroupWithPolicy.
func MsgCreateGroupWithPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create-group-with-policy [admin] [group-metadata] [group-policy-metadata] [members-json-file] [decision-policy]",
		Short: "Create a group with policy which is an aggregation " +
			"of member accounts with associated weights, " +
			"an administrator account and a decision policy. Note, the '--from' flag is " +
			"ignored as it is implied from [admin].",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a group with policy which is an aggregation of member accounts with associated weights,
an administrator account and decision policy. Note, the '--from' flag is ignored as it is implied from [admin].
Members accounts can be given through a members JSON file that contains an array of members.
If group-policy-as-admin flag is set to true, the admin of the newly created group and group policy is set with the group policy address itself.

Example:
$ %s tx group create-group-with-policy [admin] [group-metadata] [group-policy-metadata] [members-json-file] \
'{"@type":"/cosmos.group.v1beta1.ThresholdDecisionPolicy", "threshold":"1", "timeout":"1s"}'

where members.json contains:

{
	"members": [
		{
			"address": "addr1",
			"weight": "1",
			"metadata": "some metadata"
		},
		{
			"address": "addr2",
			"weight": "1",
			"metadata": "some metadata"
		}
	]
}
`,
				version.AppName,
			),
		),
		Args: cobra.MinimumNArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupPolicyAsAdmin, err := cmd.Flags().GetBool(FlagGroupPolicyAsAdmin)
			if err != nil {
				return err
			}

			members, err := parseMembers(clientCtx, args[3])
			if err != nil {
				return err
			}

			groupMetadata, err := base64.StdEncoding.DecodeString(args[1])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "group metadata is malformed, proper base64 string is required")
			}

			groupPolicyMetadata, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "group policy metadata is malformed, proper base64 string is required")
			}

			var policy group.DecisionPolicy
			if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(args[4]), &policy); err != nil {
				return err
			}

			msg, err := group.NewMsgCreateGroupWithPolicy(
				clientCtx.GetFromAddress().String(),
				members,
				groupMetadata,
				groupPolicyMetadata,
				groupPolicyAsAdmin,
				policy,
			)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().Bool(FlagGroupPolicyAsAdmin, false, "Sets admin of the newly created group and group policy with group policy address itself when true")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgCreateGroupPolicyCmd creates a CLI command for Msg/CreateGroupPolicy.
func MsgCreateGroupPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create-group-policy [admin] [group-id] [metadata] [decision-policy]",
		Short: "Create a group policy which is an account " +
			"associated with a group and a decision policy. " +
			"Note, the '--from' flag is " +
			"ignored as it is implied from [admin].",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a group policy which is an account associated with a group and a decision policy.
Note, the '--from' flag is ignored as it is implied from [admin].

Example:
$ %s tx group create-group-policy [admin] [group-id] [metadata] \
'{"@type":"/cosmos.group.v1beta1.ThresholdDecisionPolicy", "threshold":"1", "timeout":"1s"}'

Here, we can use percentage decision policy when needed, where 0 < percentage <= 1.
Ex: '{"@type":"/cosmos.group.v1beta1.PercentageDecisionPolicy", "percentage":"0.5", "timeout":"1s"}'
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			var policy group.DecisionPolicy
			if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(args[3]), &policy); err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg, err := group.NewMsgCreateGroupPolicy(
				clientCtx.GetFromAddress(),
				groupID,
				b,
				policy,
			)
			if err != nil {
				return err
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupPolicyAdminCmd creates a CLI command for Msg/UpdateGroupPolicyAdmin.
func MsgUpdateGroupPolicyAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-policy-admin [admin] [group-policy-account] [new-admin]",
		Short: "Update a group policy admin",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &group.MsgUpdateGroupPolicyAdmin{
				Admin:    clientCtx.GetFromAddress().String(),
				Address:  args[1],
				NewAdmin: args[2],
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupPolicyDecisionPolicyCmd creates a CLI command for Msg/UpdateGroupPolicyDecisionPolicy.
func MsgUpdateGroupPolicyDecisionPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-policy-decision-policy [admin] [group-policy-account] [decision-policy]",
		Short: "Update a group policy's decision policy",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var policy group.DecisionPolicy
			if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(args[2]), &policy); err != nil {
				return err
			}

			accountAddress, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg, err := group.NewMsgUpdateGroupPolicyDecisionPolicyRequest(
				clientCtx.GetFromAddress(),
				accountAddress,
				policy,
			)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupPolicyMetadataCmd creates a CLI command for Msg/MsgUpdateGroupPolicyMetadata.
func MsgUpdateGroupPolicyMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-policy-metadata [admin] [group-policy-account] [new-metadata]",
		Short: "Update a group policy metadata",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg := &group.MsgUpdateGroupPolicyMetadata{
				Admin:    clientCtx.GetFromAddress().String(),
				Address:  args[1],
				Metadata: b,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgSubmitProposalCmd creates a CLI command for Msg/SubmitProposal.
func MsgSubmitProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal [proposal_json_file]",
		Short: "Submit a new proposal",
		Long: fmt.Sprintf(`Submit a new proposal.

Parameters:
			msg_tx_json_file: path to json file with messages that will be executed if the proposal is accepted.

Example:
	$ %s tx gov submit-proposal path/to/proposal.json
	
	Where proposal.json contains:

{
	"group_policy_address": "cosmos1...",
	// array of proto-JSON-encoded sdk.Msgs
	"messages": [
	{
		"@type": "/cosmos.bank.v1beta1.MsgSend",
		"from_address": "cosmos1...",
		"to_address": "cosmos1...",
		"amount":[{"denom": "stake","amount": "10"}]
	}
	],
	"metadata: "4pIMOgIGx1vZGU=", // base64-encoded metadata
	"proposers": ["cosmos1...", "cosmos1..."],
}`, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prop, err := parseCLIProposal(args[0])
			if err != nil {
				return err
			}

			// Since the --from flag is not required on this CLI command, we
			// ignore it, and just use the 1st proposer in the JSON file.
			cmd.Flags().Set(flags.FlagFrom, prop.Proposers[0])

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msgs, err := parseMsgs(clientCtx.Codec, prop)
			if err != nil {
				return err
			}

			execStr, _ := cmd.Flags().GetString(FlagExec)

			msg, err := group.NewMsgSubmitProposalRequest(
				prop.GroupPolicyAddress,
				prop.Proposers,
				msgs,
				prop.Metadata,
				execFromString(execStr),
			)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagExec, "", "Set to 1 to try to execute proposal immediately after creation (proposers signatures are considered as Yes votes)")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgWithdrawProposalCmd creates a CLI command for Msg/WithdrawProposal.
func MsgWithdrawProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-proposal [proposal-id] [group-policy-admin-or-proposer]",
		Short: "Withdraw a submitted proposal",
		Long: `Withdraw a submitted proposal.

Parameters:
			proposal-id: unique ID of the proposal.
			group-policy-admin-or-proposer: either admin of the group policy or one the proposer of the proposal.
			(note: --from flag will be ignored here)
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := &group.MsgWithdrawProposal{
				ProposalId: proposalID,
				Address:    clientCtx.GetFromAddress().String(),
			}

			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgVoteCmd creates a CLI command for Msg/Vote.
func MsgVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [voter] [vote-option] [metadata]",
		Short: "Vote on a proposal",
		Long: `Vote on a proposal.

Parameters:
			proposal-id: unique ID of the proposal
			voter: voter account addresses.
			vote-option: choice of the voter(s)
				VOTE_OPTION_UNSPECIFIED: no-op
				VOTE_OPTION_NO: no
				VOTE_OPTION_YES: yes
				VOTE_OPTION_ABSTAIN: abstain
				VOTE_OPTION_NO_WITH_VETO: no-with-veto
			Metadata: metadata for the vote
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			voteOption, err := group.VoteOptionFromString(args[2])
			if err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[3])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			execStr, _ := cmd.Flags().GetString(FlagExec)

			msg := &group.MsgVote{
				ProposalId: proposalID,
				Voter:      args[1],
				Option:     voteOption,
				Metadata:   b,
				Exec:       execFromString(execStr),
			}
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagExec, "", "Set to 1 to try to execute proposal immediately after voting")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgExecCmd creates a CLI command for Msg/MsgExec.
func MsgExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [proposal-id]",
		Short: "Execute a proposal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := &group.MsgExec{
				ProposalId: proposalID,
				Signer:     clientCtx.GetFromAddress().String(),
			}
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
