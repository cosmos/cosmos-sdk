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
	"github.com/cosmos/cosmos-sdk/client/keys"
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
	FlagNode1    = "node1"
	FlagNode2    = "node2"
	FlagFrom1    = "from1"
	FlagFrom2    = "from2"
	FlagChainId2 = "chain-id2"
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
		GetCmdPullPackets(storeKey, cdc),
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
			// --chain-id values for each chain
			cid1 := viper.GetString(flags.FlagChainID)
			cid2 := viper.GetString(FlagChainId2)

			// --from values for each wallet
			from1 := viper.GetString(FlagFrom1)
			from2 := viper.GetString(FlagFrom2)

			// --node values for each RPC
			node1 := viper.GetString(FlagNode1)
			node2 := viper.GetString(FlagNode2)

			// port IDs
			portid1 := args[0]
			portid2 := args[3]

			// channel IDs
			chanid1 := args[1]
			chanid2 := args[4]

			// connection IDs
			connid1 := args[2]
			connid2 := args[5]

			// Create txbldr, clictx, querier for cid1
			viper.Set(flags.FlagChainID, cid1)
			txBldr1 := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContextIBC(from1, cid1, node1).WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)
			q1 := state.NewCLIQuerier(ctx1)

			// Create txbldr, clictx, querier for cid2
			viper.Set(flags.FlagChainID, cid2)
			txBldr2 := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx2 := context.NewCLIContextIBC(from2, cid2, node2).WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)
			q2 := state.NewCLIQuerier(ctx2)

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

			// Create channel objects
			chan1 := channel.NewChannel(chanid2, portid2, []string{connid1})
			chan2 := channel.NewChannel(chanid1, portid1, []string{connid2})

			// Fetch handshake data chain1
			viper.Set(flags.FlagChainID, cid1)
			obj1 := handshake(cdc, storeKey, version.DefaultPrefix(), portid1, chanid1, connid1)
			conn1, _, err := obj1.OriginConnection().ConnectionCLI(q1)
			if err != nil {
				return err
			}
			clientid1 := conn1.Client

			// Fetch handshake data chain2
			viper.Set(flags.FlagChainID, cid2)
			obj2 := handshake(cdc, storeKey, version.DefaultPrefix(), portid2, chanid2, connid2)
			conn2, _, err := obj2.OriginConnection().ConnectionCLI(q2)
			if err != nil {
				return err
			}
			clientid2 := conn2.Client

			// TODO: check state and if not Idle continue existing process
			viper.Set(flags.FlagChainID, cid1)
			msgOpenInit := channel.NewMsgOpenInit(portid1, chanid1, chan1, ctx1.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid1, msgOpenInit.Type())
			res, err := utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgOpenInit}, passphrase1)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) portid(%v) chanid(%v)\n", res.TxHash, portid1, chanid1)

			// Another block has to be passed after msginit is commited
			// to retrieve the correct proofs
			time.Sleep(8 * time.Second)

			header, err := getHeader(ctx1)
			if err != nil {
				return err
			}

			viper.Set(flags.FlagChainID, cid2)
			msgUpdateClient := client.NewMsgUpdateClient(clientid2, header, ctx2.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid2, msgUpdateClient.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgUpdateClient}, passphrase2)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}

			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, clientid2)

			q1 = state.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))
			_, pchan, err := obj1.ChannelCLI(q1)
			if err != nil {
				return err
			}
			_, pstate, err := obj1.StageCLI(q1)
			if err != nil {
				return err
			}

			msgOpenTry := channel.NewMsgOpenTry(portid2, chanid2, chan2, []commitment.Proof{pchan, pstate}, uint64(header.Height), ctx2.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid2, msgOpenTry.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgOpenTry}, passphrase2)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) portid(%v) chanid(%v)\n", res.TxHash, portid2, chanid2)

			// Another block has to be passed after msginit is commited
			// to retrieve the correct proofs
			time.Sleep(8 * time.Second)

			header, err = getHeader(ctx2)
			if err != nil {
				return err
			}

			viper.Set(flags.FlagChainID, cid1)
			msgUpdateClient = client.NewMsgUpdateClient(clientid1, header, ctx1.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid1, msgUpdateClient.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgUpdateClient}, passphrase1)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, clientid1)

			q2 = state.NewCLIQuerier(ctx2.WithHeight(header.Height - 1))
			_, pchan, err = obj2.ChannelCLI(q2)
			if err != nil {
				return err
			}
			_, pstate, err = obj2.StageCLI(q2)
			if err != nil {
				return err
			}

			msgOpenAck := channel.NewMsgOpenAck(portid1, chanid1, []commitment.Proof{pchan, pstate}, uint64(header.Height), ctx1.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid1, msgOpenAck.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgOpenAck}, passphrase1)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) portid(%v) chanid(%v)\n", res.TxHash, portid1, chanid1)

			// Another block has to be passed after msginit is commited
			// to retrieve the correct proofs
			time.Sleep(8 * time.Second)

			header, err = getHeader(ctx1)
			if err != nil {
				return err
			}

			viper.Set(flags.FlagChainID, cid2)
			msgUpdateClient = client.NewMsgUpdateClient(clientid2, header, ctx2.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid2, msgUpdateClient.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgUpdateClient}, passphrase2)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, clientid2)

			q1 = state.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))
			_, pstate, err = obj1.StageCLI(q1)
			if err != nil {
				return err
			}

			msgOpenConfirm := channel.NewMsgOpenConfirm(portid2, chanid2, []commitment.Proof{pstate}, uint64(header.Height), ctx2.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid2, msgOpenConfirm.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgOpenConfirm}, passphrase2)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) portid(%v) chanid(%v)\n", res.TxHash, portid2, chanid2)

			return nil
		},
	}

	cmd.Flags().String(FlagNode1, "tcp://localhost:26657", "RPC port for the first chain")
	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "RPC port for the second chain")
	cmd.Flags().String(FlagFrom1, "", "key in local keystore for first chain")
	cmd.Flags().String(FlagFrom2, "", "key in local keystore for second chain")
	cmd.Flags().String(FlagChainId2, "", "chain-id for the second chain")

	cmd.MarkFlagRequired(FlagFrom1)
	cmd.MarkFlagRequired(FlagFrom2)

	return cmd
}

