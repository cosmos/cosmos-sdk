package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
)

const (
	flagTimeoutEpoch     = "timeout-epoch"
	flagTimeoutHeight    = "timeout-height"
	flagTimeoutTimestamp = "timeout-timestamp"
)

// NewTransferTxCmd returns the command to create a NewMsgTransfer transaction
func NewTransferTxCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transfer [src-port] [src-channel] [receiver] [amount]",
		Short:   "Transfer a fungible token through IBC",
		Example: fmt.Sprintf("%s tx ibc-transfer transfer [src-port] [src-channel] [receiver] [amount]", version.AppName),
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx = clientCtx.InitWithInput(cmd.InOrStdin())

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
			timeoutEpoch, _ := cmd.Flags().GetUint64(flagTimeoutEpoch)

			msg := types.NewMsgTransfer(
				srcPort, srcChannel, coins, sender, receiver, timeoutEpoch, timeoutHeight, timeoutTimestamp,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(clientCtx, msg)
		},
	}

	cmd.Flags().Uint64(flagTimeoutHeight, types.DefaultAbsolutePacketTimeoutHeight, "Absolute timeout block height. The timeout is disabled when set to 0.")
	cmd.Flags().Uint64(flagTimeoutTimestamp, types.DefaultAbsolutePacketTimeoutTimestamp, "Absolute timeout timestamp in nanoseconds. The timeout is disabled when set to 0.")
	cmd.Flags().Uint64(flagTimeoutEpoch, 0, "Absolute timeout block epoch.")

	return cmd
}
