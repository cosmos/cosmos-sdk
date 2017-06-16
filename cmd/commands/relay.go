package commands

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	// 	"github.com/spf13/viper"
	// 	"github.com/tendermint/tmlibs/cli"
	// 	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/go-wire"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/basecoin/plugins/ibc"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"
)

var RelayCmd = &cobra.Command{
	Use:   "relay",
	Short: "Relay ibc packets between two chains",
}

var RelayStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start basecoin relayer to relay IBC packets between chains",
	RunE:  relayStartCmd,
}

var RelayInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Register both chains with each other, to prepare the relayer to run",
	RunE:  relayInitCmd,
}

//flags
var (
	chain1AddrFlag string
	chain2AddrFlag string

	chain1IDFlag string
	chain2IDFlag string

	fromFileFlag string

	genesisFile1Flag string
	genesisFile2Flag string
)

func init() {
	flags := []Flag2Register{
		{&chain1AddrFlag, "chain1-addr", "tcp://localhost:46657", "Node address for chain1"},
		{&chain2AddrFlag, "chain2-addr", "tcp://localhost:36657", "Node address for chain2"},
		{&chain1IDFlag, "chain1-id", "test_chain_1", "ChainID for chain1"},
		{&chain2IDFlag, "chain2-id", "test_chain_2", "ChainID for chain2"},
		{&fromFileFlag, "from", "key.json", "Path to a private key to sign the transaction"},
	}
	RegisterPersistentFlags(RelayCmd, flags)

	initFlags := []Flag2Register{
		{&genesisFile1Flag, "genesis1", "", "Path to genesis file for chain1"},
		{&genesisFile2Flag, "genesis2", "", "Path to genesis file for chain2"},
	}
	RegisterFlags(RelayInitCmd, initFlags)

	RelayCmd.AddCommand(RelayStartCmd)
	RelayCmd.AddCommand(RelayInitCmd)
}

func relayStartCmd(cmd *cobra.Command, args []string) error {
	go loop(chain1AddrFlag, chain2AddrFlag, chain1IDFlag, chain2IDFlag)
	go loop(chain2AddrFlag, chain1AddrFlag, chain2IDFlag, chain1IDFlag)

	cmn.TrapSignal(func() {
		// TODO: Cleanup
	})
	return nil
}

func relayInitCmd(cmd *cobra.Command, args []string) error {
	err := registerChain(chain1IDFlag, chain1AddrFlag, chain2IDFlag, genesisFile2Flag, fromFileFlag)
	if err != nil {
		return err
	}
	err = registerChain(chain2IDFlag, chain2AddrFlag, chain1IDFlag, genesisFile1Flag, fromFileFlag)
	return err
}

func registerChain(chainID, node, registerChainID, registerGenesis, keyFile string) error {
	genesisBytes, err := ioutil.ReadFile(registerGenesis)
	if err != nil {
		return errors.Errorf("Error reading genesis file %v: %v\n", registerGenesis, err)
	}

	ibcTx := ibc.IBCRegisterChainTx{
		ibc.BlockchainGenesis{
			ChainID: registerChainID,
			Genesis: string(genesisBytes),
		},
	}

	privKey, err := LoadKey(keyFile)
	if err != nil {
		return err
	}
	relay := newRelayer(privKey, chainID, node)
	return relay.appTx(ibcTx)
}

