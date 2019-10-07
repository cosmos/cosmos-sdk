package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/version"
)

const (
	FlagNode1 = "node1"
	FlagNode2 = "node2"
	FlagFrom1 = "from1"
	FlagFrom2 = "from2"
)

func handshake(cdc *codec.Codec, storeKey string, prefix []byte, portid, chanid, connid string) channel.HandshakeState {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	climan := client.NewManager(base)
	connman := connection.NewManager(base, climan)
	man := channel.NewHandshaker(channel.NewManager(base, connman))
	return man.CLIState(portid, chanid, []string{connid})
}

func flush(q state.ABCIQuerier, cdc *codec.Codec, storeKey string, prefix []byte, portid, chanid string) (channel.HandshakeState, error) {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	climan := client.NewManager(base)
	connman := connection.NewManager(base, climan)
	man := channel.NewHandshaker(channel.NewManager(base, connman))
	return man.CLIQuery(q, portid, chanid)
}

// TODO: import from connection/client
func conn(q state.ABCIQuerier, cdc *codec.Codec, storeKey string, prefix []byte, connid string) (connection.State, error) {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	clientManager := client.NewManager(base)
	man := connection.NewManager(base, clientManager)
	return man.CLIQuery(q, connid)
}

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "IBC channel transaction subcommands",
	}

	cmd.AddCommand(
		GetCmdHandshake(storeKey, cdc),
		GetCmdFlushPackets(storeKey, cdc),
	)

	return cmd
}

// TODO: move to 02/tendermint
func getHeader(ctx context.CLIContext) (res tendermint.Header, err error) {
	node, err := ctx.GetNode()
	if err != nil {
		return
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return
	}

	height := info.Response.LastBlockHeight
	prevheight := height - 1

	commit, err := node.Commit(&height)
	if err != nil {
		return
	}

	validators, err := node.Validators(&prevheight)
	if err != nil {
		return
	}

	nextvalidators, err := node.Validators(&height)
	if err != nil {
		return
	}

	res = tendermint.Header{
		SignedHeader:     commit.SignedHeader,
		ValidatorSet:     tmtypes.NewValidatorSet(validators.Validators),
		NextValidatorSet: tmtypes.NewValidatorSet(nextvalidators.Validators),
	}

	return
}

func GetCmdHandshake(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handshake",
		Short: "initiate connection handshake between two chains",
		Args:  cobra.ExactArgs(6),
		// Args: []string{portid1, chanid1, connid1, portid2, chanid2, connid2}
		RunE: func(cmd *cobra.Command, args []string) error {
			cid1 := viper.GetString(flags.FlagChainID)
			cid2 := viper.GetString(FlagChainId2)
			fmt.Println("setting cid1")
			viper.Set(flags.FlagChainID, cid1)
			txBldr1 := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContextIBC(viper.GetString(FlagFrom1), cid1, viper.GetString(FlagNode1)).
				WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)
			q1 := state.NewCLIQuerier(ctx1)

			fmt.Println("setting cid2")
			viper.Set(flags.FlagChainID, cid2)
			txBldr2 := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx2 := context.NewCLIContextIBC(viper.GetString(FlagFrom2), cid2, viper.GetString(FlagNode2)).
				WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)
			q2 := state.NewCLIQuerier(ctx2)

			portid1 := args[0]
			chanid1 := args[1]
			connid1 := args[2]
			portid2 := args[3]
			chanid2 := args[4]
			connid2 := args[5]

			chan1 := channel.Channel{
				Counterparty:     chanid2,
				CounterpartyPort: portid2,
				ConnectionHops:   []string{connid1},
			}

			chan2 := channel.Channel{
				Counterparty:     chanid1,
				CounterpartyPort: portid1,
				ConnectionHops:   []string{connid2},
			}

			obj1 := handshake(cdc, storeKey, version.DefaultPrefix(), portid1, chanid1, connid1)
			obj2 := handshake(cdc, storeKey, version.DefaultPrefix(), portid2, chanid2, connid2)

			conn1, _, err := obj1.OriginConnection().ConnectionCLI(q1)
			if err != nil {
				return err
			}
			clientid1 := conn1.Client

			conn2, _, err := obj2.OriginConnection().ConnectionCLI(q2)
			if err != nil {
				return err
			}
			clientid2 := conn2.Client

			fmt.Println("setting cid1")
			viper.Set(flags.FlagChainID, cid1)

			// TODO: check state and if not Idle continue existing process
			msginit := channel.MsgOpenInit{
				PortID:    portid1,
				ChannelID: chanid1,
				Channel:   chan1,
				Signer:    ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr1, []sdk.Msg{msginit})
			if err != nil {
				return err
			}

			// Another block has to be passed after msginit is commited
			// to retrieve the correct proofs
			time.Sleep(8 * time.Second)

			header, err := getHeader(ctx1)
			if err != nil {
				return err
			}

			fmt.Println("setting cid2")
			viper.Set(flags.FlagChainID, cid2)

			msgupdate := client.MsgUpdateClient{
				ClientID: clientid2,
				Header:   header,
				Signer:   ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr2, []sdk.Msg{msgupdate})

			fmt.Printf("updated apphash to %X\n", header.AppHash)

			q1 = state.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))
			fmt.Printf("querying from %d\n", header.Height-1)

			_, pchan, err := obj1.ChannelCLI(q1)
			if err != nil {
				return err
			}
			_, pstate, err := obj1.StageCLI(q1)
			if err != nil {
				return err
			}

			msgtry := channel.MsgOpenTry{
				PortID:    portid2,
				ChannelID: chanid2,
				Channel:   chan2,
				Proofs:    []commitment.Proof{pchan, pstate},
				Height:    uint64(header.Height),
				Signer:    ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr2, []sdk.Msg{msgtry})
			if err != nil {
				return err
			}

			// Another block has to be passed after msginit is commited
			// to retrieve the correct proofs
			time.Sleep(8 * time.Second)

			header, err = getHeader(ctx2)
			if err != nil {
				return err
			}

			fmt.Println("setting cid1")
			viper.Set(flags.FlagChainID, cid1)

			msgupdate = client.MsgUpdateClient{
				ClientID: clientid1,
				Header:   header,
				Signer:   ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr1, []sdk.Msg{msgupdate})

			q2 = state.NewCLIQuerier(ctx2.WithHeight(header.Height - 1))

			_, pchan, err = obj2.ChannelCLI(q2)
			if err != nil {
				return err
			}
			_, pstate, err = obj2.StageCLI(q2)
			if err != nil {
				return err
			}

			msgack := channel.MsgOpenAck{
				PortID:    portid1,
				ChannelID: chanid1,
				Proofs:    []commitment.Proof{pchan, pstate},
				Height:    uint64(header.Height),
				Signer:    ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr1, []sdk.Msg{msgack})
			if err != nil {
				return err
			}

			// Another block has to be passed after msginit is commited
			// to retrieve the correct proofs
			time.Sleep(8 * time.Second)

			header, err = getHeader(ctx1)
			if err != nil {
				return err
			}

			fmt.Println("setting cid2")
			viper.Set(flags.FlagChainID, cid2)

			msgupdate = client.MsgUpdateClient{
				ClientID: clientid2,
				Header:   header,
				Signer:   ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr2, []sdk.Msg{msgupdate})

			q1 = state.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))

			_, pstate, err = obj1.StageCLI(q1)
			if err != nil {
				return err
			}

			msgconfirm := channel.MsgOpenConfirm{
				PortID:    portid2,
				ChannelID: chanid2,
				Proofs:    []commitment.Proof{pstate},
				Height:    uint64(header.Height),
				Signer:    ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr2, []sdk.Msg{msgconfirm})
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String(FlagNode1, "tcp://localhost:26657", "")
	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "")
	cmd.Flags().String(FlagFrom1, "", "")
	cmd.Flags().String(FlagFrom2, "", "")

	cmd.MarkFlagRequired(FlagFrom1)
	cmd.MarkFlagRequired(FlagFrom2)

	return cmd
}

