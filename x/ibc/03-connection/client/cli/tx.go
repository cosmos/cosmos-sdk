package cli

import (
	"fmt"
	"io/ioutil"
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

func handshake(q state.ABCIQuerier, cdc *codec.Codec, storeKey string, prefix []byte, connid string) (connection.HandshakeObject, error) {
	base := state.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	climan := client.NewManager(base)
	man := connection.NewHandshaker(connection.NewManager(base, climan))
	return man.CLIQuery(q, connid)
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
		// Args: []string{connid1, clientid1, path1, connid2, clientid2, connfilepath2}
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

			connid1 := args[0]
			clientid1 := args[1]
			connid2 := args[3]
			clientid2 := args[4]

			var path1 commitment.Path
			path1bz, err := ioutil.ReadFile(args[2])
			if err != nil {
				return err
			}
			if err = cdc.UnmarshalJSON(path1bz, &path1); err != nil {
				return err
			}
			conn1 := connection.Connection{
				Client:       clientid1,
				Counterparty: connid2,
				Path:         path1,
			}

			obj1, err := handshake(q1, cdc, storeKey, version.DefaultPrefix(), connid1)
			if err != nil {
				return err
			}

			var path2 commitment.Path
			path2bz, err := ioutil.ReadFile(args[5])
			if err != nil {
				return err
			}
			if err = cdc.UnmarshalJSON(path2bz, &path2); err != nil {
				return err
			}
			conn2 := connection.Connection{
				Client:       clientid2,
				Counterparty: connid1,
				Path:         path2,
			}

			obj2, err := handshake(q2, cdc, storeKey, version.DefaultPrefix(), connid2)
			if err != nil {
				return err
			}

			// TODO: check state and if not Idle continue existing process
			msginit := connection.MsgOpenInit{
				ConnectionID:       connid1,
				Connection:         conn1,
				CounterpartyClient: conn2.Client,
				Signer:             ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msginit})
			if err != nil {
				return err
			}

			// Another block has to be passed after msginit is committed
			// to retrieve the correct proofs
			time.Sleep(8 * time.Second)

			header, err := getHeader(ctx1)
			if err != nil {
				return err
			}

			msgupdate := client.MsgUpdateClient{
				ClientID: conn2.Client,
				Header:   header,
				Signer:   ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgupdate})
			if err != nil {
				return err
			}

			fmt.Printf("updated apphash to %X\n", header.AppHash)

			q1 = state.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))
			fmt.Printf("querying from %d\n", header.Height-1)

			_, pconn, err := obj1.ConnectionCLI(q1)
			if err != nil {
				return err
			}
			_, pstate, err := obj1.StateCLI(q1)
			if err != nil {
				return err
			}
			_, pcounter, err := obj1.CounterpartyClientCLI(q1)
			if err != nil {
				return err
			}

			msgtry := connection.MsgOpenTry{
				ConnectionID:       connid2,
				Connection:         conn2,
				CounterpartyClient: conn1.Client,
				Proofs:             []commitment.Proof{pconn, pstate, pcounter},
				Height:             uint64(header.Height),
				Signer:             ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgtry})
			if err != nil {
				return err
			}

			// Another block has to be passed after msginit is committed
			// to retrieve the correct proofs
			time.Sleep(8 * time.Second)

			header, err = getHeader(ctx2)
			if err != nil {
				return err
			}

			msgupdate = client.MsgUpdateClient{
				ClientID: conn1.Client,
				Header:   header,
				Signer:   ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgupdate})
			if err != nil {
				return err
			}

			q2 = state.NewCLIQuerier(ctx2.WithHeight(header.Height - 1))

			_, pconn, err = obj2.ConnectionCLI(q2)
			if err != nil {
				return err
			}
			_, pstate, err = obj2.StateCLI(q2)
			if err != nil {
				return err
			}
			_, pcounter, err = obj2.CounterpartyClientCLI(q2)
			if err != nil {
				return err
			}

			msgack := connection.MsgOpenAck{
				ConnectionID: connid1,
				Proofs:       []commitment.Proof{pconn, pstate, pcounter},
				Height:       uint64(header.Height),
				Signer:       ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgack})
			if err != nil {
				return err
			}

			// Another block has to be passed after msginit is committed
			// to retrieve the correct proofs
			time.Sleep(8 * time.Second)

			header, err = getHeader(ctx1)
			if err != nil {
				return err
			}

			msgupdate = client.MsgUpdateClient{
				ClientID: conn2.Client,
				Header:   header,
				Signer:   ctx2.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgupdate})
			if err != nil {
				return err
			}

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
