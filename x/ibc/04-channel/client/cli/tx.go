/*


func lastheight(ctx context.CLIContext) (uint64, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return 0, err
	}

	return uint64(info.Response.LastBlockHeight), nil
}

func GetCmdRelay(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relay",
		Short: "relay pakcets between two channels",
		Args:  cobra.ExactArgs(4),
		// Args: []string{connid1, chanid1, chanfilepath1, connid2, chanid2, chanfilepath2}
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx1 := context.NewCLIContext().
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode1)).
				WithFrom(viper.GetString(FlagFrom1))

			ctx2 := context.NewCLIContext().
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode2)).
				WithFrom(viper.GetString(FlagFrom2))

			conn1id, chan1id, conn2id, chan2id := args[0], args[1], args[2], args[3]

			obj1 := object(ctx1, cdc, storeKey, ibc.Version, conn1id, chan1id)
			obj2 := object(ctx2, cdc, storeKey, ibc.Version, conn2id, chan2id)

			return relayLoop(cdc, ctx1, ctx2, obj1, obj2, conn1id, chan1id, conn2id, chan2id)
		},
	}

	return cmd
}

func relayLoop(cdc *codec.Codec,
	ctx1, ctx2 context.CLIContext,
	obj1, obj2 channel.CLIObject,
	conn1id, chan1id, conn2id, chan2id string,
) error {
	for {
		// TODO: relay() should be goroutine and return error by channel
		err := relay(cdc, ctx1, ctx2, obj1, obj2, conn2id, chan2id)
		// TODO: relayBetween() should retry several times before halt
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
}

func relay(cdc *codec.Codec, ctxFrom, ctxTo context.CLIContext, objFrom, objTo channel.CLIObject, connidTo, chanidTo string) error {
	txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

	seq, _, err := objTo.SeqRecv(ctxTo)
	if err != nil {
		return err
	}

	sent, _, err := objFrom.SeqSend(ctxFrom)
	if err != nil {
		return err
	}

	for i := seq; i <= sent; i++ {
		packet, proof, err := objFrom.Packet(ctxFrom, seq)
		if err != nil {
			return err
		}

		msg := channel.MsgReceive{
			ConnectionID: connidTo,
			ChannelID:    chanidTo,
			Packet:       packet,
			Proofs:       []commitment.Proof{proof},
			Signer:       ctxTo.GetFromAddress(),
		}

		err = utils.GenerateOrBroadcastMsgs(ctxTo, txBldr, []sdk.Msg{msg})
		if err != nil {
			return err
		}
	}

	return nil
}
*/
package cli

import (
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

/*
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {

}
*/
const (
	FlagNode1 = "node1"
	FlagNode2 = "node2"
	FlagFrom1 = "from1"
	FlagFrom2 = "from2"
)

func handshake(q state.ABCIQuerier, cdc *codec.Codec, storeKey string, prefix []byte, portid, chanid string) (channel.HandshakeObject, error) {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	climan := client.NewManager(base)
	connman := connection.NewManager(base, climan)
	man := channel.NewHandshaker(channel.NewManager(base, connman))
	return man.CLIQuery(q, portid, chanid)
}

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "IBC channel transaction subcommands",
	}

	cmd.AddCommand(
		GetCmdHandshake(storeKey, cdc),
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

			obj1, err := handshake(q1, cdc, storeKey, version.DefaultPrefix(), portid1, chanid1)
			if err != nil {
				return err
			}

			obj2, err := handshake(q2, cdc, storeKey, version.DefaultPrefix(), portid2, chanid2)
			if err != nil {
				return err
			}

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

			// TODO: check state and if not Idle continue existing process
			msginit := channel.MsgOpenInit{
				PortID:    portid1,
				ChannelID: chanid1,
				Channel:   chan1,
				Signer:    ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msginit})
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

			msgupdate := client.MsgUpdateClient{
				ClientID: clientid2,
				Header:   header,
				Signer:   ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgupdate})

			fmt.Printf("updated apphash to %X\n", header.AppHash)

			q1 = state.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))
			fmt.Printf("querying from %d\n", header.Height-1)

			_, pchan, err := obj1.ChannelCLI(q1)
			if err != nil {
				return err
			}
			_, pstate, err := obj1.StateCLI(q1)
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

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgtry})
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

			msgupdate = client.MsgUpdateClient{
				ClientID: clientid1,
				Header:   header,
				Signer:   ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgupdate})

			q2 = state.NewCLIQuerier(ctx2.WithHeight(header.Height - 1))

			_, pchan, err = obj2.ChannelCLI(q2)
			if err != nil {
				return err
			}
			_, pstate, err = obj2.StateCLI(q2)
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

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgack})
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

			msgupdate = client.MsgUpdateClient{
				ClientID: clientid2,
				Header:   header,
				Signer:   ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgupdate})

			q1 = state.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))

			_, pstate, err = obj1.StateCLI(q1)
			if err != nil {
				return err
			}

			msgconfirm := connection.MsgOpenConfirm{
				ConnectionID: connid2,
				Proofs:       []commitment.Proof{pstate},
				Height:       uint64(header.Height),
				Signer:       ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgconfirm})
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
