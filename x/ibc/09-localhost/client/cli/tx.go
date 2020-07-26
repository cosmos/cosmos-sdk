package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/client/flags"
	"github.com/KiraCore/cosmos-sdk/client/tx"
	"github.com/KiraCore/cosmos-sdk/version"
	"github.com/KiraCore/cosmos-sdk/x/ibc/09-localhost/types"
	host "github.com/KiraCore/cosmos-sdk/x/ibc/24-host"
)

// NewCreateClientCmd defines the command to create a new IBC Loopback Client as defined
// in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
func NewCreateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "create new localhost client",
		Long:    "create new localhost (loopback) client",
		Example: fmt.Sprintf("%s tx %s %s create --from node0 --home ../node0/<app>cli --chain-id $CID", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateClient(clientCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
