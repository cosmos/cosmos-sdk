package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// IBC transfer flags
var (
	FlagSource = "source"
)

// GetTxCmd returns the transaction commands for IBC fungible token transfer
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "transfer",
		Short: "IBC fungible token transfer transaction subcommands",
	}
	txCmd.AddCommand(
		GetTransferTxCmd(cdc),
		GetMsgRecvPacketCmd(cdc),
	)

	return txCmd
}

// GetTransferTxCmd returns the command to create a NewMsgTransfer transaction
func GetTransferTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer [src-port] [src-channel] [receiver] [amount]",
		Short: "Transfer fungible token through IBC",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx := context.NewCLIContext().WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			sender := ctx.GetFromAddress()
			srcPort := args[0]
			srcChannel := args[1]
			receiver, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			// parse coin trying to be sent
			coins, err := sdk.ParseCoins(args[3])
			if err != nil {
				return err
			}

			source := viper.GetBool(FlagSource)

			msg := types.NewMsgTransfer(srcPort, srcChannel, coins, sender, receiver, source)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(ctx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().Bool(FlagSource, false, "Pass flag for sending token from the source chain")
	return cmd
}

// GetMsgRecvPacketCmd returns the command to create a MsgRecvTransferPacket transaction
func GetMsgRecvPacketCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recv-packet [/path/to/packet-data.json] [/path/to/proof.json] [height]",
		Short: "Creates and sends a SendPacket message",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			var packet channel.Packet
			if err := cdc.UnmarshalJSON([]byte(args[0]), &packet); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...\n")
				contents, err := ioutil.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("error opening packet file: %v", err)
				}

				if err := cdc.UnmarshalJSON(contents, packet); err != nil {
					return fmt.Errorf("error unmarshalling packet file: %v", err)
				}
			}

			var proof commitment.Proof
			if err := cdc.UnmarshalJSON([]byte(args[1]), &proof); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...\n")
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return fmt.Errorf("error opening proofs file: %v", err)
				}
				if err := cdc.UnmarshalJSON(contents, &proof); err != nil {
					return fmt.Errorf("error unmarshalling proofs file: %v", err)
				}
			}

			height, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("error height: %v", err)
			}

			msg := types.NewMsgRecvPacket(packet, []commitment.Proof{proof}, height, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd = client.PostCommands(cmd)[0]
	return cmd
}
