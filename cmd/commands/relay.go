package commands

import (
	"fmt"
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
	Short: "Start basecoin relayer to relay IBC packets between chains",
	RunE:  relayCmd,
}

//flags
var (
	chain1AddrFlag string
	chain2AddrFlag string

	chain1IDFlag string
	chain2IDFlag string

	fromFileFlag string
)

func init() {

	flags := []Flag2Register{
		{&chain1AddrFlag, "chain1-addr", "tcp://localhost:46657", "Node address for chain1"},
		{&chain2AddrFlag, "chain2-addr", "tcp://localhost:36657", "Node address for chain2"},
		{&chain1IDFlag, "chain1-id", "test_chain_1", "ChainID for chain1"},
		{&chain2IDFlag, "chain2-id", "test_chain_2", "ChainID for chain2"},
		{&fromFileFlag, "from", "key.json", "Path to a private key to sign the transaction"},
	}
	RegisterFlags(RelayCmd, flags)
}

func loop(addr1, addr2, id1, id2 string) {
	latestSeq := -1

	// load the priv key
	privKey, err := LoadKey(fromFlag)
	if err != nil {
		logger.Error(err.Error())
		cmn.PanicCrisis(err.Error())
	}

	// relay from chain1 to chain2
	thisRelayer := relayer{privKey, id2, addr2}

OUTER:
	for {

		time.Sleep(time.Second)

		// get the latest ibc packet sequence number
		key := fmt.Sprintf("ibc,egress,%v,%v", id1, id2)
		query, err := Query(addr1, []byte(key))
		if err != nil {
			logger.Error(err.Error())
			continue OUTER
		}
		seq, err := strconv.ParseUint(string(query.Value), 10, 64)
		if err != nil {
			logger.Error(err.Error())
			continue OUTER
		}

		// if there's a new packet, relay the header and commit data
		if latestSeq < int(seq) {
			header, commit, err := getHeaderAndCommit(addr1, int(query.Height))
			if err != nil {
				logger.Error(err.Error())
				continue OUTER
			}

			// update the chain state on the other chain
			ibcTx := ibc.IBCUpdateChainTx{
				Header: *header,
				Commit: *commit,
			}
			if err := thisRelayer.appTx(ibcTx); err != nil {
				logger.Error(err.Error())
				continue OUTER
			}
		}

		// get all packets since the last one we relayed
		for ; latestSeq < int(seq); latestSeq++ {
			key := fmt.Sprintf("ibc,egress,%v,%v,%d", id1, id2, latestSeq)
			query, err := Query(addr1, []byte(key))
			if err != nil {
				logger.Error(err.Error())
				continue OUTER
			}

			var packet ibc.Packet
			err = wire.ReadBinaryBytes(query.Value, &packet)
			if err != nil {
				logger.Error(err.Error())
				continue OUTER
			}

			proof := new(iavl.IAVLProof)
			err = wire.ReadBinaryBytes(query.Proof, proof)
			if err != nil {
				logger.Error(err.Error())
				continue OUTER
			}

			// relay the packet and proof
			ibcTx := ibc.IBCPacketPostTx{
				FromChainID:     id1,
				FromChainHeight: uint64(query.Height),
				Packet:          packet,
				Proof:           proof,
			}
			if err := thisRelayer.appTx(ibcTx); err != nil {
				logger.Error(err.Error())
				continue OUTER
			}
		}
	}
}

type relayer struct {
	privKey  *Key
	chainID  string
	nodeAddr string
}

func (r *relayer) appTx(ibcTx ibc.IBCTx) error {
	sequence, err := getSeq(r.privKey.Address[:])
	if err != nil {
		return err
	}

	data := []byte(wire.BinaryBytes(struct {
		ibc.IBCTx `json:"unwrap"`
	}{ibcTx}))

	input := types.NewTxInput(r.privKey.PubKey, types.Coins{}, sequence)
	tx := &types.AppTx{
		Gas:   0,
		Fee:   types.Coin{"mycoin", 1},
		Name:  "IBC",
		Input: input,
		Data:  data,
	}

	tx.Input.Signature = r.privKey.Sign(tx.SignBytes(r.chainID))
	txBytes := []byte(wire.BinaryBytes(struct {
		types.Tx `json:"unwrap"`
	}{tx}))

	data, log, err := broadcastTxToNode(r.nodeAddr, txBytes)
	if err != nil {
		return err
	}
	fmt.Printf("Response: %X ; %s\n", data, log)
	return nil
}

// broadcast the transaction to tendermint
func broadcastTxToNode(nodeAddr string, tx tmtypes.Tx) ([]byte, string, error) {
	httpClient := client.NewHTTP(nodeAddr, "/websocket")
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

func relayCmd(cmd *cobra.Command, args []string) error {

	go loop(chain1AddrFlag, chain2AddrFlag, chain1IDFlag, chain2IDFlag)
	go loop(chain2AddrFlag, chain1AddrFlag, chain2IDFlag, chain1IDFlag)

	cmn.TrapSignal(func() {
		// TODO: Cleanup
	})
	return nil

}
