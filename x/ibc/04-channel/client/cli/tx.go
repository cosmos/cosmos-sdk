package cli

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	connectionutils "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// IBC Channel flags
const (
	FlagOrdered    = "ordered"
	FlagIBCVersion = "ibc-version"
	FlagProofEpoch = "proof-epoch"
)

// NewChannelOpenInitCmd returns the command to create a MsgChannelOpenInit transaction
func NewChannelOpenInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-init [port-id] [channel-id] [counterparty-port-id] [counterparty-channel-id] [connection-hops]",
		Short: "Creates and sends a ChannelOpenInit message",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			portID := args[0]
			channelID := args[1]
			counterpartyPortID := args[2]
			counterpartyChannelID := args[3]
			hops := strings.Split(args[4], "/")
			order := channelOrder(cmd.Flags())
			version, _ := cmd.Flags().GetString(FlagIBCVersion)

			msg := types.NewMsgChannelOpenInit(
				portID, channelID, version, order, hops,
				counterpartyPortID, counterpartyChannelID, clientCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagOrdered, true, "Pass flag for opening ordered channels")
	cmd.Flags().String(FlagIBCVersion, ibctransfertypes.Version, "IBC application version")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewChannelOpenTryCmd returns the command to create a MsgChannelOpenTry transaction
func NewChannelOpenTryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-try [port-id] [channel-id] [counterparty-port-id] [counterparty-channel-id] [connection-hops] [/path/to/proof_init.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenTry message",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			portID := args[0]
			channelID := args[1]
			counterpartyPortID := args[2]
			counterpartyChannelID := args[3]
			hops := strings.Split(args[4], "/")
			order := channelOrder(cmd.Flags())

			// TODO: Differentiate between channel and counterparty versions.
			version, _ := cmd.Flags().GetString(FlagIBCVersion)

			proofInit, err := connectionutils.ParseProof(clientCtx.LegacyAmino, args[5])
			if err != nil {
				return err
			}

			proofHeight, err := strconv.ParseInt(args[6], 10, 64)
			if err != nil {
				return err
			}

			proofEpoch, err := cmd.Flags().GetInt(FlagProofEpoch)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenTry(
				portID, channelID, version, order, hops,
				counterpartyPortID, counterpartyChannelID, version,
				proofInit, uint64(proofEpoch), uint64(proofHeight), clientCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagOrdered, true, "Pass flag for opening ordered channels")
	cmd.Flags().Int(FlagProofEpoch, 0, "epoch for proof height")
	cmd.Flags().String(FlagIBCVersion, ibctransfertypes.Version, "IBC application version")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewChannelOpenAckCmd returns the command to create a MsgChannelOpenAck transaction
func NewChannelOpenAckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-ack [port-id] [channel-id] [/path/to/proof_try.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenAck message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			portID := args[0]
			channelID := args[1]

			// TODO: Differentiate between channel and counterparty versions.
			version, _ := cmd.Flags().GetString(FlagIBCVersion)

			proofTry, err := connectionutils.ParseProof(clientCtx.LegacyAmino, args[5])
			if err != nil {
				return err
			}

			proofHeight, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			proofEpoch, err := cmd.Flags().GetInt(FlagProofEpoch)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenAck(
				portID, channelID, version, proofTry, uint64(proofEpoch), uint64(proofHeight), clientCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Int(FlagProofEpoch, 0, "epoch for proof height")
	cmd.Flags().String(FlagIBCVersion, ibctransfertypes.Version, "IBC application version")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewChannelOpenConfirmCmd returns the command to create a MsgChannelOpenConfirm transaction
func NewChannelOpenConfirmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-confirm [port-id] [channel-id] [/path/to/proof_ack.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenConfirm message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			portID := args[0]
			channelID := args[1]

			proofAck, err := connectionutils.ParseProof(clientCtx.LegacyAmino, args[5])
			if err != nil {
				return err
			}

			proofHeight, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			proofEpoch, err := cmd.Flags().GetInt(FlagProofEpoch)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenConfirm(
				portID, channelID, proofAck, uint64(proofEpoch), uint64(proofHeight), clientCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Int(FlagProofEpoch, 0, "epoch for proof height")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewChannelCloseInitCmd returns the command to create a MsgChannelCloseInit transaction
func NewChannelCloseInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close-init [port-id] [channel-id]",
		Short: "Creates and sends a ChannelCloseInit message",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			portID := args[0]
			channelID := args[1]

			msg := types.NewMsgChannelCloseInit(portID, channelID, clientCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewChannelCloseConfirmCmd returns the command to create a MsgChannelCloseConfirm transaction
func NewChannelCloseConfirmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close-confirm [port-id] [channel-id] [/path/to/proof_init.json] [proof-height]",
		Short: "Creates and sends a ChannelCloseConfirm message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			portID := args[0]
			channelID := args[1]

			proofInit, err := connectionutils.ParseProof(clientCtx.LegacyAmino, args[5])
			if err != nil {
				return err
			}

			proofHeight, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			proofEpoch, err := cmd.Flags().GetInt(FlagProofEpoch)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelCloseConfirm(
				portID, channelID, proofInit, uint64(proofEpoch), uint64(proofHeight), clientCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Int(FlagProofEpoch, 0, "epoch for proof height")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func channelOrder(fs *pflag.FlagSet) types.Order {
	if ordered, _ := fs.GetBool(FlagOrdered); ordered {
		return types.ORDERED
	}

	return types.UNORDERED
}
