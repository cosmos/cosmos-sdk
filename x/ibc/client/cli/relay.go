package cli

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

	"github.com/cosmos/cosmos-sdk/x/ibc"
)

// flags
const (
	FlagFromChainID   = "from-chain-id"
	FlagFromChainNode = "from-chain-node"
	FlagToChainID     = "to-chain-id"
	FlagToChainNode   = "to-chain-node"
)

type relayCommander struct {
	cdc       *wire.Codec
	address   sdk.Address
	decoder   sdk.AccountDecoder
	mainStore string
	ibcStore  string

	fromChainID   string
	fromChainNode string
	toChainID     string
	toChainNode   string

	logger log.Logger
}

func IBCRelayCmd(mainStore, ibcStore string, cdc *wire.Codec, dec sdk.AccountDecoder) *cobra.Command {
	cmdr := relayCommander{
		cdc:       cdc,
		decoder:   dec,
		ibcStore:  ibcStore,
		mainStore: mainStore,

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
	c.fromChainID = viper.GetString(FlagFromChainID)
	c.fromChainNode = viper.GetString(FlagFromChainNode)
	c.toChainID = viper.GetString(FlagToChainID)
	c.toChainNode = viper.GetString(FlagToChainNode)
	address, err := context.NewCoreContextFromViper().GetFromAddress()
	if err != nil {
		panic(err)
	}
	c.address = address

	c.loop()
}

func (c relayCommander) loop() {
	ctx := context.NewCoreContextFromViper()
	// get password
	passphrase, err := ctx.GetPassphraseFromStdin(ctx.FromAddressName)
	if err != nil {
		panic(err)
	}
	ingressKey := ibc.IngressSequenceKey(c.fromChainID)
OUTER:
	for {
		time.Sleep(5 * time.Second)

		processedbz, err := query(c.toChainNode, ingressKey, c.ibcStore)
		if err != nil {
			panic(err)
		}

		var processed int64
		if processedbz == nil {
			processed = 0
		} else if err = c.cdc.UnmarshalBinary(processedbz, &processed); err != nil {
			panic(err)
		}

		lengthKey := ibc.EgressLengthKey(c.toChainID)
		egressLengthbz, err := query(c.fromChainNode, lengthKey, c.ibcStore)
		if err != nil {
			c.logger.Error("Error querying outgoing packet list length", "err", err)
			continue OUTER //TODO replace with continue (I think it should just to the correct place where OUTER is now)
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

		seq := c.getSequence(c.toChainNode)

		for i := processed; i < egressLength; i++ {

			egressbz, proofbz, height, err := queryWithProof(c.fromChainNode, ibc.EgressKey(c.toChainID, i), c.ibcStore)
			if err != nil {
				c.logger.Error("Error querying egress packet", "err", err)
				continue OUTER // TODO replace to break, will break first loop then send back to the beginning (aka OUTER)
			}
			fmt.Printf("Got packet from height %d\n", height)
			/*
				commitKey := ibc.CommitByHeightKey(fromChainID, height+1)
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
			viper.Set(client.FlagSequence, seq)
			seq++

			err = c.broadcastTx(c.toChainNode, c.refine(egressbz, proofbz, height+1, i, passphrase))
			if err != nil {
				c.logger.Error("Error broadcasting ingress packet", "err", err)
				continue OUTER // TODO replace to break, will break first loop then send back to the beginning (aka OUTER)
			}

			c.logger.Info("Relayed IBC packet", "number", i)
		}
	}
}

func (c relayCommander) update(height int64) error {
	commit, err := getCommit(c.fromChainNode, height+1)
	if err != nil {
		return fmt.Errorf("Error querying commit: '%s'\n", err)
	}
	fmt.Printf("Commit: %+v\nHeight: %+v\n", commit.Header.AppHash.Bytes(), commit.Header.Height)
	_ = ibc.UpdateChannelMsg{
		SrcChain: c.fromChainID,
		Commit:   commit,
		Signer:   c.address,
	}
	//name := viper.GetString(client.FlagName)

	//_, err = builder.SignBuildBroadcast(name, passphrase, msg, c.cdc)

	return nil
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

func setSequence(seq int64) {
	viper.Set(client.FlagSequence, seq)
}

func (c relayCommander) refine(bz []byte, pbz []byte, sequence int64, height int64, passphrase string) []byte {
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

	msg := ibc.ReceiveMsg{
		Packet: packet,
		PacketProof: ibc.PacketProof{
			Proof:    eproof,
			Height:   height,
			Sequence: sequence,
		},
		Relayer: c.address,
	}

	name := viper.GetString(client.FlagName)

	orig := viper.GetString(client.FlagChainID)
	viper.Set(client.FlagChainID, c.toChainID)
	res, err := context.NewCoreContextFromViper().WithChainID(c.toChainID).SignAndBuild(name, passphrase, msg, c.cdc)

	if err != nil {
		panic(err)
	}
	viper.Set(client.FlagChainID, orig)

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
