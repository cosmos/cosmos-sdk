package cli

import (
	"github.com/spf13/cobra"

	"github.com/pointnetwork/cosmos-point-sdk/client"
	"github.com/pointnetwork/cosmos-point-sdk/client/flags"
	"github.com/pointnetwork/cosmos-point-sdk/client/tx"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	"github.com/pointnetwork/cosmos-point-sdk/x/slashing/types"
)

// NewTxCmd returns a root CLI command handler for all x/slashing transaction commands.
func NewTxCmd() *cobra.Command {
	slashingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Slashing transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	slashingTxCmd.AddCommand(NewUnjailTxCmd())
	return slashingTxCmd
}

// NewUnjailTxCmd returns a CLI command handler for creating a MsgUnjail transaction.
func NewUnjailTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unjail",
		Args:  cobra.NoArgs,
		Short: "unjail validator previously jailed for downtime",
		Long: `unjail a jailed validator:

$ <appd> tx slashing unjail --from mykey
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			valAddr := clientCtx.GetFromAddress()

			msg := types.NewMsgUnjail(sdk.ValAddress(valAddr))

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
