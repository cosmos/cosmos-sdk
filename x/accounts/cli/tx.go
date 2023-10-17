package cli

import (
	v1 "cosmossdk.io/x/accounts/v1"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

func GeTxCmds() []*cobra.Command {
	return []*cobra.Command{GetTxInitCmd()}
}

func GetTxInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [account-type] json-message",
		Short: "Initialize a new account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := clientCtx.GetFromAddress()

			// we need to convert the message from json to a protobuf message

			msg := v1.MsgInit{
				Sender:      sender.String(),
				AccountType: args[0],
				Message:     nil,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
