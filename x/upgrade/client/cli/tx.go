package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	FlagUpgradeHeight = "upgrade-height"
	FlagUpgradeInfo   = "upgrade-info"
	FlagNoValidate    = "no-validate"
	FlagDaemonName    = "daemon-name"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Upgrade transaction subcommands",
	}

	return cmd
}

// NewCmdSubmitUpgradeProposal implements a command handler for submitting a software upgrade proposal transaction.
func NewCmdSubmitUpgradeProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "software-upgrade [name] (--upgrade-height [height]) (--upgrade-info [info]) [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a software upgrade proposal",
		Long: "Submit a software upgrade along with an initial deposit.\n" +
			"Please specify a unique name and height for the upgrade to take effect.\n" +
			"You may include info to reference a binary download link, in a format compatible with: https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			name := args[0]
			content, err := parseArgsToContent(cmd, name)
			if err != nil {
				return err
			}
			noValidate, err := cmd.Flags().GetBool(FlagNoValidate)
			if err != nil {
				return err
			}
			if !noValidate {
				prop := content.(*types.SoftwareUpgradeProposal)
				if len(prop.Plan.Info) > 0 {
					var daemonName string
					if daemonName, err = getDaemonName(cmd); err != nil {
						return err
					}
					var planInfo *types.PlanInfo
					if planInfo, err = types.ParsePlanInfo(prop.Plan.Info); err != nil {
						return err
					}
					if err = planInfo.ValidateBasic(); err != nil {
						return err
					}
					if err = planInfo.Binaries.CheckURLs(daemonName); err != nil {
						return err
					}
				}
			}

			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			msg, err := gov.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().Int64(FlagUpgradeHeight, 0, "The height at which the upgrade must happen")
	cmd.Flags().String(FlagUpgradeInfo, "", "Optional info for the planned upgrade such as commit hash, etc.")
	cmd.Flags().Bool(FlagNoValidate, false, "Do not validate the info field (default behavior is to validate it).")
	_, defaultDaemonName := filepath.Split(os.Args[0])
	cmd.Flags().String(FlagDaemonName, defaultDaemonName, "The name of the executable binary that is being upgraded.")

	return cmd
}

// NewCmdSubmitCancelUpgradeProposal implements a command handler for submitting a software upgrade cancel proposal transaction.
func NewCmdSubmitCancelUpgradeProposal() *cobra.Command {
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
			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			content := types.NewCancelSoftwareUpgradeProposal(title, description)

			msg, err := gov.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.MarkFlagRequired(cli.FlagTitle)
	cmd.MarkFlagRequired(cli.FlagDescription)

	return cmd
}

func parseArgsToContent(cmd *cobra.Command, name string) (gov.Content, error) {
	title, err := cmd.Flags().GetString(cli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(cli.FlagDescription)
	if err != nil {
		return nil, err
	}

	height, err := cmd.Flags().GetInt64(FlagUpgradeHeight)
	if err != nil {
		return nil, err
	}

	info, err := cmd.Flags().GetString(FlagUpgradeInfo)
	if err != nil {
		return nil, err
	}

	plan := types.Plan{Name: name, Height: height, Info: info}
	content := types.NewSoftwareUpgradeProposal(title, description, plan)
	return content, nil
}

// getDaemonName gets the desired daemon name from the flag, env var, or default.
func getDaemonName(cmd *cobra.Command) (string, error) {
	// If something was provided for it, return that.
	if cmd.Flags().Changed(FlagDaemonName) {
		return cmd.Flags().GetString(FlagDaemonName)
	}
	// Check for a DAEMON_NAME env var and return that if found.
	// This is the same env var used by Cosmovisor.
	name := os.Getenv("DAEMON_NAME")
	if len(name) > 0 {
		return name, nil
	}
	// Return what the flag is (which should be the default).
	return cmd.Flags().GetString(FlagDaemonName)
}
