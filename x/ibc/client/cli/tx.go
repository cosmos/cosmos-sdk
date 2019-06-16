package cli

import (
	"errors"
	"io/ioutil"
	//	"os"

	"github.com/spf13/cobra"

	//	"github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	//	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	//	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/keeper"
)

const (
	FlagStatePath            = "state"
	FlagClientID             = "client-id"
	FlagConnectionID         = "connection-id"
	FlagChannelID            = "channel-id"
	FlagCounterpartyID       = "counterparty-id"
	FlagCounterpartyClientID = "counterparty-client-id"
	FlagSourceNode           = "source-node"
)

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcTxCmd := &cobra.Command{
		Use:                        "ibc",
		Short:                      "IBC transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcTxCmd.AddCommand(client.PostCommands(
		GetCmdCreateClient(cdc),
		GetCmdOpenConnection(cdc),
		//		GetCmdRelay(cdc),
		GetCmdUpdateClient(cdc),
	)...)

	return ibcTxCmd
}

func GetCmdCreateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-client",
		Short: "create new client with a consensus state",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			contents, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var state ibc.ConsensusState
			if err := cdc.UnmarshalJSON(contents, &state); err != nil {
				return err
			}

			msg := ibc.MsgCreateClient{
				ClientID:       args[0],
				ConsensusState: state,
				Signer:         cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

func GetCmdUpdateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-client",
		Short: "update existing client with a header",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			contents, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var header ibc.Header
			if err := cdc.UnmarshalJSON(contents, &header); err != nil {
				return err
			}

			msg := ibc.MsgUpdateClient{
				ClientID: args[0],
				Header:   header,
				Signer:   cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// gaiad tx ibc open-connection --node1 tcp://() --node2 tcp://() clientid connectionid channelid
func GetCmdOpenConnection(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-connection",
		Short: "open connection connection between two chains",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			msg := ibc.MsgOpenConnection{
				ConnectionID:         args[0],
				ClientID:             args[1],
				CounterpartyID:       args[2],
				CounterpartyClientID: args[3],
				Signer:               cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.MarkFlagRequired(FlagConnectionID)
	cmd.MarkFlagRequired(FlagClientID)
	cmd.MarkFlagRequired(FlagCounterpartyID)
	cmd.MarkFlagRequired(FlagCounterpartyClientID)

	return cmd
}

/*
func GetCmdRelay(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relay",
		Short: "relay packets",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			cliCtxSource := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc).
				WithNodeURI(viper.GetString(FlagSourceNode))

			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			log := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

			keeper := ibc.DummyKeeper()
			cdc := cliCtx.Codec

			connid := viper.GetString(FlagConnectionID)
			chanid := viper.GetString(FlagChannelID)

			obj := keeper.Channel.Object(connid, chanid)

			seqbz, _, err := query(cliCtx, obj.Seqrecv.Key())
			if err != nil {
				return err
			}
			seq, err := state.DecodeInt(seqbz, state.Dec)
			if err != nil {
				return err
			}

			sentbz, _, err := query(cliCtxSource, obj.Seqsend.Key())
			if err != nil {
				return err
			}
			sent, err := state.DecodeInt(sentbz, state.Dec)
			if err != nil {
				return err
			}

			if seq == sent {
				log.Info("No packets detected")
				return nil
			}

			log.Info("Relaying packet", "sequence", seq)

			var packet ibc.Packet
			packetbz, proof, err := query(cliCtxSource, obj.Packets.Value(seq).Key())
			if err != nil {
				return err
			}
			cdc.MustUnmarshalBinaryBare(packetbz, &packet)

			msg := ibc.MsgReceive{
				ConnectionID: connid,
				ChannelID:    chanid,
				Packet:       packet,
				Proofs:       []commitment.Proof{proof},
				Signer:       cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.MarkFlagRequired(FlagConnectionID)
	cmd.MarkFlagRequired(FlagClientID)
	cmd.MarkFlagRequired(FlagCounterpartyID)
	cmd.MarkFlagRequired(FlagCounterpartyClientID)

	return cmd
}
*/
/*
// gaiad tx ibc relay
func GetCmdRelay(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relay",
		Short: "relay packets between two chains",
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx1 := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc).
				WithNodeURI(viper.GetString(FlagNode1))

			cliCtx2 := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc).
				WithNodeURI(viper.GetString(FlagNode2))

			relayLoop(cliCtx1, cliCtx2, txBldr, viper.GetString(FlagConnectionID), viper.GetString(FlagChannelID))

			return nil
		},
	}

	return cmd
}
*/
// Copied from client/context/query.go
func query(ctx context.CLIContext, key []byte) ([]byte, merkle.Proof, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, merkle.Proof{}, err
	}

	opts := rpcclient.ABCIQueryOptions{
		Height: ctx.Height,
		Prove:  true,
	}

	result, err := node.ABCIQueryWithOptions("/store/ibc/key", key, opts)
	if err != nil {
		return nil, merkle.Proof{}, err
	}

	resp := result.Response
	if !resp.IsOK() {
		return nil, merkle.Proof{}, errors.New(resp.Log)
	}

	return resp.Value, merkle.Proof{
		Key:   key,
		Proof: resp.Proof,
	}, nil
}

/*
func relayLoop(ctx1, ctx2 context.CLIContext, txBldr auth.TxBuilder, connid, chanid string) error {
	log := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	for {
		time.Sleep(5)
		err := relay(ctx2, ctx1, txBldr, log, connid, chanid)
		if err != nil {
			log.Error("Error on relaying, retrying", "error", err)
		}
		err = relay(ctx1, ctx2, txBldr, log, connid, chanid)
		if err != nil {
			log.Error("Error on relaying, retrying", "error", err)
		}
	}
}

func relay(fromCtx, toCtx context.CLIContext, txBldr auth.TxBuilder, log log.Logger, connid, chanid string) error {
	keeper := ibc.DummyKeeper()
	cdc := fromCtx.Codec

	obj := keeper.Channel.Object(connid, chanid)

	processedbz, _, err := query(toCtx, obj.Seqrecv.Key())
	if err != nil {
		return err
	}
	processed, err := state.DecodeInt(processedbz, state.Dec)
	if err != nil {
		return err
	}

	sentbz, _, err := query(fromCtx, obj.Seqsend.Key())
	if err != nil {
		return err
	}
	sent, err := state.DecodeInt(sentbz, state.Dec)
	if err != nil {
		return err
	}

	log.Info("Detected packets", "processed", processed, "sent", sent)
	for i := processed; i < sent; i++ {
		var packet ibc.Packet
		packetbz, proof, err := query(fromCtx, obj.Packets.Value(i).Key())
		if err != nil {
			return err
		}
		cdc.MustUnmarshalBinaryBare(packetbz, &packet)

		msg := ibc.MsgReceive{
			ConnectionID: connid,
			ChannelID:    chanid,
			Packet:       packet,
			Proofs:       []commitment.Proof{proof},
			Signer:       toCtx.GetFromAddress(),
		}

		err = utils.GenerateOrBroadcastMsgs(toCtx, txBldr, []sdk.Msg{msg})
		if err != nil {
			return err
		}
		log.Info("Relayed packet", "sequence", i)
	}

	return nil
}
*/
