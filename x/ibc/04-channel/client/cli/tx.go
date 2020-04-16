package cli

import (
	"bufio"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	connectionutils "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC Channel flags
const (
	FlagOrdered    = "ordered"
	FlagIBCVersion = "ibc-version"
)

// NewChannelOpenInitTxCmd returns the command to create a MsgChannelOpenInit transaction
func NewChannelOpenInitTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-init [port-id] [channel-id] [counterparty-port-id] [counterparty-channel-id] [connection-hops]",
		Short: "Creates and sends a ChannelOpenInit message",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txf := tx.NewFactoryFromCLI(inBuf).
				WithTxGenerator(txg).
				WithAccountRetriever(ar)
			cliCtx := context.NewCLIContextWithInput(inBuf).WithMarshaler(m)

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

			return tx.GenerateOrBroadcastTx(cliCtx, txf, msg)
		},
	}

	cmd.Flags().Bool(FlagOrdered, true, "Pass flag for opening ordered channels")
	cmd.Flags().String(FlagIBCVersion, "1.0.0", "supported IBC version")

	return cmd
}

// NewChannelOpenTryTxCmd returns the command to create a MsgChannelOpenTry transaction
func NewChannelOpenTryTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-try [port-id] [channel-id] [counterparty-port-id] [counterparty-channel-id] [connection-hops] [/path/to/proof_init.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenTry message",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txf := tx.NewFactoryFromCLI(inBuf).
				WithTxGenerator(txg).
				WithAccountRetriever(ar)
			cliCtx := context.NewCLIContextWithInput(inBuf).WithMarshaler(m)

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

			return tx.GenerateOrBroadcastTx(cliCtx, txf, msg)
		},
	}
	cmd.Flags().Bool(FlagOrdered, true, "Pass flag for opening ordered channels")
	cmd.Flags().String(FlagIBCVersion, "1.0.0", "supported IBC version")

	return cmd
}

// NewChannelOpenAckTxCmd returns the command to create a MsgChannelOpenAck transaction
func NewChannelOpenAckTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-ack [port-id] [channel-id] [/path/to/proof_try.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenAck message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txf := tx.NewFactoryFromCLI(inBuf).
				WithTxGenerator(txg).
				WithAccountRetriever(ar)
			cliCtx := context.NewCLIContextWithInput(inBuf).WithMarshaler(m)

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

			return tx.GenerateOrBroadcastTx(cliCtx, txf, msg)
		},
	}
	cmd.Flags().String(FlagIBCVersion, "1.0.0", "supported IBC version")
	return cmd
}

// NewChannelOpenConfirmTxCmd returns the command to create a MsgChannelOpenConfirm transaction
func NewChannelOpenConfirmTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	return &cobra.Command{
		Use:   "open-confirm [port-id] [channel-id] [/path/to/proof_ack.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenConfirm message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txf := tx.NewFactoryFromCLI(inBuf).
				WithTxGenerator(txg).
				WithAccountRetriever(ar)
			cliCtx := context.NewCLIContextWithInput(inBuf).WithMarshaler(m)

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

			return tx.GenerateOrBroadcastTx(cliCtx, txf, msg)
		},
	}
}

// NewChannelCloseInitTxCmd returns the command to create a MsgChannelCloseInit transaction
func NewChannelCloseInitTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	return &cobra.Command{
		Use:   "close-init [port-id] [channel-id]",
		Short: "Creates and sends a ChannelCloseInit message",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txf := tx.NewFactoryFromCLI(inBuf).
				WithTxGenerator(txg).
				WithAccountRetriever(ar)
			cliCtx := context.NewCLIContextWithInput(inBuf).WithMarshaler(m)

			portID := args[0]
			channelID := args[1]

			msg := types.NewMsgChannelCloseInit(portID, channelID, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(cliCtx, txf, msg)
		},
	}
}

// NewChannelCloseConfirmTxCmd returns the command to create a MsgChannelCloseConfirm transaction
func NewChannelCloseConfirmTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	return &cobra.Command{
		Use:   "close-confirm [port-id] [channel-id] [/path/to/proof_init.json] [proof-height]",
		Short: "Creates and sends a ChannelCloseConfirm message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txf := tx.NewFactoryFromCLI(inBuf).
				WithTxGenerator(txg).
				WithAccountRetriever(ar)
			cliCtx := context.NewCLIContextWithInput(inBuf).WithMarshaler(m)

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

			return tx.GenerateOrBroadcastTx(cliCtx, txf, msg)
		},
	}
}

func channelOrder() ibctypes.Order {
	if viper.GetBool(FlagOrdered) {
		return ibctypes.ORDERED
	}
	return ibctypes.UNORDERED
}
