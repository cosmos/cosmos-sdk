package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
)

const (
	flagTimeoutHeight    = "timeout-height"
	flagTimeoutTimestamp = "timeout-timestamp"
)

// NewTransferTxCmd returns the command to create a NewMsgTransfer transaction
func NewTransferTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transfer [src-port] [src-channel] [receiver] [amount]",
		Short:   "Transfer a fungible token through IBC",
		Example: fmt.Sprintf("%s tx ibc-transfer transfer [src-port] [src-channel] [receiver] [amount]", version.AppName),
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			sender := clientCtx.GetFromAddress()
			srcPort := args[0]
			srcChannel := args[1]
			receiver := args[2]

			coins, err := sdk.ParseCoins(args[3])
			if err != nil {
				return err
			}

			timeoutHeight, _ := cmd.Flags().GetUint64(flagTimeoutHeight)
			timeoutTimestamp, _ := cmd.Flags().GetUint64(flagTimeoutHeight)

			msg := types.NewMsgTransfer(
				srcPort, srcChannel, coins, sender, receiver, timeoutHeight, timeoutTimestamp,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Uint64(flagTimeoutHeight, types.DefaultAbsolutePacketTimeoutHeight, "Absolute timeout block height. The timeout is disabled when set to 0.")
	cmd.Flags().Uint64(flagTimeoutTimestamp, types.DefaultAbsolutePacketTimeoutTimestamp, "Absolute timeout timestamp in nanoseconds. The timeout is disabled when set to 0.")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
