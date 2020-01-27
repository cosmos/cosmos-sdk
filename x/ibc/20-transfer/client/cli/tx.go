package cli

import (
	"bufio"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

// IBC transfer flags
var (
	FlagSource   = "source"
	FlagNode1    = "node1"
	FlagNode2    = "node2"
	FlagFrom1    = "from1"
	FlagFrom2    = "from2"
	FlagChainID2 = "chain-id2"
	FlagSequence = "packet-sequence"
	FlagTimeout  = "timeout"
)

// GetTransferTxCmd returns the command to create a NewMsgTransfer transaction
func GetTransferTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer [src-port] [src-channel] [receiver] [amount]",
		Short: "Transfer fungible token through IBC",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			ctx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

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

			return authclient.GenerateOrBroadcastMsgs(ctx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().Bool(FlagSource, false, "Pass flag for sending token from the source chain")
	return cmd
}

// // GetMsgRecvPacketCmd returns the command to create a MsgRecvTransferPacket transaction
// func GetMsgRecvPacketCmd(cdc *codec.Codec) *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "recv-packet [sending-port-id] [sending-channel-id] [client-id]",
// 		Short: "Creates and sends a SendPacket message",
// 		Args:  cobra.ExactArgs(3),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			inBuf := bufio.NewReader(cmd.InOrStdin())
// 			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
// 			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

// 			prove := viper.GetBool(flags.FlagProve)
// 			node2 := viper.GetString(FlagNode2)
// 			cid1 := viper.GetString(flags.FlagChainID)
// 			cid2 := viper.GetString(FlagChainID2)
// 			cliCtx2 := context.NewCLIContextIBC(cliCtx.GetFromAddress().String(), cid2, node2).
// 				WithCodec(cdc).
// 				WithBroadcastMode(flags.BroadcastBlock)

// 			header, _, err := clientutils.QueryTendermintHeader(cliCtx2)
// 			if err != nil {
// 				return err
// 			}

// 			sourcePort, sourceChannel, clientid := args[0], args[1], args[2]

// 			passphrase, err := keys.GetPassphrase(viper.GetString(flags.FlagFrom))
// 			if err != nil {
// 				return nil
// 			}

// 			viper.Set(flags.FlagChainID, cid1)
// 			msgUpdateClient := clienttypes.NewMsgUpdateClient(clientid, header, cliCtx.GetFromAddress())
// 			if err := msgUpdateClient.ValidateBasic(); err != nil {
// 				return err
// 			}

// 			res, err := utils.CompleteAndBroadcastTx(txBldr, cliCtx, []sdk.Msg{msgUpdateClient}, passphrase)
// 			if err != nil || !res.IsOK() {
// 				return err
// 			}

// 			viper.Set(flags.FlagChainID, cid2)
// 			sequence := uint64(viper.GetInt(FlagSequence))
// 			packetRes, err := channelutils.QueryPacket(cliCtx2.WithHeight(header.Height-1), sourcePort, sourceChannel, sequence, uint64(viper.GetInt(FlagTimeout)), prove)
// 			if err != nil {
// 				return err
// 			}

// 			viper.Set(flags.FlagChainID, cid1)

// 			msg := types.NewMsgRecvPacket(packetRes.Packet, []commitment.Proof{packetRes.Proof}, packetRes.ProofHeight, cliCtx.GetFromAddress())
// 			if err := msg.ValidateBasic(); err != nil {
// 				return err
// 			}

// 			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
// 		},
// 	}

// 	cmd = client.PostCommands(cmd)[0]
// 	cmd.Flags().Bool(FlagSource, false, "Pass flag for sending token from the source chain")
// 	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "RPC port for the second chain")
// 	cmd.Flags().String(FlagChainID2, "", "chain-id for the second chain")
// 	cmd.Flags().String(FlagSequence, "", "sequence for the packet")
// 	cmd.Flags().String(FlagTimeout, "", "timeout for the packet")
// 	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
// 	return cmd
// }
