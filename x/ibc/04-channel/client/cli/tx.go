package cli

import (
	//"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	//"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	//"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	//"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	//"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
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

/*
func handshake(ctx context.CLIContext, cdc *codec.Codec, storeKey string, version int64, connid, chanid string) channel.CLIHandshakeObject {
	prefix := []byte("v" + strconv.FormatInt(version, 10))
	path := merkle.NewPath([][]byte{[]byte(storeKey)}, prefix)
	base := state.NewBase(cdc, sdk.NewKVStoreKey(storeKey), prefix)
	climan := client.NewManager(base)
	connman := connection.NewManager(base, climan)
	man := channel.NewHandshaker(channel.NewManager(base, connman))
	return man.CLIQuery(ctx, path, connid, chanid)
}
*/

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

func GetCmdChannelHandshake(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel-handshake [connID1] [chanID1] [portID1] [connID2] [chanID2] [portID2]",
		Short: "initiate channel handshake between two chains",
		Args:  cobra.ExactArgs(6),
		// Args: []string{connid1, chanid1, portid1, connid2, chanid2, portid2}
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom1)).
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode1))

			ctx2 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom2)).
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode2))

			conn1id := args[0]
			chan1id := args[1]
			port1id := args[2]
			conn2id := args[3]
			chan2id := args[4]
			port2id := args[5]

			chan1 := channel.Channel{
				Port:             port1id,
				Counterparty:     chan2id,
				CounterpartyPort: port2id,
			}

			chan2 := channel.Channel{
				Port:             port2id,
				Counterparty:     chan1id,
				CounterpartyPort: port1id,
			}

			//obj1 := handshake(ctx1, cdc, storeKey, version.Version, conn1id, chan1id)

			//obj2 := handshake(ctx2, cdc, storeKey, version.Version, conn1id, chan1id)

			// TODO: check state and if not Idle continue existing process
			height, err := lastheight(ctx2)
			if err != nil {
				return err
			}
			nextTimeout := height + 1000 // TODO: parameterize
			msginit := channel.MsgOpenInit{
				ConnectionID: conn1id,
				ChannelID:    chan1id,
				Channel:      chan1,
				NextTimeout:  nextTimeout,
				Signer:       ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msginit})
			if err != nil {
				return err
			}

			time.Sleep(8 * time.Second)

			msginit = channel.MsgOpenInit{
				ConnectionID: conn2id,
				ChannelID:    chan2id,
				Channel:      chan2,
				NextTimeout:  nextTimeout,
				Signer:       ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msginit})
			if err != nil {
				return err
			}

			return nil

			/*
				header, err := clientcli.QueryHeader(ctx1)
				if err != nil {
					return err
				}

				msgupdate := client.MsgUpdateClient{
					ClientID: obj2.Connection.Client.ID,
					Header:   header,
					Signer:   ctx2.GetFromAddress(),
				}

				err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgupdate})
				if err != nil {
					return err
				}

				ctx1 = ctx1.WithHeight(header.Height)

				timeout := nextTimeout
				height, err = lastheight(ctx1)
				if err != nil {
					return err
				}
				nextTimeout = height + 1000
				_, pconn, err := obj1.Channel(ctx1)
				if err != nil {
					return err
				}
				_, pstate, err := obj1.State(ctx1)
				if err != nil {
					return err
				}
				_, ptimeout, err := obj1.NextTimeout(ctx1)
				if err != nil {
					return err
				}

				msgtry := channel.MsgOpenTry{
					ConnectionID: conn2id,
					ChannelID:    chan2id,
					Channel:      conn2,
					Timeout:      timeout,
					NextTimeout:  nextTimeout,
					Proofs:       []commitment.Proof{pconn, pstate, ptimeout},
					Signer:       ctx2.GetFromAddress(),
				}

				err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgtry})
				if err != nil {
					return err
				}

				header, err = clientcli.QueryHeader(ctx2)
				if err != nil {
					return err
				}

				msgupdate = client.MsgUpdateClient{
					ClientID: obj1.Connection.Client.ID,
					Header:   header,
					Signer:   ctx1.GetFromAddress(),
				}

				err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgupdate})
				if err != nil {
					return err
				}

				ctx2 = ctx2.WithHeight(header.Height)

				timeout = nextTimeout
				height, err = lastheight(ctx2)
				if err != nil {
					return err
				}
				nextTimeout = height + 1000
				_, pconn, err = obj2.Channel(ctx2)
				if err != nil {
					return err
				}
				_, pstate, err = obj2.State(ctx2)
				if err != nil {
					return err
				}
				_, ptimeout, err = obj2.NextTimeout(ctx2)
				if err != nil {
					return err
				}

				msgack := channel.MsgOpenAck{
					ConnectionID: conn1id,
					ChannelID:    chan1id,
					Timeout:      timeout,
					Proofs:       []commitment.Proof{pconn, pstate, ptimeout},
					Signer:       ctx1.GetFromAddress(),
				}

				err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgack})
				if err != nil {
					return err
				}

				header, err = clientcli.QueryHeader(ctx1)
				if err != nil {
					return err
				}

				msgupdate = client.MsgUpdateClient{
					ClientID: obj2.Connection.Client.ID,
					Header:   header,
					Signer:   ctx2.GetFromAddress(),
				}

				err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgupdate})
				if err != nil {
					return err
				}

				ctx1 = ctx1.WithHeight(header.Height)

				timeout = nextTimeout
				_, pstate, err = obj1.State(ctx1)
				if err != nil {
					return err
				}
				_, ptimeout, err = obj1.NextTimeout(ctx1)
				if err != nil {
					return err
				}

				msgconfirm := channel.MsgOpenConfirm{
					ConnectionID: conn2id,
					ChannelID:    chan2id,
					Timeout:      timeout,
					Proofs:       []commitment.Proof{pstate, ptimeout},
					Signer:       ctx2.GetFromAddress(),
				}

				err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgconfirm})
				if err != nil {
					return err
				}

				return nil
			*/
		},
	}

	cmd.Flags().String(FlagNode1, "", "")
	cmd.Flags().String(FlagNode2, "", "")
	cmd.Flags().String(FlagFrom1, "", "")
	cmd.Flags().String(FlagFrom2, "", "")

	return cmd
}

func GetCmdRelay(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relay [connid1] [chanid1] [connid2] [chanid2]",
		Short: "relay pakcets between two channels",
		Args:  cobra.ExactArgs(4),
		// Args: []string{connid1, chanid1, chanfilepath1, connid2, chanid2, chanfilepath2}
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx1 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom1)).
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode1))

			ctx2 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom2)).
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode2))

			conn1id, chan1id, conn2id, chan2id := args[0], args[1], args[2], args[3]

			obj1 := object(ctx1, cdc, storeKey, version.Version, conn1id, chan1id)
			obj2 := object(ctx2, cdc, storeKey, version.Version, conn2id, chan2id)

			return relayLoop(cdc, ctx1, ctx2, obj1, obj2, conn1id, chan1id, conn2id, chan2id)
		},
	}

	cmd.Flags().String(FlagNode1, "", "")
	cmd.Flags().String(FlagNode2, "", "")
	cmd.Flags().String(FlagFrom1, "", "")
	cmd.Flags().String(FlagFrom2, "", "")

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
		err = relay(cdc, ctx2, ctx1, obj2, obj1, conn1id, chan1id)
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
