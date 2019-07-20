package cli

import (
	"io/ioutil"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	//	clientcli "github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
	//	"github.com/cosmos/cosmos-sdk/x/ibc/version"
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
func handshake(ctx context.CLIContext, cdc *codec.Codec, storeKey string, version int64, id string) connection.CLIHandshakeObject {
	prefix := []byte("v" + strconv.FormatInt(version, 10))
	path := merkle.NewPath([][]byte{[]byte(storeKey)}, prefix)
	base := state.NewBase(cdc, sdk.NewKVStoreKey(storeKey), prefix)
	climan := client.NewManager(base)
	man := connection.NewHandshaker(connection.NewManager(base, climan))
	return man.CLIQuery(ctx, path, id)
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

func GetCmdConnectionHandshake(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection-handshake [connid1] [clientid1] [merklepath1] [connid2] [clientid2] [merklepath2]",
		Short: "initiate connection handshake between two chains",
		Args:  cobra.ExactArgs(6),
		// Args: []string{connid1, connfilepath1, connid2, connfilepath2}
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom1)).
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode1))

			ctx2 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom2)).
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode2))

			conn1id := args[0]
			client1id := args[1]
			path1bz, err := ioutil.ReadFile(args[2])
			if err != nil {
				return err
			}
			var path1 commitment.Path
			if err := cdc.UnmarshalJSON(path1bz, &path1); err != nil {
				return err
			}

			//			obj1 := handshake(ctx1, cdc, storeKey, version.Version, conn1id)

			conn2id := args[3]
			client2id := args[4]
			path2bz, err := ioutil.ReadFile(args[5])
			if err != nil {
				return err
			}
			var path2 commitment.Path
			if err := cdc.UnmarshalJSON(path2bz, &path2); err != nil {
				return err
			}

			//		obj2 := handshake(ctx2, cdc, storeKey, version.Version, conn1id)

			// TODO: check state and if not Idle continue existing process
			height, err := lastheight(ctx2)
			if err != nil {
				return err
			}
			nextTimeout := height + 1000 // TODO: parameterize
			msginit := connection.MsgOpenInit{
				ConnectionID: conn1id,
				Connection: connection.Connection{
					Client:       client1id,
					Counterparty: conn2id,
					Path:         path1,
				},
				CounterpartyClient: client2id,
				NextTimeout:        nextTimeout,
				Signer:             ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msginit})
			if err != nil {
				return err
			}

			time.Sleep(8 * time.Second)

			height, err = lastheight(ctx2)
			if err != nil {
				return err
			}
			ctx2 = ctx2.WithHeight(int64(height))

			msginit = connection.MsgOpenInit{
				ConnectionID: conn2id,
				Connection: connection.Connection{
					Client:       client2id,
					Counterparty: conn1id,
					Path:         path2,
				},
				CounterpartyClient: client1id,
				//				NextTimeout: nextTimeout,
				Signer: ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msginit})

			return err
			/*
				header, err := clientcli.QueryHeader(ctx1)
				if err != nil {
					return err
				}

				msgupdate := client.MsgUpdateClient{
					ClientID: obj2.Client.ID,
					Header:   header,
					Signer:   ctx2.GetFromAddress(),
				}

				ctx1 = ctx1.WithHeight(header.Height)

				timeout := nextTimeout
				height, err = lastheight(ctx1)
				if err != nil {
					return err
				}
				nextTimeout = height + 1000
				_, pconn, err := obj1.Connection(ctx1)
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
				_, pcounter, err := obj1.CounterpartyClient(ctx1)
				if err != nil {
					return err
				}

				msgtry := connection.MsgOpenTry{
					ConnectionID: conn2id,
					Connection: connection.Connection{
						Client:       client2id,
						Counterparty: conn1id,
						Path:         path2,
					},
					CounterpartyClient: client1id,
					Timeout:            timeout,
					NextTimeout:        nextTimeout,
					Proofs:             []commitment.Proof{pconn, pstate, ptimeout, pcounter},
					Signer:             ctx2.GetFromAddress(),
				}

				err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgupdate, msgtry})
				if err != nil {
					return err
				}

				header, err = clientcli.QueryHeader(ctx2)
				if err != nil {
					return err
				}

				msgupdate = client.MsgUpdateClient{
					ClientID: obj1.Client.ID,
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
				_, pconn, err = obj2.Connection(ctx2)
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
				_, pcounter, err = obj2.CounterpartyClient(ctx2)
				if err != nil {
					return err
				}

				msgack := connection.MsgOpenAck{
					ConnectionID: conn1id,
					Timeout:      timeout,
					NextTimeout:  nextTimeout,
					Proofs:       []commitment.Proof{pconn, pstate, ptimeout, pcounter},
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
					ClientID: obj2.Client.ID,
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

				msgconfirm := connection.MsgOpenConfirm{
					ConnectionID: conn2id,
					Timeout:      timeout,
					Proofs:       []commitment.Proof{pstate, ptimeout},
					Signer:       ctx2.GetFromAddress(),
				}

				err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgconfirm})
				if err != nil {
					return err
				}
			*/
		},
	}

	cmd.Flags().String(FlagNode1, "", "")
	cmd.Flags().String(FlagNode2, "", "")
	cmd.Flags().String(FlagFrom1, "", "")
	cmd.Flags().String(FlagFrom2, "", "")

	return cmd
}
