package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	cutils "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// IBC transfer flags
var (
	FlagSource   = "source"
	FlagNode1    = "node1"
	FlagNode2    = "node2"
	FlagFrom1    = "from1"
	FlagFrom2    = "from2"
	FlagChainId2 = "chain-id2"
	FlagSequence = "packet-sequence"
	FlagTimeout  = "timeout"
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
	cmd.Flags().String(flags.FlagFrom, "", "key in local keystore to send from")
	return cmd
}

// GetMsgRecvPacketCmd returns the command to create a MsgRecvTransferPacket transaction
func GetMsgRecvPacketCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recv-packet [sending-port-id] [sending-channel-id]",
		Short: "Creates and sends a SendPacket message",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			node2 := viper.GetString(FlagNode2)
			cid2 := viper.GetString(FlagChainId2)
			cliCtx2 := context.NewCLIContextIBC(cliCtx.GetFromAddress().String(), cid2, node2).WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			viper.Set(flags.FlagChainID, cid2)
			sourcePort := args[0]
			sourceChannel := args[1]
			sequence := uint64(viper.GetInt(FlagSequence))
			packetRes, err := cutils.QueryPacket(cliCtx2, sourcePort, sourceChannel, sequence, "")
			if err != nil {
				return err
			}
			viper.Set(flags.FlagChainID, viper.GetString(flags.FlagChainID))

			msg := types.NewMsgRecvPacket(packetRes.Packet, []commitment.Proof{packetRes.Proof}, packetRes.ProofHeight, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			fmt.Printf("%+v\n", msg)

			fmt.Println(99999)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd = client.PostCommands(cmd)[0]
	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "RPC port for the second chain")
	cmd.Flags().String(FlagChainId2, "", "chain-id for the second chain")
	cmd.Flags().String(FlagSequence, "", "sequence for the packet")
	cmd.Flags().String(FlagTimeout, "", "timeout for the packet")
	return cmd
}
