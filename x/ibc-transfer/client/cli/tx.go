package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	channelutils "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/client/utils"
)

const (
<<<<<<< HEAD
	flagTimeoutEpoch     = "timeout-epoch"
	flagTimeoutHeight    = "timeout-height"
	flagTimeoutTimestamp = "timeout-timestamp"
	flagAbsoluteTimeouts = "absolute-timeouts"
=======
	flagPacketTimeoutHeight    = "packet-timeout-height"
	flagPacketTimeoutTimestamp = "packet-timeout-timestamp"
	flagAbsoluteTimeouts       = "absolute-timeouts"
>>>>>>> d9fd4d2ca9a3f70fbabcd3eb6a1427395fdedf74
)

// NewTransferTxCmd returns the command to create a NewMsgTransfer transaction
func NewTransferTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer [src-port] [src-channel] [receiver] [amount]",
		Short: "Transfer a fungible token through IBC",
		Long: strings.TrimSpace(`Transfer a fungible token through IBC. Timeouts can be specified
as absolute or relative using the "absolute-timeouts" flag. Relative timeouts are added to
the block height and block timestamp queried from the latest consensus state corresponding
to the counterparty channel. Any timeout set to 0 is disabled.`),
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

			coin, err := sdk.ParseCoin(args[3])
			if err != nil {
				return err
			}

			if !strings.HasPrefix(coin.Denom, "ibc/") {
				denomTrace := types.ParseDenomTrace(coin.Denom)
				coin.Denom = denomTrace.IBCDenom()
			}

			timeoutHeight, err := cmd.Flags().GetUint64(flagPacketTimeoutHeight)
			if err != nil {
				return err
			}

<<<<<<< HEAD
			timeoutEpoch, err := cmd.Flags().GetUint64(flagTimeoutEpoch)
			if err != nil {
				return err
			}

			timeoutTimestamp, err := cmd.Flags().GetUint64(flagTimeoutHeight)
=======
			timeoutTimestamp, err := cmd.Flags().GetUint64(flagPacketTimeoutTimestamp)
>>>>>>> d9fd4d2ca9a3f70fbabcd3eb6a1427395fdedf74
			if err != nil {
				return err
			}

			absoluteTimeouts, err := cmd.Flags().GetBool(flagAbsoluteTimeouts)
			if err != nil {
				return err
			}

			// if the timeouts are not absolute, retrieve latest block height and block timestamp
			// for the consensus state connected to the destination port/channel
			if !absoluteTimeouts {
				consensusState, _, err := channelutils.QueryCounterpartyConsensusState(clientCtx, srcPort, srcChannel, uint64(clientCtx.Height))
				if err != nil {
					return err
				}

				// If timeoutEpoch is not specified, use same epoch as latest consensus state
				if timeoutEpoch != 0 {
					timeoutEpoch = consensusState.GetHeight().EpochNumber
				}

				if timeoutHeight != 0 {
					timeoutHeight = consensusState.GetHeight().EpochHeight + timeoutHeight
				}

				if timeoutTimestamp != 0 {
					timeoutTimestamp = consensusState.GetTimestamp() + timeoutTimestamp
				}
			}

			msg := types.NewMsgTransfer(
<<<<<<< HEAD
				srcPort, srcChannel, coins, sender, receiver, timeoutEpoch, timeoutHeight, timeoutTimestamp,
=======
				srcPort, srcChannel, coin, sender, receiver, timeoutHeight, timeoutTimestamp,
>>>>>>> d9fd4d2ca9a3f70fbabcd3eb6a1427395fdedf74
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

<<<<<<< HEAD
	cmd.Flags().Uint64(flagTimeoutEpoch, 0, "Absolute timeout block epoch.")
	cmd.Flags().Uint64(flagTimeoutHeight, types.DefaultRelativePacketTimeoutHeight, "Timeout block height. The timeout is disabled when set to 0.")
	cmd.Flags().Uint64(flagTimeoutTimestamp, types.DefaultRelativePacketTimeoutTimestamp, "Timeout timestamp in nanoseconds. Default is 10 minutes. The timeout is disabled when set to 0.")
=======
	cmd.Flags().Uint64(flagPacketTimeoutHeight, types.DefaultRelativePacketTimeoutHeight, "Packet timeout block height. The timeout is disabled when set to 0.")
	cmd.Flags().Uint64(flagPacketTimeoutTimestamp, types.DefaultRelativePacketTimeoutTimestamp, "Packet timeout timestamp in nanoseconds. Default is 10 minutes. The timeout is disabled when set to 0.")
>>>>>>> d9fd4d2ca9a3f70fbabcd3eb6a1427395fdedf74
	cmd.Flags().Bool(flagAbsoluteTimeouts, false, "Timeout flags are used as absolute timeouts.")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
