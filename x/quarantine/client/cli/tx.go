package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

const (
	// FlagPermanent is the flag indicating a permanent accept/decline.
	FlagPermanent = "permanent"
)

// exampleTxCmdBase is the base command that gets a user to one of the tx commands in here.
var exampleTxCmdBase = fmt.Sprintf("%s tx %s", version.AppName, quarantine.ModuleName)

// TxCmd returns the command with sub-commands for specific quarantine module Tx interaction.
func TxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        quarantine.ModuleName,
		Short:                      "Quarantine transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		TxOptInCmd(),
		TxOptOutCmd(),
		TxAcceptCmd(),
		TxDeclineCmd(),
		TxUpdateAutoResponsesCmd(),
	)

	return txCmd
}

// TxOptInCmd returns the command for executing an OptIn Tx.
func TxOptInCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "opt-in [<to_name_or_address>]",
		Short: "Activate quarantine for an account",
		Long: `Activate quarantine for an account.
Note, the '--from' flag is ignored as it is implied from [to_name_or_address] (the signer of the message).`,
		Example: fmt.Sprintf(`
$ %[1]s opt-in %[2]s
$ %[1]s opt-in personal
$ %[1]s opt-in --from %[2]s
$ %[1]s opt-in --from personal
`,
			exampleTxCmdBase, exampleAddr1),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args[0]) == 0 {
				return fmt.Errorf("no to_name_or_address provided")
			}

			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := quarantine.NewMsgOptIn(clientCtx.GetFromAddress())
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// TxOptOutCmd returns the command for executing an OptOut Tx.
func TxOptOutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "opt-out [<to_name_or_address>]",
		Short: "Deactivate quarantine for an account",
		Long: `Deactivate quarantine for an account.
Note, the '--from' flag is ignored as it is implied from [to_name_or_address] (the signer of the message).`,
		Example: fmt.Sprintf(`
$ %[1]s opt-out %[2]s
$ %[1]s opt-out personal
$ %[1]s opt-out --from %[2]s
$ %[1]s opt-out --from personal
`,
			exampleTxCmdBase, exampleAddr1),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args[0]) == 0 {
				return fmt.Errorf("no to_name_or_address provided")
			}

			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := quarantine.NewMsgOptOut(clientCtx.GetFromAddress())
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// TxAcceptCmd returns the command for executing an Accept Tx.
func TxAcceptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accept <to_name_or_address> <from_address> [<from_address 2> ...]",
		Short: "Accept quarantined funds sent to <to_name_or_address> from <from_address>",
		Long: `Accept quarantined funds sent to <to_name_or_address> from <from_address>.
Note, the '--from' flag is ignored as it is implied from [to_name_or_address] (the signer of the message).`,
		Example: fmt.Sprintf(`
$ %[1]s accept %[2]s %[3]s
$ %[1]s accept personal %[3]s
$ %[1]s accept personal %[3]s %[4]s
`,
			exampleTxCmdBase, exampleAddr1, exampleAddr2, exampleAddr3),
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args[0]) == 0 {
				return fmt.Errorf("no to_name_or_address provided")
			}
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			toAddr := clientCtx.GetFromAddress()

			fromAddrsStrs := make([]string, len(args)-1)
			for i, fromAddrStr := range args[1:] {
				fromAddrsStrs[i], err = validateAddress(fromAddrStr, fmt.Sprintf("from_address %d", i+1))
				if err != nil {
					return err
				}
			}

			permanent, err := cmd.Flags().GetBool(FlagPermanent)
			if err != nil {
				return err
			}

			msg := quarantine.NewMsgAccept(toAddr, fromAddrsStrs, permanent)
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagPermanent, false, "Also set auto-accept for sends from any of the from_addresses to to_address")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// TxDeclineCmd returns the command for executing a Decline Tx.
func TxDeclineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decline <to_name_or_address> <from_address> [<from_address 2> ...]",
		Short: "Decline quarantined funds sent to <to_name_or_address> from <from_address>",
		Long: `Decline quarantined funds sent to <to_name_or_address> from <from_address>.
Note, the '--from' flag is ignored as it is implied from [to_name_or_address] (the signer of the message).`,
		Example: fmt.Sprintf(`
$ %[1]s decline %[2]s %[3]s
$ %[1]s decline personal %[3]s
$ %[1]s decline personal %[3]s %[4]s
`,
			exampleTxCmdBase, exampleAddr1, exampleAddr2, exampleAddr3),
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args[0]) == 0 {
				return fmt.Errorf("no to_name_or_address provided")
			}
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			toAddr := clientCtx.GetFromAddress()

			fromAddrsStrs := make([]string, len(args)-1)
			for i, fromAddrStr := range args[1:] {
				fromAddrsStrs[i], err = validateAddress(fromAddrStr, fmt.Sprintf("from_address %d", i+1))
				if err != nil {
					return err
				}
			}

			permanent, err := cmd.Flags().GetBool(FlagPermanent)
			if err != nil {
				return err
			}

			msg := quarantine.NewMsgDecline(toAddr, fromAddrsStrs, permanent)
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagPermanent, false, "Also set auto-decline for sends from any of the from_addresses to to_address")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// TxUpdateAutoResponsesCmd returns the command for executing an UpdateAutoResponses Tx.
func TxUpdateAutoResponsesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-auto-responses <to_name_or_address> <auto-response> <from_address> [<from_address 2> ...] [<auto-response 2> <from_address 3> [<from_address 4> ...] ...]",
		Aliases: []string{"auto-responses", "uar"},
		Short:   "Update auto-responses",
		Long: `Update auto-responses for transfers to <to_name_or_address> from one or more addresses.
Note, the '--from' flag is ignored as it is implied from [to_name_or_address] (the signer of the message).

The <to_name_or_address> is required.
At least one <auto-response> and <from_address> must be provided.

Valid <auto-response> values:
  "accept" or "a" - turn on auto-accept for the following <from_address>(es).
  "decline" or "d" - turn on auto-decline for the following <from_address>(es).
  "unspecified", "u", "off", or "o" - turn off auto-responses for the following <from_address>(es).

Each <auto-response> value can be repeated as an arg as many times as needed as long as each is followed by at least one <from_address>.
Each <from_address> will be assigned the nearest preceding <auto-response> value.
`,
		Example: fmt.Sprintf(`
$ %[1]s update-auto-responses %[2]s accept %[3]s
$ %[1]s update-auto-responses personal decline %[4]s unspecified %[5]s
$ %[1]s auto-responses personal accept %[3]s %[6]s off %[5]s
`,
			exampleTxCmdBase, exampleAddr1, exampleAddr2, exampleAddr("exampleAddr3"),
			exampleAddr("exampleAddr4"), exampleAddr("exampleAddr5")),
		Args: cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args[0]) == 0 {
				return fmt.Errorf("no to_name_or_address provided")
			}
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			toAddr := clientCtx.GetFromAddress()

			var updates []*quarantine.AutoResponseUpdate
			updates, err = ParseAutoResponseUpdatesFromArgs(args, 1)
			if err != nil {
				return err
			}

			msg := quarantine.NewMsgUpdateAutoResponses(toAddr, updates)
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
