package ibc

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/testutils"
	"github.com/tendermint/basecoin/types"
	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	eyes "github.com/tendermint/merkleeyes/client"
	tm "github.com/tendermint/tendermint/types"
)

func genGenesisDoc(chainID string, numVals int) (*tm.GenesisDoc, []types.PrivAccount) {
	var privAccs []types.PrivAccount
	genDoc := &tm.GenesisDoc{
		ChainID:    chainID,
		Validators: nil,
	}

	for i := 0; i < numVals; i++ {
		name := cmn.Fmt("%v_val_%v", chainID, i)
		privAcc := testutils.PrivAccountFromSecret(name)
		genDoc.Validators = append(genDoc.Validators, tm.GenesisValidator{
			PubKey: privAcc.Account.PubKey,
			Amount: 1,
			Name:   name,
		})
		privAccs = append(privAccs, privAcc)
	}

	return genDoc, privAccs
}

func TestIBCPlugin(t *testing.T) {

	tree := eyes.NewLocalClient("", 0)
	store := types.NewKVCache(tree)
	store.SetLogging() // Log all activity

	ibcPlugin := New()
	ctx := types.CallContext{
		CallerAddress: nil,
		CallerAccount: nil,
		Coins:         types.Coins{},
	}

	chainID_1 := "test_chain"
	genDoc_1, privAccs_1 := genGenesisDoc(chainID_1, 4)
	genDocJSON_1 := wire.JSONBytesPretty(genDoc_1)

	// Register a malformed chain
	res := ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCRegisterChainTx{
		BlockchainGenesis{
			ChainID: "test_chain",
			Genesis: "<THIS IS NOT JSON>",
		},
	}}))
	assert.Equal(t, res.Code, IBCCodeEncodingError)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Successfully register a chain
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCRegisterChainTx{
		BlockchainGenesis{
			ChainID: "test_chain",
			Genesis: string(genDocJSON_1),
		},
	}}))
	assert.True(t, res.IsOK(), res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Duplicate request fails
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCRegisterChainTx{
		BlockchainGenesis{
			ChainID: "test_chain",
			Genesis: string(genDocJSON_1),
		},
	}}))
	assert.Equal(t, res.Code, IBCCodeChainAlreadyExists, res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Create a new packet (for testing)
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketCreateTx{
		Packet{
			SrcChainID: "test_chain",
			DstChainID: "dst_chain",
			Sequence:   0,
			Type:       "data",
			Payload:    []byte("hello world"),
		},
	}}))
	assert.Equal(t, res.Code, abci.CodeType(0), res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Post a duplicate packet
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketCreateTx{
		Packet{
			SrcChainID: "test_chain",
			DstChainID: "dst_chain",
			Sequence:   0,
			Type:       "data",
			Payload:    []byte("hello world"),
		},
	}}))
	assert.Equal(t, res.Code, IBCCodePacketAlreadyExists, res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Construct a Header that includes the above packet.
	store.Sync()
	resCommit := tree.CommitSync()
	appHash := resCommit.Data
	header := tm.Header{
		ChainID: "test_chain",
		Height:  999,
		AppHash: appHash,
	}

	// Construct a Commit that signs above header
	blockHash := header.Hash()
	blockID := tm.BlockID{Hash: blockHash}
	commit := tm.Commit{
		BlockID:    blockID,
		Precommits: make([]*tm.Vote, len(privAccs_1)),
	}
	for i, privAcc := range privAccs_1 {
		vote := &tm.Vote{
			ValidatorAddress: privAcc.Account.PubKey.Address(),
			ValidatorIndex:   i,
			Height:           999,
			Round:            0,
			Type:             tm.VoteTypePrecommit,
			BlockID:          tm.BlockID{},
		}
		vote.Signature = privAcc.PrivKey.Sign(
			tm.SignBytes("test_chain", vote),
		)
		fmt.Println(">>", i, privAcc, vote)
		commit.Precommits[i] = vote
	}

	// Update a chain
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCUpdateChainTx{
		Header: header,
		Commit: commit,
	}}))
	assert.Equal(t, res.Code, abci.CodeType(0), res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()
}