func GetCmdPullPackets(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "pull packets from the counterparty channel",
		Args:  cobra.ExactArgs(2),
		// Args: []string{portid, chanid}
		RunE: func(cmd *cobra.Command, args []string) error {
			cid1 := viper.GetString(flags.FlagChainID)
			cid2 := viper.GetString(FlagChainId2)
			node1 := viper.GetString(FlagNode1)
			node2 := viper.GetString(FlagNode2)

			viper.Set(flags.FlagChainID, cid1)

			txBldr1 := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContextIBC(viper.GetString(FlagFrom1), cid1, node1).
				WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)
			q1 := state.NewCLIQuerier(ctx1)

			viper.Set(flags.FlagChainID, cid2)

			ctx2 := context.NewCLIContextIBC(viper.GetString(FlagFrom2), cid2, node2).
				WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)
			q2 := state.NewCLIQuerier(ctx2)

			// txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			// ctx1 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom1)).
			// 	WithCodec(cdc).
			// 	WithNodeURI(viper.GetString(FlagNode1)).
			// 	WithBroadcastMode(flags.BroadcastBlock)
			// q1 := state.NewCLIQuerier(ctx1)₩

			// ctx2 := context.NewCLIContextWithFrom(viper.GetString(FlagFrom2)).
			// 	WithCodec(cdc).
			// 	WithNodeURI(viper.GetString(FlagNode2)).
			// 	WithBroadcastMode(flags.BroadcastBlock)
			// q2 := state.NewCLIQuerier(ctx2)

			viper.Set(flags.FlagChainID, cid1)

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

			viper.Set(flags.FlagChainID, cid2)

			obj2, err := flush(q2, cdc, storeKey, version.DefaultPrefix(), portid2, chanid2)
			if err != nil {
				return err
			}

			connobj1, err := conn(q1, cdc, storeKey, version.DefaultPrefix(), chan1.ConnectionHops[0])
			if err != nil {
				return err
			}

			conn1, _, err := connobj1.ConnectionCLI(q1)
			if err != nil {
				return err
			}

			client1 := conn1.Client

			seqrecv, _, err := obj1.SeqRecvCLI(q1)
			if err != nil {
				return err
			}

			seqsend, _, err := obj2.SeqSendCLI(q2)
			if err != nil {
				return err
			}

			// SeqRecv is the latest received packet index(0 if not exists)
			// SeqSend is the latest sent packet index (0 if not exists)
			if !(seqsend > seqrecv) {
				return errors.New("no unsent packets")
			}

			viper.Set(flags.FlagChainID, cid1)

			// TODO: optimize, don't updateclient if already updated
			header, err := getHeader(ctx2)
			if err != nil {
				return err
			}

			q2 = state.NewCLIQuerier(ctx2.WithHeight(header.Height - 1))

			viper.Set(flags.FlagChainID, cid1)

			msgupdate := client.MsgUpdateClient{
				ClientID: client1,
				Header:   header,
				Signer:   ctx1.GetFromAddress(),
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr1, []sdk.Msg{msgupdate})
			if err != nil {
				return err
			}

			msgs := []sdk.Msg{}

			for i := seqrecv + 1; i <= seqsend; i++ {
				packet, proof, err := obj2.PacketCLI(q2, i)
				if err != nil {
					return err
				}

				msg := channel.MsgPacket{
					packet,
					chanid1,
					[]commitment.Proof{proof},
					uint64(header.Height),
					ctx1.GetFromAddress(),
				}

				msgs = append(msgs, msg)
			}

			err = utils.GenerateOrBroadcastMsgs(ctx1, txBldr1, msgs)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String(FlagNode1, "tcp://localhost:26657", "RPC port for the first chain")
	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "RPC port for the second chain")
	cmd.Flags().String(FlagFrom1, "", "key in local keystore for first chain")
	cmd.Flags().String(FlagFrom2, "", "key in local keystore for second chain")
	cmd.Flags().String(FlagChainId2, "", "chain-id for the second chain")

	cmd.MarkFlagRequired(FlagFrom1)
	cmd.MarkFlagRequired(FlagFrom2)

	return cmd
}
