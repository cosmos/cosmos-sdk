package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			connectionID := args[0]
			clientID := args[1]
			counterpartyConnectionID := args[2]
			counterpartyClientID := args[3]

			bz, err := ioutil.ReadFile(args[4])
			if err != nil {
				return err
			}

			var counterpartyPrefix commitment.Prefix
			if err := cdc.UnmarshalJSON(bz, &counterpartyPrefix); err != nil {
				return err
			}

			msg := types.NewMsgConnectionOpenInit(
				connectionID, clientID, counterpartyConnectionID, counterpartyClientID,
				counterpartyPrefix, cliCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
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
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).
				WithCodec(cdc).
				WithHeight(viper.GetInt64(flags.FlagHeight))

			connectionID := args[0]
			clientID := args[1]
			counterpartyConnectionID := args[2]
			counterpartyClientID := args[3]

			prefixBz, err := ioutil.ReadFile(args[4])
			if err != nil {
				return err
			}

			var counterpartyPrefix commitment.Prefix
			if err := cdc.UnmarshalJSON(prefixBz, &counterpartyPrefix); err != nil {
				return err
			}

			// TODO: parse strings?
			counterpartyVersions := args[5]

			proofBz, err := ioutil.ReadFile(args[6])
			if err != nil {
				return err
			}

			var proofInit commitment.Proof
			if err := cdc.UnmarshalJSON(proofBz, &proofInit); err != nil {
				return err
			}

			proofHeight := uint64(cliCtx.Height)
			consensusHeight, err := lastHeight(cliCtx)
			if err != nil {
				return err
			}

			msg := types.NewMsgConnectionOpenTry(
				connectionID, clientID, counterpartyConnectionID, counterpartyClientID,
				counterpartyPrefix, []string{counterpartyVersions}, proofInit, proofInit, proofHeight,
				consensusHeight, cliCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
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
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			connectionID := args[0]
			proofBz, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var proofTry commitment.Proof
			if err := cdc.UnmarshalJSON(proofBz, &proofTry); err != nil {
				return err
			}

			proofHeight := uint64(cliCtx.Height)
			consensusHeight, err := lastHeight(cliCtx)
			if err != nil {
				return err
			}

			version := args[4]

			msg := types.NewMsgConnectionOpenAck(
				connectionID, proofTry, proofTry, proofHeight,
				consensusHeight, version, cliCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
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
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).
				WithCodec(cdc).
				WithHeight(viper.GetInt64(flags.FlagHeight))

			connectionID := args[0]

			proofBz, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var proofAck commitment.Proof
			if err := cdc.UnmarshalJSON(proofBz, &proofAck); err != nil {
				return err
			}

			proofHeight := uint64(cliCtx.Height)

			msg := types.NewMsgConnectionOpenConfirm(
				connectionID, proofAck, proofHeight, cliCtx.GetFromAddress(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// lastHeight util function to get the consensus height from the node
func lastHeight(cliCtx context.CLIContext) (uint64, error) {
	node, err := cliCtx.GetNode()
	if err != nil {
		return 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return 0, err
	}

	return uint64(info.Response.LastBlockHeight), nil
}

func parsePath(cdc *codec.Codec, arg string) (commitment.Prefix, error) {
	var path commitment.Prefix
	if err := cdc.UnmarshalJSON([]byte(arg), &path); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
		contents, err := ioutil.ReadFile(arg)
		if err != nil {
			return path, errors.Wrap(err, "error opening path file")
		}
		if err := cdc.UnmarshalJSON(contents, &path); err != nil {
			return path, errors.Wrap(err, "error unmarshalling path file")
		}
	}
	return path, nil
}
