package commands

import (
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/lite"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/tendermint/iavl"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/builder"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/commands"
	ibcm "github.com/cosmos/cosmos-sdk/x/ibc"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

type relayCommander struct {
	cdc       *wire.Codec
	address   sdk.Address
	parser    sdk.ParseAccount
	mainStore string
	ibcStore  string
}

func IBCRelayCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := relayCommander{
		cdc:       cdc,
		parser:    authcmd.GetParseAccount(cdc),
		ibcStore:  "ibc",
		mainStore: "main",
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
	address, err := builder.GetFromAddress()
	if err != nil {
		panic(err)
	}
	c.address = address

	c.loop(fromChainID, fromChainNode, toChainID, toChainNode)
}

func (c relayCommander) loop(fromChainID, fromChainNode, toChainID, toChainNode string) {
	// get password
	name := viper.GetString(client.FlagName)
	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password to sign with '%s':", name)
	passphrase, err := client.GetPassword(prompt, buf)
	if err != nil {
		panic(err)
	}

	ingressKey := ibcm.IngressSequenceKey(fromChainID)

	processedbz, _, _, err := query(toChainNode, ingressKey, c.ibcStore)
	if err != nil {
		panic(err)
	}

	var processed int64
	if processedbz == nil {
		processed = 0
	} else if err = c.cdc.UnmarshalBinary(processedbz, &processed); err != nil {
		panic(err)
	}

OUTER:
	for {
		time.Sleep(time.Second / 2)

		lengthKey := ibcm.EgressLengthKey(toChainID)
		egressLengthbz, _, _, err := query(fromChainNode, lengthKey, c.ibcStore)
		if err != nil {
			fmt.Printf("Error querying outgoing packet list length: '%s'\n", err)
			continue OUTER
		}
		var egressLength int64
		if egressLengthbz == nil {
			egressLength = 0
		} else if err = c.cdc.UnmarshalBinary(egressLengthbz, &egressLength); err != nil {
			panic(err)
		}
		fmt.Printf("egressLength queried: %d\n", egressLength)

		seq := c.getSequence(toChainNode)

		for i := processed; i < egressLength; i++ {
			egressbz, proofbz, height, err := query(fromChainNode, ibcm.EgressKey(toChainID, i), c.ibcStore)
			if err != nil {
				fmt.Printf("Error querying egress packet: '%s'\n", err)
				continue OUTER
			}

			fmt.Printf("Got packet from height %d\n", height)
			commitKey := ibcm.CommitByHeightKey(fromChainID, height+1)
			exists, _, _, err := query(toChainNode, commitKey, c.ibcStore)
			if err != nil {
				fmt.Printf("Error querying commit: '%s'\n", err)
				continue OUTER
			}
			if exists == nil {
				commit, err := getCommit(fromChainNode, height+1)
				if err != nil {
					fmt.Printf("Error querying commit: '%s'\n", err)
					continue OUTER
				}
				fmt.Printf("Commit: %+v\nHeight: %+v\n", commit.Header.AppHash.Bytes(), commit.Header.Height)
				msg := ibcm.UpdateChannelMsg{
					SrcChain: fromChainID,
					Commit:   commit,
					Signer:   c.address,
				}
				name := viper.GetString(client.FlagName)
				viper.Set(client.FlagSequence, seq)
				seq++
				_, err = builder.SignBuildBroadcast(name, passphrase, msg, c.cdc)
				if err != nil {
					fmt.Printf("Error broadcasting update: '%s'\n", err)
					continue OUTER
				}
			}

			viper.Set(client.FlagSequence, seq)
			seq++
			err = c.broadcastTx(toChainNode, c.refine(egressbz, proofbz, height+1, i, passphrase))
			if err != nil {
				fmt.Printf("Error broadcasting ingress packet: '%s'\n", err)
				continue OUTER
			}

			fmt.Printf("Relayed packet: %d\n", i)
		}

		processed = egressLength
	}
}

func query(nodeAddr string, key []byte, storeName string) (res []byte, proof []byte, height int64, err error) {
	orig := viper.GetString(client.FlagNode)
	viper.Set(client.FlagNode, nodeAddr)

	// copied from sdk/client/builder/builder.go to access to proof
	path := fmt.Sprintf("/%s/key", storeName)
	node, err := client.GetNode()
	if err != nil {
		return
	}
	opts := rpcclient.ABCIQueryOptions{
		Height:  viper.GetInt64(client.FlagHeight),
		Trusted: viper.GetBool(client.FlagTrustNode),
	}
	result, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return
	}
	resp := result.Response
	if resp.Code != uint32(0) {
		err = errors.Errorf("Query failed: (%d) %s", resp.Code, resp.Log)
		return
	}

	viper.Set(client.FlagNode, orig)
	return resp.Value, resp.Proof, resp.Height, err
}

func (c relayCommander) broadcastTx(node string, tx []byte) error {
	orig := viper.GetString(client.FlagNode)
	viper.Set(client.FlagNode, node)
	_, err := builder.BroadcastTx(tx)
	viper.Set(client.FlagNode, orig)
	return err
}

func (c relayCommander) getSequence(node string) int64 {
	res, _, _, err := query(node, c.address, c.mainStore)
	if err != nil {
		panic(err)
	}
	account, err := c.parser(res)
	if err != nil {
		panic(err)
	}

	return account.GetSequence()
}

func (c relayCommander) refine(bz []byte, pbz []byte, height int64, sequence int64, passphrase string) []byte {
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

	fmt.Printf("Proof: %+v\nHeight: %+v\n", eproof.Root(), height)

	msg := ibcm.ReceiveMsg{
		Packet:   packet,
		Proof:    eproof,
		Height:   height,
		Relayer:  c.address,
		Sequence: sequence,
	}

	name := viper.GetString(client.FlagName)
	res, err := builder.SignAndBuild(name, passphrase, msg, c.cdc)
	if err != nil {
		panic(err)
	}
	return res
}

func getCommit(nodeAddr string, height int64) (res lite.FullCommit, err error) {
	node := rpcclient.NewHTTP(nodeAddr, "/websocket")

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
