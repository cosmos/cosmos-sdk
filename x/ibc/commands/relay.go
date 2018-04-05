package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/iavl"
	"github.com/tendermint/tendermint/lite"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/commands"
	ibcm "github.com/cosmos/cosmos-sdk/x/ibc"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

type relayCommander struct {
	cdc       *wire.Codec
	address   sdk.Address
	decoder   sdk.AccountDecoder
	mainStore string
	ibcStore  string

	logger log.Logger
}

func IBCRelayCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := relayCommander{
		cdc:       cdc,
		decoder:   authcmd.GetAccountDecoder(cdc),
		ibcStore:  "ibc",
		mainStore: "main",

		logger: log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	}

	cmd := &cobra.Command{
		Use: "relay",
		Run: cmdr.runIBCRelay,
	}

	cmd.Flags().String(FlagFromChainID, "", "Chain ID for ibc node to check outgoing packets")
	cmd.Flags().String(FlagFromChainNode, "tcp://localhost:46657", "<host>:<port> to tendermint rpc interface for this chain")
	cmd.Flags().String(FlagToChainID, "", "Chain ID for ibc node to broadcast incoming packets")
	cmd.Flags().String(FlagToChainNode, "tcp://localhost:36657", "<host>:<port> to tendermint rpc interface for this chain")

	cmd.MarkFlagRequired(FlagFromChainID)
	cmd.MarkFlagRequired(FlagFromChainNode)
	cmd.MarkFlagRequired(FlagToChainID)
	cmd.MarkFlagRequired(FlagToChainNode)
	cmd.MarkFlagRequired(client.FlagChainID)

	viper.BindPFlag(FlagFromChainID, cmd.Flags().Lookup(FlagFromChainID))
	viper.BindPFlag(FlagFromChainNode, cmd.Flags().Lookup(FlagFromChainNode))
	viper.BindPFlag(FlagToChainID, cmd.Flags().Lookup(FlagToChainID))
	viper.BindPFlag(FlagToChainNode, cmd.Flags().Lookup(FlagToChainNode))

	return cmd
}

func (c relayCommander) runIBCRelay(cmd *cobra.Command, args []string) {
	fromChainID := viper.GetString(FlagFromChainID)
	fromChainNode := viper.GetString(FlagFromChainNode)
	toChainID := viper.GetString(FlagToChainID)
	toChainNode := viper.GetString(FlagToChainNode)
	address, err := context.NewCoreContextFromViper().GetFromAddress()
	if err != nil {
		panic(err)
	}
	c.address = address

	c.loop(fromChainID, fromChainNode, toChainID, toChainNode)
}

func (c relayCommander) loop(fromChainID, fromChainNode, toChainID, toChainNode string) {
	ctx := context.NewCoreContextFromViper()
	// get password
	passphrase, err := ctx.GetPassphraseFromStdin(ctx.FromAddressName)
	if err != nil {
		panic(err)
	}

	ingressKey := ibcm.IngressSequenceKey(fromChainID)
OUTER:
	for {
		time.Sleep(5 * time.Second)

		processedbz, err := query(toChainNode, ingressKey, c.ibcStore)
		if err != nil {
			panic(err)
		}

		var processed int64
		if processedbz == nil {
			processed = 0
		} else if err = c.cdc.UnmarshalBinary(processedbz, &processed); err != nil {
			panic(err)
		}

		lengthKey := ibcm.EgressLengthKey(toChainID)
		egressLengthbz, err := query(fromChainNode, lengthKey, c.ibcStore)
		if err != nil {
			c.logger.Error("Error querying outgoing packet list length", "err", err)
			continue OUTER
		}
		var egressLength int64
		if egressLengthbz == nil {
			egressLength = 0
		} else if err = c.cdc.UnmarshalBinary(egressLengthbz, &egressLength); err != nil {
			panic(err)
		}
		if egressLength > processed {
			c.logger.Info("Detected IBC packet", "number", egressLength-1)
		}

		seq := c.getSequence(toChainNode)

		for i := processed; i < egressLength; i++ {

			egressbz, proofbz, height, err := queryWithProof(fromChainNode, ibcm.EgressKey(toChainID, i), c.ibcStore)
			if err != nil {
				c.logger.Error("Error querying egress packet", "err", err)
				continue OUTER
			}
			fmt.Printf("Got packet from height %d\n", height)
			/*
				commitKey := ibcm.CommitByHeightKey(fromChainID, height+1)
				exists, err := query(toChainNode, commitKey, c.ibcStore)
				if err != nil {
					fmt.Printf("Error querying commit: '%s'\n", err)
					continue OUTER
				}
				if exists == nil {
					err = update()
					if err != nil {
						fmt.Printf("Error broadcasting update: '%s'\n", err)
						continue OUTER
					}
				}
			*/
			err = c.broadcastTx(toChainNode, c.refine(egressbz, proofbz, height, i, seq, passphrase))
			seq++

			if err != nil {
				c.logger.Error("Error broadcasting ingress packet", "err", err)
				continue OUTER
			}

			c.logger.Info("Relayed IBC packet", "number", i)
		}
	}
}

