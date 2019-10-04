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
	storestate "github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/version"
)

const (
	FlagNode1    = "node1"
	FlagNode2    = "node2"
	FlagFrom1    = "from1"
	FlagFrom2    = "from2"
	FlagChainId2 = "chain-id2"
)

func handshake(q storestate.ABCIQuerier, cdc *codec.Codec, storeKey string, prefix []byte, connid string) (connection.HandshakeState, error) {
	base := storestate.NewMapping(sdk.NewKVStoreKey(storeKey), cdc, prefix)
	clientManager := client.NewManager(base)
	man := connection.NewHandshaker(connection.NewManager(base, clientManager))
	return man.CLIQuery(q, connid)
}

func lastHeight(ctx context.CLIContext) (uint64, error) {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			cid1 := viper.GetString(flags.FlagChainID)
			cid2 := viper.GetString(FlagChainId2)
			fmt.Println("setting cid1")
			viper.Set(flags.FlagChainID, cid1)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContextIBC(viper.GetString(FlagFrom1), cid1, viper.GetString(FlagNode1)).
				WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)
			q1 := storestate.NewCLIQuerier(ctx1)

			fmt.Println("setting cid2")
			viper.Set(flags.FlagChainID, cid2)

			ctx2 := context.NewCLIContextIBC(viper.GetString(FlagFrom2), cid2, viper.GetString(FlagNode2)).
				WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)
			q2 := storestate.NewCLIQuerier(ctx2)

			connId1 := args[0]
			clientId1 := args[1]
			connId2 := args[3]
			clientId2 := args[4]

			var path1 commitment.Prefix
			path1bz, err := ioutil.ReadFile(args[2])
			if err != nil {
				return err
			}
			if err = cdc.UnmarshalJSON(path1bz, &path1); err != nil {
				return err
			}
			conn1 := connection.Connection{
				Client:       clientId1,
				Counterparty: connId2,
				Path:         path1,
			}

			obj1, err := handshake(q1, cdc, storeKey, version.DefaultPrefix(), connId1)
			if err != nil {
				return err
			}

			var path2 commitment.Prefix
			path2bz, err := ioutil.ReadFile(args[5])
			if err != nil {
				return err
			}
			if err = cdc.UnmarshalJSON(path2bz, &path2); err != nil {
				return err
			}
			conn2 := connection.Connection{
				Client:       clientId2,
				Counterparty: connId1,
				Path:         path2,
			}

			obj2, err := handshake(q2, cdc, storeKey, version.DefaultPrefix(), connId2)
			if err != nil {
				return err
			}

			// TODO: check state and if not Idle continue existing process
			msgInit := connection.MsgOpenInit{
				ConnectionID:       connId1,
				Connection:         conn1,
				CounterpartyClient: conn2.Client,
				Signer:             ctx1.GetFromAddress(),
			}

			fmt.Println("setting cid1")
			viper.Set(flags.FlagChainID, cid1)
			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgInit})
			if err != nil {
				return err
			}

			// Another block has to be passed after msgInit is committed
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)

			header, err := getHeader(ctx1)
			if err != nil {
				return err
			}

			msgUpdate := client.MsgUpdateClient{
				ClientID: conn2.Client,
				Header:   header,
				Signer:   ctx2.GetFromAddress(),
			}

			fmt.Println("setting cid2")
			viper.Set(flags.FlagChainID, cid2)
			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgUpdate})
			if err != nil {
				return err
			}

			fmt.Println("setting cid1")
			viper.Set(flags.FlagChainID, cid1)
			q1 = storestate.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))
			fmt.Printf("querying from %d\n", header.Height-1)

			_, pconn, err := obj1.ConnectionCLI(q1)
			if err != nil {
				return err
			}
			_, pstate, err := obj1.StageCLI(q1)
			if err != nil {
				return err
			}
			_, pcounter, err := obj1.CounterpartyClientCLI(q1)
			if err != nil {
				return err
			}

			msgTry := connection.MsgOpenTry{
				ConnectionID:       connId2,
				Connection:         conn2,
				CounterpartyClient: conn1.Client,
				Proofs:             []commitment.Proof{pconn, pstate, pcounter},
				Height:             uint64(header.Height),
				Signer:             ctx2.GetFromAddress(),
			}

			fmt.Println("setting cid2")
			viper.Set(flags.FlagChainID, cid2)
			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgTry})
			if err != nil {
				return err
			}

			// Another block has to be passed after msgInit is committed
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)

			header, err = getHeader(ctx2)
			if err != nil {
				return err
			}

			msgUpdate = client.MsgUpdateClient{
				ClientID: conn1.Client,
				Header:   header,
				Signer:   ctx1.GetFromAddress(),
			}

			fmt.Println("setting cid1")
			viper.Set(flags.FlagChainID, cid1)
			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgUpdate})
			if err != nil {
				return err
			}

			fmt.Println("setting cid2")
			viper.Set(flags.FlagChainID, cid2)
			q2 = storestate.NewCLIQuerier(ctx2.WithHeight(header.Height - 1))

			_, pconn, err = obj2.ConnectionCLI(q2)
			if err != nil {
				return err
			}
			_, pstate, err = obj2.StageCLI(q2)
			if err != nil {
				return err
			}
			_, pcounter, err = obj2.CounterpartyClientCLI(q2)
			if err != nil {
				return err
			}

			msgAck := connection.MsgOpenAck{
				ConnectionID: connId1,
				Proofs:       []commitment.Proof{pconn, pstate, pcounter},
				Height:       uint64(header.Height),
				Signer:       ctx1.GetFromAddress(),
			}

			fmt.Println("setting cid1")
			viper.Set(flags.FlagChainID, cid1)
			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr, []sdk.Msg{msgAck})
			if err != nil {
				return err
			}

			// Another block has to be passed after msgInit is committed
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)

			header, err = getHeader(ctx1)
			if err != nil {
				return err
			}

			msgUpdate = client.MsgUpdateClient{
				ClientID: conn2.Client,
				Header:   header,
				Signer:   ctx2.GetFromAddress(),
			}

			fmt.Println("setting cid2")
			viper.Set(flags.FlagChainID, cid2)
			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgUpdate})
			if err != nil {
				return err
			}

			fmt.Println("setting cid1")
			viper.Set(flags.FlagChainID, cid1)
			q1 = storestate.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))

			_, pstate, err = obj1.StageCLI(q1)
			if err != nil {
				return err
			}

			msgConfirm := connection.MsgOpenConfirm{
				ConnectionID: connId2,
				Proofs:       []commitment.Proof{pstate},
				Height:       uint64(header.Height),
				Signer:       ctx2.GetFromAddress(),
			}

			fmt.Println("setting cid2")
			viper.Set(flags.FlagChainID, cid2)
			err = utils.GenerateOrBroadcastMsgs(ctx2, txBldr, []sdk.Msg{msgConfirm})
			if err != nil {
				return err
			}

			return nil
		},
	}

	// TODO: Provide flag description
	cmd.Flags().String(FlagNode1, "tcp://localhost:26657", "RPC port for the first chain")
	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "RPC port for the second chain")
	cmd.Flags().String(FlagFrom1, "", "key in local keystore for first chain")
	cmd.Flags().String(FlagFrom2, "", "key in local keystore for second chain")
	cmd.Flags().String(FlagChainId2, "", "chain-id for the second chain")

	cmd.MarkFlagRequired(FlagFrom1)
	cmd.MarkFlagRequired(FlagFrom2)

	return cmd
}
