package cli

import (
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
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

func handshake(ctx context.CLIContext, cdc *codec.Codec, storeKey string, prefix []byte, connid string) (connection.HandshakeObject, error) {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	climan := client.NewManager(base)
	man := connection.NewHandshaker(connection.NewManager(base, climan))
	return man.CLIQuery(ctx, connid)
}

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

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection",
		Short: "IBC connection transaction subcommands",
	}

	cmd.AddCommand(
		GetCmdHandshake(storeKey, cdc),
	)

	return cmd
}

func GetCmdHandshake(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handshake",
		Short: "initiate connection handshake between two chains",
		Args:  cobra.ExactArgs(4),
		// Args: []string{connid1, connfilepath1, connid2, connfilepath2}
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContext().
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode1)).
				WithFrom(viper.GetString(FlagFrom1))

			ctx2 := context.NewCLIContext().
				WithCodec(cdc).
				WithNodeURI(viper.GetString(FlagNode2)).
				WithFrom(viper.GetString(FlagFrom2))

			conn1id := args[0]
			conn1bz, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}
			var conn1 connection.Connection
			if err := cdc.UnmarshalJSON(conn1bz, &conn1); err != nil {
				return err
			}

			obj1, err := handshake(ctx1, cdc, storeKey, version.DefaultPrefix(), conn1id)
			if err != nil {
				return err
			}

			conn2id := args[2]
			conn2bz, err := ioutil.ReadFile(args[3])
			if err != nil {
				return err
			}
			var conn2 connection.Connection
			if err := cdc.UnmarshalJSON(conn2bz, &conn2); err != nil {
				return err
			}

			obj2, err := handshake(ctx2, cdc, storeKey, version.DefaultPrefix(), conn1id)
			if err != nil {
				return err
			}

			// TODO: check state and if not Idle continue existing process
			height, err := lastheight(ctx2)
			if err != nil {
				return err
			}
			nextTimeout := height + 1000 // TODO: parameterize
			msginit := connection.MsgOpenInit{
				ConnectionID:       conn1id,
				Connection:         conn1,
				CounterpartyClient: conn2.Client,
				NextTimeout:        nextTimeout,
				Signer:             ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msginit})
			if err != nil {
				return err
			}

			timeout := nextTimeout
			height, err = lastheight(ctx1)
			if err != nil {
				return err
			}
			nextTimeout = height + 1000
			_, pconn, err := obj1.ConnectionCLI(ctx1)
			if err != nil {
				return err
			}
			_, pstate, err := obj1.StateCLI(ctx1)
			if err != nil {
				return err
			}
			_, ptimeout, err := obj1.NextTimeoutCLI(ctx1)
			if err != nil {
				return err
			}
			_, pcounter, err := obj1.CounterpartyClientCLI(ctx1)
			if err != nil {
				return err
			}

			msgtry := connection.MsgOpenTry{
				ConnectionID:       conn2id,
				Connection:         conn2,
				CounterpartyClient: conn1.Client,
				Timeout:            timeout,
				NextTimeout:        nextTimeout,
				Proofs:             []commitment.Proof{pconn, pstate, ptimeout, pcounter},
				Signer:             ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgtry})
			if err != nil {
				return err
			}

			timeout = nextTimeout
			height, err = lastheight(ctx2)
			if err != nil {
				return err
			}
			nextTimeout = height + 1000
			_, pconn, err = obj2.ConnectionCLI(ctx2)
			if err != nil {
				return err
			}
			_, pstate, err = obj2.StateCLI(ctx2)
			if err != nil {
				return err
			}
			_, ptimeout, err = obj2.NextTimeoutCLI(ctx2)
			if err != nil {
				return err
			}
			_, pcounter, err = obj2.CounterpartyClientCLI(ctx2)
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

			timeout = nextTimeout
			_, pstate, err = obj1.StateCLI(ctx1)
			if err != nil {
				return err
			}
			_, ptimeout, err = obj1.NextTimeoutCLI(ctx1)
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

			return nil
		},
	}

	return cmd
}
