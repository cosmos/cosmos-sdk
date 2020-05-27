package cli

import (
	"bufio"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	connectionutils "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// IBC Channel flags
const (
	FlagOrdered    = "ordered"
	FlagIBCVersion = "ibc-version"
)

// GetMsgChannelOpenInitCmd returns the command to create a MsgChannelOpenInit transaction
func GetMsgChannelOpenInitCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-init [port-id] [channel-id] [counterparty-port-id] [counterparty-channel-id] [connection-hops]",
		Short: "Creates and sends a ChannelOpenInit message",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			portID := args[0]
			channelID := args[1]
			counterpartyPortID := args[2]
			counterpartyChannelID := args[3]
			hops := strings.Split(args[4], "/")
			order := channelOrder()
			version := viper.GetString(FlagIBCVersion)

			msg := types.NewMsgChannelOpenInit(
				portID, channelID, version, order, hops,
				counterpartyPortID, counterpartyChannelID, cliCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Bool(FlagOrdered, true, "Pass flag for opening ordered channels")
	cmd.Flags().String(FlagIBCVersion, "1.0.0", "supported IBC version")

	return cmd
}

// GetMsgChannelOpenTryCmd returns the command to create a MsgChannelOpenTry transaction
func GetMsgChannelOpenTryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-try [port-id] [channel-id] [counterparty-port-id] [counterparty-channel-id] [connection-hops] [/path/to/proof_init.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenTry message",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			portID := args[0]
			channelID := args[1]
			counterpartyPortID := args[2]
			counterpartyChannelID := args[3]
			hops := strings.Split(args[4], "/")
			order := channelOrder()
			version := viper.GetString(FlagIBCVersion) // TODO: diferenciate between channel and counterparty versions

			proofInit, err := connectionutils.ParseProof(cliCtx.Codec, args[5])
			if err != nil {
				return err
			}

			proofHeight, err := strconv.ParseInt(args[6], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenTry(
				portID, channelID, version, order, hops,
				counterpartyPortID, counterpartyChannelID, version,
				proofInit, uint64(proofHeight), cliCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().Bool(FlagOrdered, true, "Pass flag for opening ordered channels")
	cmd.Flags().String(FlagIBCVersion, "1.0.0", "supported IBC version")

	return cmd
}

// GetMsgChannelOpenAckCmd returns the command to create a MsgChannelOpenAck transaction
func GetMsgChannelOpenAckCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-ack [port-id] [channel-id] [/path/to/proof_try.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenAck message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			portID := args[0]
			channelID := args[1]
			version := viper.GetString(FlagIBCVersion) // TODO: diferenciate between channel and counterparty versions

			proofTry, err := connectionutils.ParseProof(cliCtx.Codec, args[5])
			if err != nil {
				return err
			}

			proofHeight, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenAck(
				portID, channelID, version, proofTry, uint64(proofHeight), cliCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(FlagIBCVersion, "1.0.0", "supported IBC version")
	return cmd
}

// GetMsgChannelOpenConfirmCmd returns the command to create a MsgChannelOpenConfirm transaction
func GetMsgChannelOpenConfirmCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "open-confirm [port-id] [channel-id] [/path/to/proof_ack.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenConfirm message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			portID := args[0]
			channelID := args[1]

			proofAck, err := connectionutils.ParseProof(cliCtx.Codec, args[5])
			if err != nil {
				return err
			}

			proofHeight, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenConfirm(
				portID, channelID, proofAck, uint64(proofHeight), cliCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetMsgChannelCloseInitCmd returns the command to create a MsgChannelCloseInit transaction
func GetMsgChannelCloseInitCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "close-init [port-id] [channel-id]",
		Short: "Creates and sends a ChannelCloseInit message",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			portID := args[0]
			channelID := args[1]

			msg := types.NewMsgChannelCloseInit(portID, channelID, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetMsgChannelCloseConfirmCmd returns the command to create a MsgChannelCloseConfirm transaction
func GetMsgChannelCloseConfirmCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "close-confirm [port-id] [channel-id] [/path/to/proof_init.json] [proof-height]",
		Short: "Creates and sends a ChannelCloseConfirm message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			portID := args[0]
			channelID := args[1]

			proofInit, err := connectionutils.ParseProof(cliCtx.Codec, args[5])
			if err != nil {
				return err
			}

			proofHeight, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelCloseConfirm(
				portID, channelID, proofInit, uint64(proofHeight), cliCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func channelOrder() types.Order {
	if viper.GetBool(FlagOrdered) {
		return types.ORDERED
	}
	return types.UNORDERED
}