func update() error {
	commit, err := getCommit(fromChainNode, height+1)
	if err != nil {
		fmt.Printf("Error querying commit: '%s'\n", err)
		continue OUTER
	}
	fmt.Printf("Commit: %+v\nHeight: %+v\n", commit.Header.AppHash.Bytes(), commit.Header.Height)
	_ = ibcm.UpdateChannelMsg{
		SrcChain: fromChainID,
		Commit:   commit,
		Signer:   c.address,
	}
	//name := viper.GetString(client.FlagName)
	viper.Set(client.FlagSequence, seq)
	seq++
	//_, err = builder.SignBuildBroadcast(name, passphrase, msg, c.cdc)

}

func query(node string, key []byte, storeName string) (res []byte, err error) {
	return context.NewCoreContextFromViper().WithNodeURI(node).Query(key, storeName)
}

func queryWithProof(nodeAddr string, key []byte, storeName string) (res []byte, proof []byte, height int64, err error) {
	ctx := context.NewCoreContextFromViper().WithNodeURI(nodeAddr)
	node, err := ctx.GetNode()
	if err != nil {
		return
	}

	opts := rpcclient.ABCIQueryOptions{
		Height:  ctx.Height,
		Trusted: ctx.TrustNode,
	}

	path := fmt.Sprintf("/%s/key", storeName)
	result, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return
	}
	resp := result.Response
	if resp.Code != uint32(0) {
		err = errors.Errorf("Query failed: (%d) %s", resp.Code, resp.Log)
		return
	}
	return resp.Value, resp.Proof, resp.Height, nil
}

func (c relayCommander) broadcastTx(node string, tx []byte) error {
	_, err := context.NewCoreContextFromViper().WithNodeURI(node).BroadcastTx(tx)
	return err
}

func (c relayCommander) getSequence(node string) int64 {
	res, err := query(node, c.address, c.mainStore)
	if err != nil {
		panic(err)
	}

	account, err := c.decoder(res)
	if err != nil {
		panic(err)
	}

	return account.GetSequence()
}
func (c relayCommander) refine(bz []byte, pbz []byte, height int64, packetSeq int64, txSeq int64, passphrase string) []byte {
	var packet ibc.Packet
	if err := c.cdc.UnmarshalBinary(bz, &packet); err != nil {
		panic(err)
	}

	proof, err := iavl.ReadKeyProof(pbz)
	if err != nil {
		panic(err)
	}

	eproof, ok := proof.(*iavl.KeyExistsProof)
	if !ok {
		panic("Expected KeyExistsProof for non-empty value")
	}

	fmt.Printf("Proof: %+v\n", eproof)
	fmt.Printf("ProofRoot: %+v\nHeight: %+v\n", eproof.Root(), height)

	msg := ibcm.ReceiveMsg{
		Packet:   packet,
		Proof:    eproof,
		Height:   height,
		Relayer:  c.address,
		Sequence: packetSeq,
	}

	ctx := context.NewCoreContextFromViper().WithSequence(txSeq)
	res, err := ctx.SignAndBuild(ctx.FromAddressName, passphrase, msg, c.cdc)
	if err != nil {
		panic(err)
	}
	return res
}

func getCommit(nodeAddr string, height int64) (res lite.FullCommit, err error) {
	node, err := context.NewCoreContextFromViper().WithNodeURI(nodeAddr).GetNode()
	if err != nil {
		return
	}

	commit, err := node.Commit(&height)
	if err != nil {
		return
	}
	valset, err := node.Validators(&height)
	if err != nil {
		return
	}

	return lite.NewFullCommit(
		lite.Commit(commit.SignedHeader),
		tmtypes.NewValidatorSet(valset.Validators),
	), nil
}
