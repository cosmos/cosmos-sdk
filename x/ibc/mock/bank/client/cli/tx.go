package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/mock/bank/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "ibcmockbank",
		Short: "IBC mockbank module transaction subcommands",
		// RunE:  client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetTransferTxCmd(cdc),
		GetMsgRecvPacketCmd(cdc),
	)

	return txCmd
}

func GetTransferTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer --src-port <src port> --src-channel <src channel> --denom <denomination> --amount <amount> --receiver <receiver> --source <source>",
		Short: "Transfer tokens across chains through IBC",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx := context.NewCLIContext().WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			sender := ctx.GetFromAddress()
			receiver := viper.GetString(FlagReceiver)
			denom := viper.GetString(FlagDenom)
			srcPort := viper.GetString(FlagSrcPort)
			srcChan := viper.GetString(FlagSrcChannel)
			source := viper.GetBool(FlagSource)

			amount, ok := sdk.NewIntFromString(viper.GetString(FlagAmount))
			if !ok {
				return fmt.Errorf("invalid amount")
			}

			msg := types.NewMsgTransfer(srcPort, srcChan, denom, amount, sender, receiver, source)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(ctx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().AddFlagSet(FsTransfer)

	cmd.MarkFlagRequired(FlagSrcPort)
	cmd.MarkFlagRequired(FlagSrcChannel)
	cmd.MarkFlagRequired(FlagDenom)
	cmd.MarkFlagRequired(FlagAmount)
	cmd.MarkFlagRequired(FlagReceiver)

	cmd = client.PostCommands(cmd)[0]

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

			var packet types.Packet
			if err := cdc.UnmarshalJSON([]byte(args[0]), &packet); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...\n")
				contents, err := ioutil.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("error opening packet file: %v", err)
				}
				if err := packet.UnmarshalJSON(contents); err != nil {
					return fmt.Errorf("error unmarshalling packet file: %v", err)
				}
			}

			var proof ics23.Proof
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

			msg := types.NewMsgRecvTransferPacket(packet, []ics23.Proof{proof}, height, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd = client.PostCommands(cmd)[0]
	return cmd
}