func GetCmdFlushPackets(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flush",
		Short: "flush packets on queue",
		Args:  cobra.ExactArgs(2),
		// Args: []string{portid, chanid}
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom1)).
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode1)).
				WithBroadcastMode(flags.BroadcastBlock)
			q1 := state.NewCLIQuerier(ctx1)

			ctx2 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom2)).
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode2)).
				WithBroadcastMode(flags.BroadcastBlock)
			q2 := state.NewCLIQuerier(ctx2)

			portid1, chanid1 := args[0], args[1]

			obj1, err := flush(q1, cdc, storeKey, version.DefaultPrefix(), portid1, chanid1)
			if err != nil {
				return err
			}

			chan1, _, err := obj1.ChannelCLI(q1)
			if err != nil {
				return err
			}

			portid2, chanid2 := chan1.CounterpartyPort, chan1.Counterparty

			obj2, err := flush(q2, cdc, storeKey, version.DefaultPrefix(), portid2, chanid2)
			if err != nil {
				return err
			}

			chan2, _, err := obj2.ChannelCLI(q2)
			if err != nil {
				return err
			}

			connobj2, err := conn(q2, cdc, storeKey, version.DefaultPrefix(), chan2.ConnectionHops[0])
			if err != nil {
				return err
			}

			conn2, _, err := connobj2.ConnectionCLI(q2)
			if err != nil {
				return err
			}

			client2 := conn2.Client

			seqrecv, _, err := obj2.SeqRecvCLI(q2)
			if err != nil {
				return err
			}

			seqsend, _, err := obj1.SeqSendCLI(q1)
			if err != nil {
				return err
			}

			// SeqRecv is the latest received packet index(0 if not exists)
			// SeqSend is the latest sent packet index (0 if not exists)
			if !(seqsend > seqrecv) {
				return errors.New("no unsent packets")
			}

			// TODO: optimize, don't updateclient if already updated
			header, err := getHeader(ctx1)
			if err != nil {
				return err
			}

			q1 = state.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))

			msgupdate := client.MsgUpdateClient{
				ClientID: client2,
				Header:   header,
				Signer:   ctx2.GetFromAddress(),
			}

			msgs := []sdk.Msg{msgupdate}

			for i := seqrecv + 1; i <= seqsend; i++ {
				packet, proof, err := obj1.PacketCLI(q1, i)
				if err != nil {
					return err
				}

				msg := channel.MsgPacket{
					packet,
					chanid2,
					[]commitment.Proof{proof},
					uint64(header.Height),
					ctx2.GetFromAddress(),
				}

				msgs = append(msgs, msg)
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, msgs)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String(FlagNode1, "tcp://localhost:26657", "")
	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "")
	cmd.Flags().String(FlagFrom1, "", "")
	cmd.Flags().String(FlagFrom2, "", "")

	cmd.MarkFlagRequired(FlagFrom1)
	cmd.MarkFlagRequired(FlagFrom2)

	return cmd

}
