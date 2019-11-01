package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	ibcclient "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	FlagNode1    = "node1"
	FlagNode2    = "node2"
	FlagFrom1    = "from1"
	FlagFrom2    = "from2"
	FlagChainID2 = "chain-id2"
)

// GetTxCmd returns the transaction commands for IBC Connections
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ics03ConnectionTxCmd := &cobra.Command{
		Use:   "connection",
		Short: "IBC connection transaction subcommands",
	}

	ics03ConnectionTxCmd.AddCommand(client.PostCommands(
		GetCmdConnectionOpenInit(storeKey, cdc),
		GetCmdConnectionOpenTry(storeKey, cdc),
		GetCmdConnectionOpenAck(storeKey, cdc),
		GetCmdConnectionOpenConfirm(storeKey, cdc),
		GetCmdHandshakeState(storeKey, cdc),
	)...)

	return ics03ConnectionTxCmd
}

// GetCmdConnectionOpenInit defines the command to initialize a connection on
// chain A with a given counterparty chain B
func GetCmdConnectionOpenInit(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: strings.TrimSpace(`open-init [connection-id] [client-id] [counterparty-connection-id] 
		[counterparty-client-id] [path/to/counterparty_prefix.json]`),
		Short: "initialize connection on chain A",
		Long: strings.TrimSpace(
			fmt.Sprintf(`initialize a connection on chain A with a given counterparty chain B:

Example:
$ %s tx ibc connection open-init [connection-id] [client-id] [counterparty-connection-id] 
[counterparty-client-id] [path/to/counterparty_prefix.json]
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

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
$ %s tx ibc connection open-try connection-id] [client-id] 
[counterparty-connection-id] [counterparty-client-id] [path/to/counterparty_prefix.json] 
[counterparty-versions] [path/to/proof_init.json]
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
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
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

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
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
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

func GetCmdHandshakeState(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handshake [conn-id-chain-1] [client-id-chain-1] [path-chain-1] [conn-id-chain-2] [client-id-chain-2] [path-chain-2] ",
		Short: "initiate connection handshake between two chains",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {

			viper.Set(flags.FlagProve, true)

			// --chain-id values for each chain
			cid1 := viper.GetString(flags.FlagChainID)
			cid2 := viper.GetString(FlagChainID2)

			// --from values for each wallet
			from1 := viper.GetString(FlagFrom1)
			from2 := viper.GetString(FlagFrom2)

			// --node values for each RPC
			rpc1 := viper.GetString(FlagNode1)
			rpc2 := viper.GetString(FlagNode2)

			// ibc connection-id for each chain
			connID1 := args[0]
			connID2 := args[3]

			// ibc client-id for each chain
			clientID1 := args[1]
			clientID2 := args[4]

			// Get default version
			version := types.GetCompatibleVersions()[0]

			// Create txbldr, clictx, querier for cid1
			viper.Set(flags.FlagChainID, cid1)
			txBldr1 := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContextIBC(from1, cid1, rpc1).WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)

			// Create txbldr, clictx, querier for cid1
			viper.Set(flags.FlagChainID, cid2)
			txBldr2 := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx2 := context.NewCLIContextIBC(from2, cid2, rpc2).WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)

			// read in path for cid1
			path1, err := parsePath(ctx1.Codec, args[2])
			if err != nil {
				return err
			}

			// read in path for cid2
			path2, err := parsePath(ctx1.Codec, args[5])
			if err != nil {
				return err
			}

			// get passphrase for key from1
			passphrase1, err := keys.GetPassphrase(from1)
			if err != nil {
				return err
			}

			// get passphrase for key from2
			passphrase2, err := keys.GetPassphrase(from2)
			if err != nil {
				return err
			}

			viper.Set(flags.FlagChainID, cid1)
			msgOpenInit := types.NewMsgConnectionOpenInit(
				connID1, clientID1, connID2, clientID2,
				path2, ctx1.GetFromAddress(),
			)

			fmt.Printf("%v <- %-23v", cid1, msgOpenInit.Type())
			res, err := utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgOpenInit}, passphrase1)
			if err != nil || !res.IsOK() {
				return err
			}

			fmt.Printf(" [OK] txid(%v) client(%v) connection(%v)\n", res.TxHash, clientID1, connID1)

			// Another block has to be passed after msgOpenInit is committed
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)

			header, err := tendermint.GetHeader(ctx1)
			if err != nil {
				return err
			}

			// Create and send msgUpdateClient
			viper.Set(flags.FlagChainID, cid2)
			msgUpdateClient := ibcclient.NewMsgUpdateClient(clientID2, header, ctx2.GetFromAddress())
			fmt.Printf("%v <- %-23v", cid2, msgUpdateClient.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgUpdateClient}, passphrase2)
			if err != nil || !res.IsOK() {
				return err
			}
			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, clientID1)

			// Fetch proofs from cid1
			viper.Set(flags.FlagChainID, cid1)
			proofs, err := queryProofs(ctx1.WithHeight(header.Height-1), connID1, storeKey)
			if err != nil {
				return err
			}

			csProof, err := queryConsensusStateProof(ctx1.WithHeight(header.Height-1), clientID1)
			if err != nil {
				return err
			}

			// Create and send msgOpenTry
			viper.Set(flags.FlagChainID, cid2)
			msgOpenTry := types.NewMsgConnectionOpenTry(connID2, clientID2, connID1, clientID1, path1, []string{version}, proofs.Proof, csProof.Proof, uint64(header.Height), uint64(header.Height), ctx2.GetFromAddress())

			fmt.Printf("%v <- %-23v", cid2, msgOpenTry.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgOpenTry}, passphrase2)
			if err != nil || !res.IsOK() {
				return err
			}

			fmt.Printf(" [OK] txid(%v) client(%v) connection(%v)\n", res.TxHash, clientID2, connID2)

			// Another block has to be passed after msgOpenInit is committed
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)

			header, err = tendermint.GetHeader(ctx2)
			if err != nil {
				return err
			}

			// Update the client for cid2 on cid1
			viper.Set(flags.FlagChainID, cid1)
			msgUpdateClient = ibcclient.NewMsgUpdateClient(clientID1, header, ctx1.GetFromAddress())
			fmt.Printf("%v <- %-23v", cid1, msgUpdateClient.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgUpdateClient}, passphrase1)
			if err != nil || !res.IsOK() {
				return err
			}
			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, clientID2)

			// Fetch proofs from cid2
			viper.Set(flags.FlagChainID, cid2)
			proofs, err = queryProofs(ctx2.WithHeight(header.Height-1), connID2, storeKey)
			if err != nil {
				return err
			}

			csProof, err = queryConsensusStateProof(ctx2.WithHeight(header.Height-1), clientID2)
			if err != nil {
				return err
			}

			// Create and send msgOpenAck
			viper.Set(flags.FlagChainID, cid1)
			msgOpenAck := types.NewMsgConnectionOpenAck(connID1, proofs.Proof, csProof.Proof, uint64(header.Height), uint64(header.Height), version, ctx1.GetFromAddress())
			fmt.Printf("%v <- %-23v", cid1, msgOpenAck.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgOpenAck}, passphrase1)
			if err != nil || !res.IsOK() {
				return err
			}
			fmt.Printf(" [OK] txid(%v) connection(%v)\n", res.TxHash, connID1)

			// Another block has to be passed after msgOpenInit is committed
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)

			header, err = tendermint.GetHeader(ctx1)
			if err != nil {
				return err
			}

			// Update client for cid1 on cid2
			viper.Set(flags.FlagChainID, cid2)
			msgUpdateClient = ibcclient.NewMsgUpdateClient(clientID2, header, ctx2.GetFromAddress())
			fmt.Printf("%v <- %-23v", cid2, msgUpdateClient.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgUpdateClient}, passphrase2)
			if err != nil || !res.IsOK() {
				return err
			}
			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, clientID1)

			// Fetch proof from cid1
			viper.Set(flags.FlagChainID, cid1)
			proofs, err = queryProofs(ctx1.WithHeight(header.Height-1), connID1, storeKey)
			if err != nil {
				return err
			}

			// Create and send msgOpenConfirm
			viper.Set(flags.FlagChainID, cid2)
			msgOpenConfirm := types.NewMsgConnectionOpenConfirm(connID2, proofs.Proof, uint64(header.Height), ctx2.GetFromAddress())
			fmt.Printf("%v <- %-23v", cid1, msgOpenConfirm.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgOpenConfirm}, passphrase2)
			if err != nil || !res.IsOK() {
				return err
			}
			fmt.Printf(" [OK] txid(%v) connection(%v)\n", res.TxHash, connID2)

			return nil
		},
	}

	cmd.Flags().String(FlagNode1, "tcp://localhost:26657", "RPC port for the first chain")
	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "RPC port for the second chain")
	cmd.Flags().String(FlagFrom1, "", "key in local keystore for first chain")
	cmd.Flags().String(FlagFrom2, "", "key in local keystore for second chain")
	cmd.Flags().String(FlagChainID2, "", "chain-id for the second chain")

	cmd.MarkFlagRequired(FlagNode1)
	cmd.MarkFlagRequired(FlagNode2)
	cmd.MarkFlagRequired(FlagFrom1)
	cmd.MarkFlagRequired(FlagFrom2)
	cmd.MarkFlagRequired(FlagChainID2)

	return cmd
}

func queryProofs(ctx client.CLIContext, connectionID string, queryRoute string) (types.ConnectionResponse, error) {

	var connRes types.ConnectionResponse

	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  []byte(fmt.Sprintf("connections/%s", connectionID)),
		Prove: true,
	}
	res, err := ctx.QueryABCI(req)
	if err != nil {
		return connRes, err
	}

	var connection types.ConnectionEnd
	if err := ctx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &connection); err != nil {
		return connRes, err
	}
	return types.NewConnectionResponse(connectionID, connection, res.Proof, res.Height), nil
}

func queryConsensusStateProof(ctx client.CLIContext, clientID string) (ibcclient.ConsensusStateResponse, error) {
	var csRes ibcclient.ConsensusStateResponse
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  []byte(fmt.Sprintf("clients/%s/consensusState", clientID)),
		Prove: true,
	}
	res, err := ctx.QueryABCI(req)
	if err != nil {
		return csRes, err
	}

	var cs tendermint.ConsensusState
	if err := ctx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &cs); err != nil {
		return csRes, err
	}
	return ibcclient.NewConsensusStateResponse(clientID, cs, res.Proof, res.Height), nil
}

func parsePath(cdc *codec.Codec, arg string) (commitment.Prefix, error) {
	var path commitment.Prefix
	if err := cdc.UnmarshalJSON([]byte(arg), &path); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
		contents, err := ioutil.ReadFile(arg)
		if err != nil {
			return path, fmt.Errorf("error opening path file: %v\n", err)
		}
		if err := cdc.UnmarshalJSON(contents, &path); err != nil {
			return path, fmt.Errorf("error unmarshalling path file: %v\n", err)
		}
	}
	return path, nil
}
