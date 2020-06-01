package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// Connection Handshake flags
const (
	FlagNode1    = "node1"
	FlagNode2    = "node2"
	FlagFrom1    = "from1"
	FlagFrom2    = "from2"
	FlagChainID2 = "chain-id2"
)

// GetCmdConnectionOpenInit defines the command to initialize a connection on
// chain A with a given counterparty chain B
func GetCmdConnectionOpenInit(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   strings.TrimSpace(`open-init [connection-id] [client-id] [counterparty-connection-id] [counterparty-client-id] [path/to/counterparty_prefix.json]`),
		Short: "initialize connection on chain A",
		Long: strings.TrimSpace(
			fmt.Sprintf(`initialize a connection on chain A with a given counterparty chain B:

Example:
$ %s tx ibc connection open-init [connection-id] [client-id] \
[counterparty-connection-id] [counterparty-client-id] \
[path/to/counterparty_prefix.json]
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			clientCtx := client.NewContextWithInput(inBuf).WithCodec(cdc)

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

			return authclient.GenerateOrBroadcastMsgs(clientCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdConnectionOpenTry defines the command to relay a try open a connection on
// chain B
func GetCmdConnectionOpenTry(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: strings.TrimSpace(`open-try [connection-id] [client-id]
[counterparty-connection-id] [counterparty-client-id] [path/to/counterparty_prefix.json] 
[counterparty-versions] [path/to/proof_init.json]`),
		Short: "initiate connection handshake between two chains",
		Long: strings.TrimSpace(
			fmt.Sprintf(`initialize a connection on chain A with a given counterparty chain B:

Example:
$ %s tx ibc connection open-try connection-id] [client-id] \
[counterparty-connection-id] [counterparty-client-id] [path/to/counterparty_prefix.json] \
[counterparty-versions] [path/to/proof_init.json]
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			clientCtx := client.NewContextWithInput(inBuf).
				WithCodec(cdc).
				WithHeight(viper.GetInt64(flags.FlagHeight))

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

			proofInit, err := utils.ParseProof(clientCtx.Codec, args[1])
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
				counterpartyPrefix, []string{counterpartyVersions}, proofInit, proofInit, proofHeight,
				consensusHeight, clientCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(clientCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdConnectionOpenAck defines the command to relay the acceptance of a
// connection open attempt from chain B to chain A
func GetCmdConnectionOpenAck(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-ack [connection-id] [path/to/proof_try.json] [version]",
		Short: "relay the acceptance of a connection open attempt from chain B to chain A",
		Long: strings.TrimSpace(
			fmt.Sprintf(`relay the acceptance of a connection open attempt from chain B to chain A:

Example:
$ %s tx ibc connection open-ack [connection-id] [path/to/proof_try.json] [version]
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			clientCtx := client.NewContextWithInput(inBuf).WithCodec(cdc)

			connectionID := args[0]

			proofTry, err := utils.ParseProof(clientCtx.Codec, args[1])
			if err != nil {
				return err
			}

			proofHeight := uint64(clientCtx.Height)
			consensusHeight, err := lastHeight(clientCtx)
			if err != nil {
				return err
			}

			version := args[4]

			msg := types.NewMsgConnectionOpenAck(
				connectionID, proofTry, proofTry, proofHeight,
				consensusHeight, version, clientCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(clientCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdConnectionOpenConfirm defines the command to initialize a connection on
// chain A with a given counterparty chain B
func GetCmdConnectionOpenConfirm(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-confirm [connection-id] [path/to/proof_ack.json]",
		Short: "confirm to chain B that connection is open on chain A",
		Long: strings.TrimSpace(
			fmt.Sprintf(`confirm to chain B that connection is open on chain A:

Example:
$ %s tx ibc connection open-confirm [connection-id] [path/to/proof_ack.json]
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			clientCtx := client.NewContextWithInput(inBuf).
				WithCodec(cdc).
				WithHeight(viper.GetInt64(flags.FlagHeight))

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

			return authclient.GenerateOrBroadcastMsgs(clientCtx, txBldr, []sdk.Msg{msg})
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
