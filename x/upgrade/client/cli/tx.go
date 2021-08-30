package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	FlagUpgradeHeight       = "upgrade-height"
	FlagUpgradeInfo         = "upgrade-info"
	FlagUpgradeInstructions = "upgrade-instructions"
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
		Use:   fmt.Sprintf("software-upgrade <name> --%s <height> [--%s <info>] [--%s <instructions>] [flags]", FlagUpgradeHeight, FlagUpgradeInfo, FlagUpgradeInstructions),
		Args:  cobra.ExactArgs(1),
		Short: "Submit a software upgrade proposal",
		Long: "Submit a software upgrade along with an initial deposit.\n" +
			"You must use a unique name and height for the upgrade to take effect.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			name := args[0]
			content, err := parseArgsToContent(clientCtx.Codec, cmd, name)
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

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().Int64(FlagUpgradeHeight, 0, "The height at which the upgrade must happen")
	cmd.Flags().String(FlagUpgradeInfo, "", "Optional info for the planned upgrade such as commit hash or a binary download link, in a format compatible with: https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor")
	cmd.Flags().String(FlagUpgradeInstructions, "", `Optional, not app specific download instructions. It set, it must have the UpgradeInstructions JSON format.   Example 1: '{"pre_run": "./upgrade-v1", "assets": [{"platform": "linux/amd64", "url": "https://ipfs.io/ipfs/Qme7ss...", checksum: "0cdbd28e71a2e37830dabee99adffb68a568488f6fcfcf051217984151b769ee"}]}'   Example 2: '{"pre_run": "simd pre-upgrade"}'`)

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

func parseArgsToContent(cdc codec.Codec, cmd *cobra.Command, name string) (gov.Content, error) {
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
	upgradeInstructions, err := cmd.Flags().GetString(FlagUpgradeInstructions)
	if err != nil {
		return nil, err
	}
	var instructions types.UpgradeInstructions
	if upgradeInstructions != "" {
		if err = cdc.UnmarshalJSON([]byte(upgradeInstructions), &instructions); err != nil {
			return nil, errors.ErrJSONUnmarshal.Wrapf("Can't parse upgrade-instructions [%v]", err)
		}
	}

	plan := types.Plan{Name: name, Height: height, Info: info, Upgrade: instructions}
	content := types.NewSoftwareUpgradeProposal(title, description, plan)
	return content, nil
}
