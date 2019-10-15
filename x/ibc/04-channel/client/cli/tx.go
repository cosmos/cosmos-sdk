package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	FlagUnordered = "unordered"
	IBCVersion    = "version"
)

// GetTxCmd returns the transaction commands for IBC Connections
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ics04ChannelTxCmd := &cobra.Command{
		Use:   "connection",
		Short: "IBC connection transaction subcommands",
	}

	ics04ChannelTxCmd.AddCommand(client.PostCommands(
		GetMsgChannelOpenInitCmd(storeKey, cdc),
		GetMsgChannelOpenTryCmd(storeKey, cdc),
		GetMsgChannelOpenAckCmd(storeKey, cdc),
		GetMsgChannelOpenConfirmCmd(storeKey, cdc),
		GetMsgChannelCloseInitCmd(storeKey, cdc),
		GetMsgChannelCloseConfirmCmd(storeKey, cdc),
		GetMsgSendPacketCmd(storeKey, cdc),
	)...)

	return ics04ChannelTxCmd
}

// GetMsgChannelOpenInitCmd returns the command to create a MsgChannelOpenInit transaction
func GetMsgChannelOpenInitCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-init [port-id] [channel-id] [cp-port-id] [cp-channel-id] [connection-hops]",
		Short: "Creates and sends a ChannelOpenInit message",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID, err := validatePortID(args[0])
			if err != nil {
				return err
			}

			channelID, err := validateChannelID(args[1])
			if err != nil {
				return err
			}

			channel, err := createChannelFromArgs(args[2], args[3], args[4])
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenInit(portID, channelID, channel, cliCtx.GetFromAddress())

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Bool(FlagUnordered, false, "Pass flag for opening unordered channels")

	return cmd
}

// GetMsgChannelOpenTryCmd returns the command to create a MsgChannelOpenTry transaction
func GetMsgChannelOpenTryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-try [port-id] [channel-id] [cp-port-id] [cp-channel-id] [connection-hops] [/path/to/proof-init.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenTry message",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID, err := validatePortID(args[0])
			if err != nil {
				return err
			}

			channelID, err := validateChannelID(args[1])
			if err != nil {
				return err
			}

			channel, err := createChannelFromArgs(args[2], args[3], args[4])
			if err != nil {
				return err
			}

			var proof ics23.Proof
			if err := cdc.UnmarshalJSON([]byte(args[5]), &proof); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[5])
				if err != nil {
					return fmt.Errorf("error opening proof file: %v\n", err)
				}
				if err := cdc.UnmarshalJSON(contents, &proof); err != nil {
					return fmt.Errorf("error unmarshalling proof file: %v\n", err)
				}
			}

			proofHeight, err := validateProofHeight(args[6])
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenTry(portID, channelID, channel, IBCVersion, proof, proofHeight, cliCtx.GetFromAddress())

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Bool(FlagUnordered, false, "Pass flag for opening unordered channels")

	return cmd
}

