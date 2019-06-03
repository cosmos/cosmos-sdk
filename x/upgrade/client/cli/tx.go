package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"strings"
	time2 "time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
)

const (
	TimeFormat = "2006-01-02T15:04:05Z"
)

// GetCmdSubmitProposal implements a command handler for submitting a software upgrade proposal transaction.
func GetCmdSubmitProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "software-upgrade --upgrade-name [name] (--upgrade-height [height] | --upgrade-time [time]) (--upgrade-info [info])",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a software upgrade proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a software upgrade along with an initial deposit.
`,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			from := cliCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString("deposit")
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoins(depositStr)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString("title")
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString("upgrade-name")
			if err != nil {
				return err
			}
			if len(name) == 0 {
				name = title
			}

			height, err := cmd.Flags().GetInt64("upgrade-height")
			if err != nil {
				return err
			}

			timeStr, err := cmd.Flags().GetString("upgrade-time")
			if err != nil {
				return err
			}

			if height != 0 {
				if len(timeStr) != 0 {
					return fmt.Errorf("only one of --upgrade-time or --upgrade-height should be specified")
				}
			}

			var time time2.Time
			if len(timeStr) != 0 {
				time, err = time2.Parse(TimeFormat, timeStr)
				if err != nil {
					return err
				}
			}

			info, err := cmd.Flags().GetString("upgrade-info")
			if err != nil {
				return err
			}

			content := upgrade.NewSoftwareUpgradeProposal(title, description,
				upgrade.Plan{Name: name, Time: time, Height: height, Info: info})

			msg := gov.NewMsgSubmitProposal(content, deposit, from)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().String("upgrade-name", "", "The name of the upgrade (if not specified title will be used)")
	cmd.Flags().Int64("upgrade-height", 0, "The height at which the upgrade must happen (not to be used together with --upgrade-time)")
	cmd.Flags().String("upgrade-time", "", fmt.Sprintf("The time at which the upgrade must happen (ex. %s) (not to be used together with --upgrade-height)", TimeFormat))
	cmd.Flags().String("upgrade-info", "", "Optional info for the planned upgrade such as commit hash, etc.")

	return cmd
}
