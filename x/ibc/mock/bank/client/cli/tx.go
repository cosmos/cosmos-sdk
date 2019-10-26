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
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/mock/bank/internal/types"
	"github.com/spf13/cobra"
)

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "ibcmockbank",
		Short: "IBC mockbank module transaction subcommands",
		// RunE:  client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetMsgRecvPacketCmd(cdc),
	)

	return txCmd
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
