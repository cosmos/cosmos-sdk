package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewConnectionOpenInitCmd defines the command to initialize a connection on
// chain A with a given counterparty chain B
func NewConnectionOpenInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-init [connection-id] [client-id] [counterparty-connection-id] [counterparty-client-id] [path/to/counterparty_prefix.json]",
		Short: "Initialize connection on chain A",
		Long:  "Initialize a connection on chain A with a given counterparty chain B",
		Example: fmt.Sprintf(
			"%s tx %s %s open-init [connection-id] [client-id] [counterparty-connection-id] [counterparty-client-id] [path/to/counterparty_prefix.json]",
			version.AppName, host.ModuleName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			connectionID := args[0]
			clientID := args[1]
			counterpartyConnectionID := args[2]
			counterpartyClientID := args[3]

			counterpartyPrefix, err := utils.ParsePrefix(clientCtx.Codec, args[4])
			if err != nil {
				return err
			}

			msg := types.NewMsgConnectionOpenInit(
				connectionID, clientID, counterpartyConnectionID, counterpartyClientID,
				counterpartyPrefix, clientCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

// NewConnectionOpenTryCmd defines the command to relay a try open a connection on
// chain B
func NewConnectionOpenTryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: strings.TrimSpace(`open-try [connection-id] [client-id]
[counterparty-connection-id] [counterparty-client-id] [path/to/counterparty_prefix.json] 
[counterparty-versions] [path/to/proof_init.json] [path/to/proof_consensus.json]`),
		Short: "initiate connection handshake between two chains",
		Long:  "Initialize a connection on chain A with a given counterparty chain B",
		Example: fmt.Sprintf(
			`%s tx %s %s open-try connection-id] [client-id] \
[counterparty-connection-id] [counterparty-client-id] [path/to/counterparty_prefix.json] \
[counterparty-versions] [path/to/proof_init.json] [path/tp/proof_consensus.json]`,
			version.AppName, host.ModuleName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(8),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			connectionID := args[0]
			clientID := args[1]
			counterpartyConnectionID := args[2]
			counterpartyClientID := args[3]

			counterpartyPrefix, err := utils.ParsePrefix(clientCtx.Codec, args[4])
			if err != nil {
				return err
			}

			// TODO: parse strings?
			counterpartyVersions := args[5]

			proofInit, err := utils.ParseProof(clientCtx.Codec, args[6])
			if err != nil {
				return err
			}

			proofConsensus, err := utils.ParseProof(clientCtx.Codec, args[7])
			if err != nil {
				return err
			}

			proofHeight := uint64(clientCtx.Height)
			consensusHeight, err := lastHeight(clientCtx)
			if err != nil {
				return err
			}

			msg := types.NewMsgConnectionOpenTry(
				connectionID, clientID, counterpartyConnectionID, counterpartyClientID,
				counterpartyPrefix, []string{counterpartyVersions}, proofInit, proofConsensus, proofHeight,
				consensusHeight, clientCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

// NewConnectionOpenAckCmd defines the command to relay the acceptance of a
// connection open attempt from chain B to chain A
func NewConnectionOpenAckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-ack [connection-id] [path/to/proof_try.json] [path/to/proof_consensus.json] [version]",
		Short: "relay the acceptance of a connection open attempt",
		Long:  "Relay the acceptance of a connection open attempt from chain B to chain A",
		Example: fmt.Sprintf(
			"%s tx %s %s open-ack [connection-id] [path/to/proof_try.json] [path/to/proof_consensus.json] [version]",
			version.AppName, host.ModuleName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			connectionID := args[0]

			proofTry, err := utils.ParseProof(clientCtx.Codec, args[1])
			if err != nil {
				return err
			}

			proofConsensus, err := utils.ParseProof(clientCtx.Codec, args[2])
			if err != nil {
				return err
			}

			proofHeight := uint64(clientCtx.Height)
			consensusHeight, err := lastHeight(clientCtx)
			if err != nil {
				return err
			}

			version := args[3]

			msg := types.NewMsgConnectionOpenAck(
				connectionID, proofTry, proofConsensus, proofHeight,
				consensusHeight, version, clientCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

// NewConnectionOpenConfirmCmd defines the command to initialize a connection on
// chain A with a given counterparty chain B
func NewConnectionOpenConfirmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-confirm [connection-id] [path/to/proof_ack.json]",
		Short: "confirm to chain B that connection is open on chain A",
		Long:  "Confirm to chain B that connection is open on chain A",
		Example: fmt.Sprintf(
			"%s tx %s %s open-confirm [connection-id] [path/to/proof_ack.json]",
			version.AppName, host.ModuleName, types.SubModuleName,
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			connectionID := args[0]

			proofAck, err := utils.ParseProof(clientCtx.Codec, args[1])
			if err != nil {
				return err
			}

			proofHeight := uint64(clientCtx.Height)
			if err != nil {
				return err
			}

			msg := types.NewMsgConnectionOpenConfirm(
				connectionID, proofAck, proofHeight, clientCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

// lastHeight util function to get the consensus height from the node
func lastHeight(clientCtx client.Context) (uint64, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return 0, err
	}

	return uint64(info.Response.LastBlockHeight), nil
}
