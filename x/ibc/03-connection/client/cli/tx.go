package cli

import (
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// const (
// 	FlagNode2 = "node2"
// 	FlagFrom2 = "from2"
// )

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
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			// cliCtx2 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom2)).
			// 	WithCodec(cdc).
			// 	WithNodeURI(viper.GetString(FlagNode2)).
			// 	WithBroadcastMode(flags.BroadcastBlock)

			connectionID := args[0]
			clientID := args[1]
			counterpartyConnectionID := args[2]
			counterpartyClientID := args[3]

			bz, err := ioutil.ReadFile(args[4])
			if err != nil {
				return err
			}

			var counterpartyPrefix ics23.Prefix
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

// GetCmdConnectionOpenTry defines the command to initialize a connection on
// chain A with a given counterparty chain B
func GetCmdConnectionOpenTry(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: strings.TrimSpace(`open-try [connection-id] [client-id] 
[counterparty-connection-id] [counterparty-client-id] [path/to/counterparty_prefix.json] 
[counterparty-versions] [path/to/proof_init.json]`),
		Short: "initiate connection handshake between two chains",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithHeight(viper.GetInt64(flags.FlagHeight))
			// cliCtx2 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom2)).
			// 	WithCodec(cdc).
			// 	WithNodeURI(viper.GetString(FlagNode2)).

			connectionID := args[0]
			clientID := args[1]
			counterpartyConnectionID := args[2]
			counterpartyClientID := args[3]

			prefixBz, err := ioutil.ReadFile(args[4])
			if err != nil {
				return err
			}

			var counterpartyPrefix ics23.Prefix
			if err := cdc.UnmarshalJSON(prefixBz, &counterpartyPrefix); err != nil {
				return err
			}

			// TODO: parse strings?
			counterpartyVersions := args[5]

			proofBz, err := ioutil.ReadFile(args[6])
			if err != nil {
				return err
			}

			var proofInit ics23.Proof
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
				counterpartyPrefix, []string{counterpartyVersions}, proofInit, proofHeight,
				consensusHeight, cliCtx.GetFromAddress(),
			)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// GetCmdConnectionOpenAck defines the command to initialize a connection on
// chain A with a given counterparty chain B
func GetCmdConnectionOpenAck(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-ack [connection-id] [path/to/proof_try.json] [version]",
		Short: "relay the acceptance of a connection open attempt from chain B to chain A",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			// cliCtx2 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom2)).
			// 	WithCodec(cdc).
			// 	WithNodeURI(viper.GetString(FlagNode2)).
			// 	WithBroadcastMode(flags.BroadcastBlock)

			connectionID := args[0]
			proofBz, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var proofTry ics23.Proof
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
				connectionID, proofTry, proofHeight,
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
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithHeight(viper.GetInt64(flags.FlagHeight))
			// cliCtx2 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom2)).
			// WithCodec(cdc).
			// WithNodeURI(viper.GetString(FlagNode2)).
			// WithBroadcastMode(flags.BroadcastBlock)

			connectionID := args[0]

			proofBz, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var proofAck ics23.Proof
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

// func GetCmdHandshake(storeKey string, cdc *codec.Codec) *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "handshake",
// 		Short: "initiate connection handshake between two chains",
// 		Args:  cobra.ExactArgs(6),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
// 			ctx1 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom1)).
// 				WithCodec(cdc).
// 				WithNodeURI(viper.GetString(FlagNode1)).
// 				WithBroadcastMode(flags.BroadcastBlock)
// 			q1 := storestate.NewCLIQuerier(ctx1)

// 			ctx2 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom2)).
// 				WithCodec(cdc).
// 				WithNodeURI(viper.GetString(FlagNode2)).
// 				WithBroadcastMode(flags.BroadcastBlock)
// 			q2 := storestate.NewCLIQuerier(ctx2)

// 			connId1 := args[0]
// 			clientId1 := args[1]
// 			connId2 := args[3]
// 			clientId2 := args[4]

// 			var path1 commitment.Prefix
// 			path1bz, err := ioutil.ReadFile(args[2])
// 			if err != nil {
// 				return err
// 			}
// 			if err = cdc.UnmarshalJSON(path1bz, &path1); err != nil {
// 				return err
// 			}
// 			conn1 := connection.Connection{
// 				Client:       clientId1,
// 				Counterparty: connId2,
// 				Path:         path1,
// 			}

// 			obj1, err := handshake(q1, cdc, storeKey, version.DefaultPrefix(), connId1)
// 			if err != nil {
// 				return err
// 			}

// 			var path2 commitment.Prefix
// 			path2bz, err := ioutil.ReadFile(args[5])
// 			if err != nil {
// 				return err
// 			}
// 			if err = cdc.UnmarshalJSON(path2bz, &path2); err != nil {
// 				return err
// 			}
// 			conn2 := connection.Connection{
// 				Client:       clientId2,
// 				Counterparty: connId1,
// 				Path:         path2,
// 			}

// 			obj2, err := handshake(q2, cdc, storeKey, version.DefaultPrefix(), connId2)
// 			if err != nil {
// 				return err
// 			}

// 			// TODO: check state and if not Idle continue existing process
// 			msgInit := connection.MsgOpenInit{
// 				ConnectionID:       connId1,
// 				Connection:         conn1,
// 				CounterpartyClient: conn2.Client,
// 				Signer:             ctx1.GetFromAddress(),
// 			}

// 			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgInit})
// 			if err != nil {
// 				return err
// 			}

// 			// Another block has to be passed after msgInit is committed
// 			// to retrieve the correct proofs
// 			// TODO: Modify this to actually check two blocks being processed, and
// 			// remove hardcoding this to 8 seconds.
// 			time.Sleep(8 * time.Second)

// 			header, err := getHeader(ctx1)
// 			if err != nil {
// 				return err
// 			}

// 			msgUpdate := client.MsgUpdateClient{
// 				ClientID: conn2.Client,
// 				Header:   header,
// 				Signer:   ctx2.GetFromAddress(),
// 			}

// 			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgUpdate})
// 			if err != nil {
// 				return err
// 			}

// 			q1 = storestate.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))
// 			fmt.Printf("querying from %d\n", header.Height-1)

// 			_, pconn, err := obj1.ConnectionCLI(q1)
// 			if err != nil {
// 				return err
// 			}
// 			_, pstate, err := obj1.StageCLI(q1)
// 			if err != nil {
// 				return err
// 			}
// 			_, pcounter, err := obj1.CounterpartyClientCLI(q1)
// 			if err != nil {
// 				return err
// 			}

// 			msgTry := connection.MsgOpenTry{
// 				ConnectionID:       connId2,
// 				Connection:         conn2,
// 				CounterpartyClient: conn1.Client,
// 				Proofs:             []commitment.Proof{pconn, pstate, pcounter},
// 				Height:             uint64(header.Height),
// 				Signer:             ctx2.GetFromAddress(),
// 			}

// 			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgTry})
// 			if err != nil {
// 				return err
// 			}

// 			// Another block has to be passed after msgInit is committed
// 			// to retrieve the correct proofs
// 			// TODO: Modify this to actually check two blocks being processed, and
// 			// remove hardcoding this to 8 seconds.
// 			time.Sleep(8 * time.Second)

// 			header, err = getHeader(ctx2)
// 			if err != nil {
// 				return err
// 			}

// 			msgUpdate = client.MsgUpdateClient{
// 				ClientID: conn1.Client,
// 				Header:   header,
// 				Signer:   ctx1.GetFromAddress(),
// 			}

// 			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgUpdate})
// 			if err != nil {
// 				return err
// 			}

// 			q2 = storestate.NewCLIQuerier(ctx2.WithHeight(header.Height - 1))

// 			_, pconn, err = obj2.ConnectionCLI(q2)
// 			if err != nil {
// 				return err
// 			}
// 			_, pstate, err = obj2.StageCLI(q2)
// 			if err != nil {
// 				return err
// 			}
// 			_, pcounter, err = obj2.CounterpartyClientCLI(q2)
// 			if err != nil {
// 				return err
// 			}

// 			msgAck := connection.MsgOpenAck{
// 				ConnectionID: connId1,
// 				Proofs:       []commitment.Proof{pconn, pstate, pcounter},
// 				Height:       uint64(header.Height),
// 				Signer:       ctx1.GetFromAddress(),
// 			}

// 			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgAck})
// 			if err != nil {
// 				return err
// 			}

// 			// Another block has to be passed after msgInit is committed
// 			// to retrieve the correct proofs
// 			// TODO: Modify this to actually check two blocks being processed, and
// 			// remove hardcoding this to 8 seconds.
// 			time.Sleep(8 * time.Second)

// 			header, err = getHeader(ctx1)
// 			if err != nil {
// 				return err
// 			}

// 			msgUpdate = client.MsgUpdateClient{
// 				ClientID: conn2.Client,
// 				Header:   header,
// 				Signer:   ctx2.GetFromAddress(),
// 			}

// 			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgUpdate})
// 			if err != nil {
// 				return err
// 			}

// 			q1 = storestate.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))

// 			_, pstate, err = obj1.StageCLI(q1)
// 			if err != nil {
// 				return err
// 			}

// 			msgConfirm := connection.MsgOpenConfirm{
// 				ConnectionID: connId2,
// 				Proofs:       []commitment.Proof{pstate},
// 				Height:       uint64(header.Height),
// 				Signer:       ctx2.GetFromAddress(),
// 			}

// 			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgConfirm})
// 			if err != nil {
// 				return err
// 			}

// 			return nil
// 		},
// 	}
// 	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "")
// 	cmd.Flags().String(FlagFrom2, "", "")

// 	return cmd
// }