func loop(addr1, addr2, id1, id2 string) {
	nextSeq := 0

	// load the priv key
	privKey, err := LoadKey(fromFileFlag)
	if err != nil {
		logger.Error(err.Error())
		cmn.PanicCrisis(err.Error())
	}

	// relay from chain1 to chain2
	thisRelayer := newRelayer(privKey, id2, addr2)

	logger.Info(fmt.Sprintf("Relaying from chain %v on %v to chain %v on %v", id1, addr1, id2, addr2))

	httpClient := client.NewHTTP(addr1, "/websocket")

OUTER:
	for {

		time.Sleep(time.Second)

		// get the latest ibc packet sequence number
		key := fmt.Sprintf("ibc,egress,%v,%v", id1, id2)
		query, err := queryWithClient(httpClient, []byte(key))
		if err != nil {
			logger.Error("Error querying for latest sequence", "key", key, "error", err.Error())
			continue OUTER
		}
		if len(query.Value) == 0 {
			// nothing yet
			continue OUTER
		}

		seq, err := strconv.ParseUint(string(query.Value), 10, 64)
		if err != nil {
			logger.Error("Error parsing sequence number from query", "query.Value", query.Value, "error", err.Error())
			continue OUTER
		}
		seq -= 1 // seq is the packet count. -1 because 0-indexed

		if nextSeq <= int(seq) {
			logger.Info("Got new packets", "last-sequence", nextSeq-1, "new-sequence", seq)
		}

		// get all packets since the last one we relayed
		for ; nextSeq <= int(seq); nextSeq++ {
			key := fmt.Sprintf("ibc,egress,%v,%v,%d", id1, id2, nextSeq)
			query, err := queryWithClient(httpClient, []byte(key))
			if err != nil {
				logger.Error("Error querying for packet", "seqeuence", nextSeq, "key", key, "error", err.Error())
				continue OUTER
			}

			var packet ibc.Packet
			err = wire.ReadBinaryBytes(query.Value, &packet)
			if err != nil {
				logger.Error("Error unmarshalling packet", "key", key, "query.Value", query.Value, "error", err.Error())
				continue OUTER
			}

			proof := new(iavl.IAVLProof)
			err = wire.ReadBinaryBytes(query.Proof, &proof)
			if err != nil {
				logger.Error("Error unmarshalling proof", "query.Proof", query.Proof, "error", err.Error())
				continue OUTER
			}

			// query.Height is actually for the next block,
			// so wait a block before we fetch the header & commit
			if err := waitForBlock(httpClient); err != nil {
				logger.Error("Error waiting for a block", "addr", addr1, "error", err.Error())
				continue OUTER
			}

			// get the header and commit from the height the query was done at
			res, err := httpClient.Commit(int(query.Height))
			if err != nil {
				logger.Error("Error fetching header and commits", "height", query.Height, "error", err.Error())
				continue OUTER
			}

			// update the chain state on the other chain
			updateTx := ibc.IBCUpdateChainTx{
				Header: *res.Header,
				Commit: *res.Commit,
			}
			logger.Info("Updating chain", "src-chain", id1, "height", res.Header.Height, "appHash", res.Header.AppHash)
			if err := thisRelayer.appTx(updateTx); err != nil {
				logger.Error("Error creating/sending IBCUpdateChainTx", "error", err.Error())
				continue OUTER
			}

			// relay the packet and proof
			logger.Info("Relaying packet", "src-chain", id1, "height", query.Height, "sequence", nextSeq)
			postTx := ibc.IBCPacketPostTx{
				FromChainID:     id1,
				FromChainHeight: query.Height,
				Packet:          packet,
				Proof:           proof,
			}

			if err := thisRelayer.appTx(postTx); err != nil {
				logger.Error("Error creating/sending IBCPacketPostTx", "error", err.Error())
				// dont `continue OUTER` here. the error might be eg. Already exists
				// TODO: catch this programmatically ?
			}
		}
	}
}

type relayer struct {
	privKey  *Key
	chainID  string
	nodeAddr string
	client   *client.HTTP
}

func newRelayer(privKey *Key, chainID, nodeAddr string) *relayer {
	httpClient := client.NewHTTP(nodeAddr, "/websocket")
	return &relayer{
		privKey:  privKey,
		chainID:  chainID,
		nodeAddr: nodeAddr,
		client:   httpClient,
	}
}

func (r *relayer) appTx(ibcTx ibc.IBCTx) error {
	acc, err := getAccWithClient(r.client, r.privKey.Address[:])
	if err != nil {
		return err
	}
	sequence := acc.Sequence + 1

	data := []byte(wire.BinaryBytes(struct {
		ibc.IBCTx `json:"unwrap"`
	}{ibcTx}))

	smallCoins := types.Coin{"mycoin", 1}

	input := types.NewTxInput(r.privKey.PubKey, types.Coins{smallCoins}, sequence)
	tx := &types.AppTx{
		Gas:   0,
		Fee:   smallCoins,
		Name:  "IBC",
		Input: input,
		Data:  data,
	}

	tx.Input.Signature = r.privKey.Sign(tx.SignBytes(r.chainID))
	txBytes := []byte(wire.BinaryBytes(struct {
		types.Tx `json:"unwrap"`
	}{tx}))

	data, log, err := broadcastTxWithClient(r.client, txBytes)
	if err != nil {
		return err
	}
	_, _ = data, log
	return nil
}

// broadcast the transaction to tendermint
func broadcastTxWithClient(httpClient *client.HTTP, tx tmtypes.Tx) ([]byte, string, error) {
	res, err := httpClient.BroadcastTxCommit(tx)
	if err != nil {
		return nil, "", errors.Errorf("Error on broadcast tx: %v", err)
	}

	if !res.CheckTx.Code.IsOK() {
		r := res.CheckTx
		return nil, "", errors.Errorf("BroadcastTxCommit got non-zero exit code: %v. %X; %s", r.Code, r.Data, r.Log)
	}

	if !res.DeliverTx.Code.IsOK() {
		r := res.DeliverTx
		return nil, "", errors.Errorf("BroadcastTxCommit got non-zero exit code: %v. %X; %s", r.Code, r.Data, r.Log)
	}

	return res.DeliverTx.Data, res.DeliverTx.Log, nil
}
