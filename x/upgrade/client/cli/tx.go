package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	// TimeFormat specifies ISO UTC format for submitting the time for a new upgrade proposal
	TimeFormat = "2006-01-02T15:04:05Z"

	FlagUpgradeHeight = "upgrade-height"
	FlagUpgradeTime   = "upgrade-time"
	FlagUpgradeInfo   = "upgrade-info"
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
		Use:   "software-upgrade [name] (--upgrade-height [height] | --upgrade-time [time]) (--upgrade-info [info]) [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a software upgrade proposal",
		Long: "Submit a software upgrade along with an initial deposit.\n" +
			"Please specify a unique name and height OR time for the upgrade to take effect.\n" +
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

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().Int64(FlagUpgradeHeight, 0, "The height at which the upgrade must happen (not to be used together with --upgrade-time)")
	cmd.Flags().String(FlagUpgradeTime, "", fmt.Sprintf("The time at which the upgrade must happen (ex. %s) (not to be used together with --upgrade-height)", TimeFormat))
	cmd.Flags().String(FlagUpgradeInfo, "", "Optional info for the planned upgrade such as commit hash, etc.")

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

			if err = msg.ValidateBasic(); err != nil {
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

	timeStr, err := cmd.Flags().GetString(FlagUpgradeTime)
	if err != nil {
		return nil, err
	}

	if height != 0 && len(timeStr) != 0 {
		return nil, fmt.Errorf("only one of --upgrade-time or --upgrade-height should be specified")
	}

	var upgradeTime time.Time
	if len(timeStr) != 0 {
		upgradeTime, err = time.Parse(TimeFormat, timeStr)
		if err != nil {
			return nil, err
		}
	}

	info, err := cmd.Flags().GetString(FlagUpgradeInfo)
	if err != nil {
		return nil, err
	}

	plan := types.Plan{Name: name, Time: upgradeTime, Height: height, Info: info}
	content := types.NewSoftwareUpgradeProposal(title, description, plan)
	return content, nil
}
