package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

var (
	// DefaultAuthorityAddr is the default authority to provide in the sanction module's governance proposal messages.
	// It should match the value provided to the sanction keeper constructor.
	// It is defined as a sdk.AccAddress to be independent of global bech32 HRP definition.
	DefaultAuthorityAddr = authtypes.NewModuleAddress(govtypes.ModuleName)

	// exampleTxCmdBase is the base command that gets a user to one of the tx commands in here.
	exampleTxCmdBase = fmt.Sprintf("%s tx %s", version.AppName, sanction.ModuleName)
	// exampleTxAddr1 is a constant address for use in example strings.
	exampleTxAddr1 = sdk.AccAddress("exampleTxAddr1______")
	// exampleTxAddr2 is a constant address for use in example strings.
	exampleTxAddr2 = sdk.AccAddress("exampleTxAddr2______")
)

// TxCmd returns the command with sub-commands for specific sanction module Tx interaction.
func TxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        sanction.ModuleName,
		Short:                      "Sanction transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		TxSanctionCmd(),
		TxUnsanctionCmd(),
		TxUpdateParamsCmd(),
	)

	return txCmd
}

// TxSanctionCmd returns the command for submitting a MsgSanction governance proposal tx.
func TxSanctionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sanction <address 1> [<address 2> ...]",
		Short: "Submit a governance proposal to sanction one or more addresses",
		Long: `Submit a governance proposal to sanction one or more addresses.
At least one address is required; any number of addresses can be provided.
Each address should be a valid bech32 encoded string.`,
		Example: fmt.Sprintf(`
$ %[1]s sanction %[2]s
$ %[1]s sanction %[3]s %[2]s
`,
			exampleTxCmdBase, exampleTxAddr1, exampleTxAddr2),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()

			msgSanction := &sanction.MsgSanction{
				Addresses: args,
				Authority: getAuthority(flagSet),
			}
			if err = msgSanction.ValidateBasic(); err != nil {
				return err
			}

			return govcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msgSanction)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)
	addAuthorityFlagToCmd(cmd)

	return cmd
}

// TxUnsanctionCmd returns the command for submitting a MsgUnsanction governance proposal tx.
func TxUnsanctionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsanction <address 1> [<address 2> ...]",
		Short: "Submit a governance proposal to unsanction one or more addresses",
		Long: `Submit a governance proposal to unsanction one or more addresses.
At least one address is required; any number of addresses can be provided.
Each address should be a valid bech32 encoded string.`,
		Example: fmt.Sprintf(`
$ %[1]s unsanction %[3]s
$ %[1]s unsanction %[2]s %[3]s
`,
			exampleTxCmdBase, exampleTxAddr1, exampleTxAddr2),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()

			msgUnsanction := &sanction.MsgUnsanction{
				Addresses: args,
				Authority: getAuthority(flagSet),
			}
			if err = msgUnsanction.ValidateBasic(); err != nil {
				return err
			}

			return govcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msgUnsanction)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)
	addAuthorityFlagToCmd(cmd)

	return cmd
}

// TxUpdateParamsCmd returns the command for submitting a MsgUpdateParams governance proposal tx.
func TxUpdateParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params <immediate_sanction_min_deposit> <immediate_unsanction_min_deposit>",
		Short: "Submit a governance proposal to update the sanction module's params",
		Long: `Submit a governance proposal to update the sanction module's params.
Both <immediate_sanction_min_deposit> and <immediate_unsanction_min_deposit> are required.
They must be coins or empty strings.`,
		Example: fmt.Sprintf(`
$ %[1]s update-params 100%[2]s 150%[2]s
$ %[1]s update-params '' 50%[2]s
$ %[1]s update-params 75%[2]s ''
$ %[1]s update-params '' ''
`,
			exampleTxCmdBase, sdk.DefaultBondDenom),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()

			msgUpdateParams := &sanction.MsgUpdateParams{
				Params:    &sanction.Params{},
				Authority: getAuthority(flagSet),
			}

			if len(args[0]) > 0 {
				msgUpdateParams.Params.ImmediateSanctionMinDeposit, err = sdk.ParseCoinsNormalized(args[0])
				if err != nil {
					return fmt.Errorf("invalid immediate_sanction_min_deposit string %q: %w", args[0], err)
				}
			}

			if len(args[1]) > 0 {
				msgUpdateParams.Params.ImmediateUnsanctionMinDeposit, err = sdk.ParseCoinsNormalized(args[1])
				if err != nil {
					return fmt.Errorf("invalid immediate_unsanction_min_deposit string %q: %w", args[1], err)
				}
			}

			if err = msgUpdateParams.ValidateBasic(); err != nil {
				return err
			}

			return govcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msgUpdateParams)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)
	addAuthorityFlagToCmd(cmd)

	return cmd
}

// addAuthorityFlagToCmd adds the authority flag to a command.
func addAuthorityFlagToCmd(cmd *cobra.Command) {
	// Note: Not setting a default here because the HRP might not yet be set correctly.
	cmd.Flags().String(flags.FlagAuthority, "", "The authority to use. If not provided, a default is used")
}

// getAuthority gets the authority string from the flagSet or returns the default.
func getAuthority(flagSet *pflag.FlagSet) string {
	// Ignoring the error here since we really don't care,
	// and it's easier if this just returns a string.
	authority, _ := flagSet.GetString(flags.FlagAuthority)
	if len(authority) > 0 {
		return authority
	}
	return DefaultAuthorityAddr.String()
}
