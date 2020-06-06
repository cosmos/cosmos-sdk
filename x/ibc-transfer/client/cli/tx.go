package cli

import (
	"bufio"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
)

const (
	flagTimeoutHeight    = "timeout-height"
	flagTimeoutTimestamp = "timeout-timestamp"
)

// GetTransferTxCmd returns the command to create a NewMsgTransfer transaction
func GetTransferTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transfer [src-port] [src-channel] [receiver] [amount]",
		Short:   "Transfer a fungible token through IBC",
		Example: fmt.Sprintf("%s tx ibc-transfer transfer [src-port] [src-channel] [receiver] [amount]", version.ClientName),
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			clientCtx := client.NewContextWithInput(inBuf).WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			sender := clientCtx.GetFromAddress()
			srcPort := args[0]
			srcChannel := args[1]
			receiver := args[2]

			coins, err := sdk.ParseCoins(args[3])
			if err != nil {
				return err
			}

			timeoutHeight := viper.GetUint64(flagTimeoutHeight)
			timeoutTimestamp := viper.GetUint64(flagTimeoutHeight)

			msg := types.NewMsgTransfer(
				srcPort, srcChannel, coins, sender, receiver, timeoutHeight, timeoutTimestamp,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(clientCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().Uint64(flagTimeoutHeight, types.DefaultAbsolutePacketTimeoutHeight, "Absolute timeout block height. The timeout is disabled when set to 0.")
	cmd.Flags().Uint64(flagTimeoutTimestamp, types.DefaultAbsolutePacketTimeoutTimestamp, "Absolute timeout timestamp in nanoseconds. The timeout is disabled when set to 0.")
	return cmd
}
