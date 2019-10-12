package cli

import (
	"io/ioutil"
	"time"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
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
	FlagChainID2 = "chain-id2"
)

func handshakeState(q storestate.ABCIQuerier, cdc *codec.Codec, storeKey string, prefix []byte, connid string) (connection.HandshakeState, error) {
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
		GetCmdHandshakeState(storeKey, cdc),
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

// TODO: Modify TxResponse type to have .IsOK method that allows for checking proper response codes
func GetCmdHandshakeState(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handshake [conn-id-chain-1] [client-id-chain-1] [path-chain-1] [conn-id-chain-2] [client-id-chain-2] [path-chain-2] ",
		Short: "initiate connection handshake between two chains",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			// --chain-id values for each chain
			cid1 := viper.GetString(flags.FlagChainID)
			cid2 := viper.GetString(FlagChainID2)

			// --from values for each wallet
			from1 := viper.GetString(FlagFrom1)
			from2 := viper.GetString(FlagFrom2) 

			// --node values for each RPC
			rpc1 := viper.GetString(FlagNode1)
			rpc2 := viper.GetString(FlagNode2)

			// ibc connection-id for each chain
			connID1 := args[0]
			connID2 := args[3]

			// ibc client-id for each chain
			clientID1 := args[1]
			clientID2 := args[4]
			
			// Create txbldr, clictx, querier for cid1
			viper.Set(flags.FlagChainID, cid1)
			txBldr1 := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx1 := context.NewCLIContextIBC(from1, cid1, rpc1).WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)
			q1 := storestate.NewCLIQuerier(ctx1)
			
			// Create txbldr, clictx, querier for cid1
			viper.Set(flags.FlagChainID, cid2)
			txBldr2 := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			ctx2 := context.NewCLIContextIBC(from2, cid2, rpc2).WithCodec(cdc).
				WithBroadcastMode(flags.BroadcastBlock)
			q2 := storestate.NewCLIQuerier(ctx2)

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

			// read in path for cid1
			path1, err := parsePath(ctx1.Codec, args[2])
			if err != nil {
				return err 
			}

			// read in path for cid2
			path2, err := parsePath(ctx1.Codec, args[5])
			if err != nil {
				return err 
			}

			// Create connection objects for each chain
			conn1 := connection.NewConnection(clientID1, connID2, path1)
			conn2 := connection.NewConnection(clientID2, connID1, path2)

			// Fetch handshake state object for cid1
			hs1, err := handshakeState(q1, cdc, storeKey, version.DefaultPrefix(), connID1)
			if err != nil {
				return err
			}
			
			// Fetch handshake state object for cid2
			hs2, err := handshakeState(q2, cdc, storeKey, version.DefaultPrefix(), connID2)
			if err != nil {
				return err
			}

			// TODO: check state and if not Idle continue existing process
			// Create and send msgOpenInit
			viper.Set(flags.FlagChainID, cid1)
			msgOpenInit := connection.NewMsgOpenInit(connID1, conn1, conn2.Client, 0, ctx1.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid1, msgOpenInit.Type())
			res, err := utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgOpenInit}, passphrase1)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) client(%v) conn(%v)\n", res.TxHash, conn2.Client, connID1)
			
			// Another block has to be passed after msgOpenInit is committed
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)
			
			header, err := getHeader(ctx1)
			if err != nil {
				return err
			}
			
			// Create and send msgUpdateClient
			viper.Set(flags.FlagChainID, cid2)
			msgUpdateClient := client.NewMsgUpdateClient(conn2.Client, header, ctx2.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid2, msgUpdateClient.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgUpdateClient}, passphrase2)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, conn2.Client)
			
			// Fetch proofs from cid1
			viper.Set(flags.FlagChainID, cid1)
			q1 = storestate.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))
			proofs, err := queryProofs(hs1, q1)
			if err != nil {
				return err
			}
			
			// Create and send msgOpenTry
			viper.Set(flags.FlagChainID, cid2)
			msgOpenTry := connection.NewMsgOpenTry(connID2, conn2, conn1.Client, 0, 0, proofs, uint64(header.Height),ctx2.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid2, msgOpenTry.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgOpenTry}, passphrase2)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) client(%v) connection(%v)\n", res.TxHash, conn1.Client, connID2)
			
			// Another block has to be passed after msgOpenInit is committed
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)
			
			header, err = getHeader(ctx2)
			if err != nil {
				return err
			}
			
			// Update the client for cid2 on cid1
			viper.Set(flags.FlagChainID, cid1)
			msgUpdateClient = client.NewMsgUpdateClient(conn1.Client, header, ctx1.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid1, msgUpdateClient.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgUpdateClient}, passphrase1)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, conn1.Client)
			
			// Fetch proofs from cid2
			viper.Set(flags.FlagChainID, cid2)
			q2 = storestate.NewCLIQuerier(ctx2.WithHeight(header.Height - 1))
			proofs, err = queryProofs(hs2, q2)
			if err != nil {
				return err
			}
			
			// Create and send msgOpenAck
			viper.Set(flags.FlagChainID, cid1)
			msgOpenAck := connection.NewMsgOpenAck(connID1, 0, 0, proofs, uint64(header.Height), ctx1.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid1, msgOpenAck.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgOpenAck}, passphrase1)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) connection(%v)\n", res.TxHash, connID1)
			
			// Another block has to be passed after msgOpenInit is committed
			// to retrieve the correct proofs
			// TODO: Modify this to actually check two blocks being processed, and
			// remove hardcoding this to 8 seconds.
			time.Sleep(8 * time.Second)
			
			header, err = getHeader(ctx1)
			if err != nil {
				return err
			}
			
			// Update client for cid1 on cid2
			viper.Set(flags.FlagChainID, cid2)
			msgUpdateClient = client.NewMsgUpdateClient(conn2.Client, header, ctx2.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid2, msgUpdateClient.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgUpdateClient}, passphrase2)
			if err != nil || !res.IsOK() {
				fmt.Println(res)
				return err
			}
			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, conn2.Client)
			
			// Fetch proof from cid1
			viper.Set(flags.FlagChainID, cid1)
			q1 = storestate.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))
			_, pstate, err := hs1.StageCLI(q1)
			if err != nil {
				return err
			}
			
			// Create and send msgOpenConfirm
			viper.Set(flags.FlagChainID, cid2)
			msgOpenConfirm := connection.NewMsgOpenConfirm(connID2, 0, []commitment.Proof{pstate}, uint64(header.Height), ctx2.GetFromAddress())
			fmt.Printf("%v <- %-14v", cid1, msgOpenConfirm.Type())
			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgOpenConfirm}, passphrase2)
			if err != nil || !res.IsOK() {
				return err
			}
			fmt.Printf(" [OK] txid(%v) connection(%v)\n", res.TxHash, connID2)

			return nil
		},
	}

	cmd.Flags().String(FlagNode1, "tcp://localhost:26657", "RPC port for the first chain")
	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "RPC port for the second chain")
	cmd.Flags().String(FlagFrom1, "", "key in local keystore for first chain")
	cmd.Flags().String(FlagFrom2, "", "key in local keystore for second chain")
	cmd.Flags().String(FlagChainID2, "", "chain-id for the second chain")

	cmd.MarkFlagRequired(FlagNode1)
	cmd.MarkFlagRequired(FlagNode2)
	cmd.MarkFlagRequired(FlagFrom1)
	cmd.MarkFlagRequired(FlagFrom2)
	cmd.MarkFlagRequired(FlagChainID2)

	return cmd
}


func parsePath(cdc *codec.Codec, arg string) (commitment.Prefix, error) {
	var path commitment.Prefix
	if err := cdc.UnmarshalJSON([]byte(arg), &path); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
		contents, err := ioutil.ReadFile(arg)
		if err != nil {
			return path, fmt.Errorf("error opening path file: %v\n", err)
		}
		if err := cdc.UnmarshalJSON(contents, &path); err != nil {
			return path, fmt.Errorf("error unmarshalling path file: %v\n", err)
		}
	}
	return path, nil
}

func queryProofs(hs connection.HandshakeState, q storestate.CLIQuerier) ([]commitment.Proof, error) {
	_, pconn, err := hs.ConnectionCLI(q)
	if err != nil {
		return nil, err
	}
	_, pstate, err := hs.StageCLI(q)
	if err != nil {
		return nil, err
	}
	_, pcounter, err := hs.CounterpartyClientCLI(q)
	if err != nil {
		return nil, err
	}
	return []commitment.Proof{pconn, pstate, pcounter}, nil
}