// GetMsgChannelOpenAckCmd returns the command to create a MsgChannelOpenAck transaction
func GetMsgChannelOpenAckCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-ack [port-id] [channel-id] [/path/to/proof-try.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenAck message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID, err := validatePortID(args[0])
			if err != nil {
				return err
			}

			channelID, err := validateChannelID(args[1])
			if err != nil {
				return err
			}

			var proof ics23.Proof
			if err := cdc.UnmarshalJSON([]byte(args[2]), &proof); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[2])
				if err != nil {
					return fmt.Errorf("error opening proof file: %v\n", err)
				}
				if err := cdc.UnmarshalJSON(contents, &proof); err != nil {
					return fmt.Errorf("error unmarshalling proof file: %v\n", err)
				}
			}

			proofHeight, err := validateProofHeight(args[3])
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenAck(portID, channelID, IBCVersion, proof, proofHeight, cliCtx.GetFromAddress())

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// GetMsgChannelOpenConfirmCmd returns the command to create a MsgChannelOpenConfirm transaction
func GetMsgChannelOpenConfirmCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-confirm [port-id] [channel-id] [/path/to/proof-ack.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenConfirm message",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID, err := validatePortID(args[0])
			if err != nil {
				return err
			}

			channelID, err := validateChannelID(args[1])
			if err != nil {
				return err
			}

			var proof ics23.Proof
			if err := cdc.UnmarshalJSON([]byte(args[2]), &proof); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[2])
				if err != nil {
					return fmt.Errorf("error opening proof file: %v\n", err)
				}
				if err := cdc.UnmarshalJSON(contents, &proof); err != nil {
					return fmt.Errorf("error unmarshalling proof file: %v\n", err)
				}
			}

			proofHeight, err := validateProofHeight(args[3])
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenConfirm(portID, channelID, proof, proofHeight, cliCtx.GetFromAddress())

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// GetMsgChannelCloseInitCmd returns the command to create a MsgChannelCloseInit transaction
func GetMsgChannelCloseInitCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close-init [port-id] [channel-id]",
		Short: "Creates and sends a ChannelCloseInit message",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID, err := validatePortID(args[0])
			if err != nil {
				return err
			}

			channelID, err := validateChannelID(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelCloseInit(portID, channelID, cliCtx.GetFromAddress())

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// GetMsgChannelCloseConfirmCmd returns the command to create a MsgChannelCloseConfirm transaction
func GetMsgChannelCloseConfirmCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close-confirm [port-id] [channel-id] [/path/to/proof-init.json] [proof-height]",
		Short: "Creates and sends a ChannelCloseConfirm message",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID, err := validatePortID(args[0])
			if err != nil {
				return err
			}

			channelID, err := validateChannelID(args[1])
			if err != nil {
				return err
			}

			var proof ics23.Proof
			if err := cdc.UnmarshalJSON([]byte(args[2]), &proof); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[2])
				if err != nil {
					return fmt.Errorf("error opening proof file: %v\n", err)
				}
				if err := cdc.UnmarshalJSON(contents, &proof); err != nil {
					return fmt.Errorf("error unmarshalling proof file: %v\n", err)
				}
			}

			proofHeight, err := validateProofHeight(args[3])
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelCloseConfirm(portID, channelID, proof, proofHeight, cliCtx.GetFromAddress())

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// GetMsgSendPacketCmd returns the command to create a MsgChannelCloseConfirm transaction
func GetMsgSendPacketCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-packet [channel-id] [/path/to/packet-proof.json] [proof-height] [/path/to/packet-data.json]",
		Short: "Creates and sends a SendPacket message",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			channelID, err := validateChannelID(args[0])
			if err != nil {
				return err
			}

			var proofs []ics23.Proof
			if err := cdc.UnmarshalJSON([]byte(args[1]), &proofs); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return fmt.Errorf("error opening proofs file: %v\n", err)
				}
				if err := cdc.UnmarshalJSON(contents, &proofs); err != nil {
					return fmt.Errorf("error unmarshalling proofs file: %v\n", err)
				}
			}

			proofHeight, err := validateProofHeight(args[2])
			if err != nil {
				return err
			}

			var packet exported.PacketI
			if err := cdc.UnmarshalJSON([]byte(args[3]), &packet); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[3])
				if err != nil {
					return fmt.Errorf("error opening packet file: %v\n", err)
				}
				if err := cdc.UnmarshalJSON(contents, &packet); err != nil {
					return fmt.Errorf("error unmarshalling packet file: %v\n", err)
				}
			}

			msg := types.NewMsgSendPacket(packet, channelID, proofs, proofHeight, cliCtx.GetFromAddress())

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

func channelOrder() types.ChannelOrder {
	if viper.GetBool(FlagUnordered) {
		return types.UNORDERED
	}
	return types.ORDERED
}

func validatePortID(pid string) (string, error) {
	// TODO: Add validation here
	return pid, nil
}

func validateChannelID(cid string) (string, error) {
	// TODO: Add validation here
	return cid, nil
}

func validateChannelHops(hops string) ([]string, error) {
	// TODO: Add validation here
	return strings.Split(hops, ","), nil
}

func validateProofHeight(height string) (uint64, error) {
	// TODO: More validation?
	i, err := strconv.ParseInt(height, 10, 64)
	return uint64(i), err
}

func createChannelFromArgs(pid string, cid string, hops string) (types.Channel, error) {
	var channel types.Channel
	portID, err := validatePortID(pid)
	if err != nil {
		return channel, err
	}

	channelID, err := validateChannelID(cid)
	if err != nil {
		return channel, err
	}

	channelHops, err := validateChannelHops(hops)
	if err != nil {
		return channel, err
	}

	channel = types.Channel{
		State:          types.INIT,
		Ordering:       channelOrder(),
		Counterparty:   types.Counterparty{portID, channelID},
		ConnectionHops: channelHops,
		Version:        IBCVersion,
	}

	return channel, nil
}
