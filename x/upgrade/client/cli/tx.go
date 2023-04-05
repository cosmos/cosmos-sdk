package cli

import (
	"fmt"
	"os"
	"path/filepath"

	addresscodec "cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/spf13/cobra"

	"cosmossdk.io/x/upgrade/plan"
	"cosmossdk.io/x/upgrade/types"
)

const (
	FlagUpgradeHeight = "upgrade-height"
	FlagUpgradeInfo   = "upgrade-info"
	FlagNoValidate    = "no-validate"
	FlagDaemonName    = "daemon-name"
	FlagAuthority     = "authority"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(ac addresscodec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Upgrade transaction subcommands",
	}

	cmd.AddCommand(
		NewCmdSubmitUpgradeProposal(ac),
		NewCmdSubmitCancelUpgradeProposal(ac),
	)

	return cmd
}

// NewCmdSubmitUpgradeProposal implements a command handler for submitting a software upgrade proposal transaction.
func NewCmdSubmitUpgradeProposal(ac addresscodec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "software-upgrade [name] (--upgrade-height [height]) (--upgrade-info [info]) [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a software upgrade proposal",
		Long: "Submit a software upgrade along with an initial deposit.\n" +
			"Please specify a unique name and height for the upgrade to take effect.\n" +
			"You may include info to reference a binary download link, in a format compatible with: https://docs.cosmos.network/main/tooling/cosmovisor",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := cli.ReadGovPropFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			name := args[0]
			p, err := parsePlan(cmd.Flags(), name)
			if err != nil {
				return err
			}

			noValidate, err := cmd.Flags().GetBool(FlagNoValidate)
			if err != nil {
				return err
			}

			if !noValidate {
				var daemonName string
				if daemonName, err = cmd.Flags().GetString(FlagDaemonName); err != nil {
					return err
				}

				var planInfo *plan.Info
				if planInfo, err = plan.ParseInfo(p.Info); err != nil {
					return err
				}
				if err = planInfo.ValidateFull(daemonName); err != nil {
					return err
				}
			}

			authority, _ := cmd.Flags().GetString(FlagAuthority)
			if authority != "" {
				if _, err = ac.StringToBytes(authority); err != nil {
					return fmt.Errorf("invalid authority address: %w", err)
				}
			} else {
				authority = sdk.AccAddress(address.Module("gov")).String()
			}

			if err := proposal.SetMsgs([]sdk.Msg{
				&types.MsgSoftwareUpgrade{
					Authority: authority,
					Plan:      p,
				},
			}); err != nil {
				return fmt.Errorf("failed to create cancel upgrade message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposal)
		},
	}

	cmd.Flags().Int64(FlagUpgradeHeight, 0, "The height at which the upgrade must happen")
	cmd.Flags().String(FlagUpgradeInfo, "", "Info for the upgrade plan such as new version download urls, etc.")
	cmd.Flags().Bool(FlagNoValidate, false, "Skip validation of the upgrade info (dangerous!)")
	cmd.Flags().String(FlagDaemonName, getDefaultDaemonName(), "The name of the executable being upgraded (for upgrade-info validation). Default is the DAEMON_NAME env var if set, or else this executable")
	cmd.Flags().String(FlagAuthority, "", "The address of the upgrade module authority (defaults to gov)")

	// add common proposal flags
	flags.AddTxFlagsToCmd(cmd)
	cli.AddGovPropFlagsToCmd(cmd)
	cmd.MarkFlagRequired(cli.FlagTitle)

	return cmd
}

// NewCmdSubmitCancelUpgradeProposal implements a command handler for submitting a software upgrade cancel proposal transaction.
func NewCmdSubmitCancelUpgradeProposal(ac addresscodec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-software-upgrade [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Cancel the current software upgrade proposal",
		Long:  "Cancel a software upgrade along with an initial deposit.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := cli.ReadGovPropFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			authority, _ := cmd.Flags().GetString(FlagAuthority)
			if authority != "" {
				if _, err = ac.StringToBytes(authority); err != nil {
					return fmt.Errorf("invalid authority address: %w", err)
				}
			} else {
				authority = sdk.AccAddress(address.Module("gov")).String()
			}

			if err := proposal.SetMsgs([]sdk.Msg{
				&types.MsgCancelUpgrade{
					Authority: authority,
				},
			}); err != nil {
				return fmt.Errorf("failed to create cancel upgrade message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposal)
		},
	}

	cmd.Flags().String(FlagAuthority, "", "The address of the upgrade module authority (defaults to gov)")

	// add common proposal flags
	flags.AddTxFlagsToCmd(cmd)
	cli.AddGovPropFlagsToCmd(cmd)
	cmd.MarkFlagRequired(cli.FlagTitle)

	return cmd
}

// getDefaultDaemonName gets the default name to use for the daemon.
// If a DAEMON_NAME env var is set, that is used.
// Otherwise, the last part of the currently running executable is used.
func getDefaultDaemonName() string {
	// DAEMON_NAME is specifically used here to correspond with the Cosmovisor setup env vars.
	name := os.Getenv("DAEMON_NAME")
	if len(name) == 0 {
		_, name = filepath.Split(os.Args[0])
	}
	return name
}
