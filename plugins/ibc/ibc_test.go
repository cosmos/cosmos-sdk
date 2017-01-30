package ibc

import (
	"bytes"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/testutils"
	"github.com/tendermint/basecoin/types"
	cmn "github.com/tendermint/go-common"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-merkle"
	"github.com/tendermint/go-wire"
	eyes "github.com/tendermint/merkleeyes/client"
	tm "github.com/tendermint/tendermint/types"
)

// NOTE: PrivAccounts are sorted by Address,
// GenesisDoc, not necessarily.
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

	// Sort PrivAccounts
	sort.Sort(PrivAccountsByAddress(privAccs))

	return genDoc, privAccs
}

//-------------------------------------
// Implements sort for sorting PrivAccount by address.

type PrivAccountsByAddress []types.PrivAccount

func (pas PrivAccountsByAddress) Len() int {
	return len(pas)
}

func (pas PrivAccountsByAddress) Less(i, j int) bool {
	return bytes.Compare(pas[i].Account.PubKey.Address(), pas[j].Account.PubKey.Address()) == -1
}

func (pas PrivAccountsByAddress) Swap(i, j int) {
	it := pas[i]
	pas[i] = pas[j]
	pas[j] = it
}

//--------------------------------------------------------------------------------

func TestIBCPlugin(t *testing.T) {

	eyesClient := eyes.NewLocalClient("", 0)
	store := types.NewKVCache(eyesClient)
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
	assert.Equal(t, IBCCodeEncodingError, res.Code)
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
	assert.Equal(t, IBCCodeChainAlreadyExists, res.Code, res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Create a new packet (for testing)
	packet := Packet{
		SrcChainID: "test_chain",
		DstChainID: "dst_chain",
		Sequence:   0,
		Type:       "data",
		Payload:    []byte("hello world"),
	}
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketCreateTx{
		Packet: packet,
	}}))
	assert.Equal(t, abci.CodeType_OK, res.Code, res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Post a duplicate packet
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketCreateTx{
		Packet: packet,
	}}))
	assert.Equal(t, IBCCodePacketAlreadyExists, res.Code, res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Construct a Header that includes the above packet.
	store.Sync()
	resCommit := eyesClient.CommitSync()
	appHash := resCommit.Data
	header := tm.Header{
		ChainID:        "test_chain",
		Height:         999,
		AppHash:        appHash,
		ValidatorsHash: []byte("must_exist"), // TODO make optional
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
			BlockID:          tm.BlockID{Hash: blockHash},
		}
		vote.Signature = privAcc.PrivKey.Sign(
			tm.SignBytes("test_chain", vote),
		)
		commit.Precommits[i] = vote
	}

	// Update a chain
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCUpdateChainTx{
		Header: header,
		Commit: commit,
	}}))
	assert.Equal(t, abci.CodeType_OK, res.Code, res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Get proof for the packet
	packetKey := toKey(_IBC, _EGRESS,
		packet.SrcChainID,
		packet.DstChainID,
		cmn.Fmt("%v", packet.Sequence),
	)
	resQuery, err := eyesClient.QuerySync(abci.RequestQuery{
		Path:  "/store",
		Data:  packetKey,
		Prove: true,
	})
	assert.Nil(t, err)
	var proof *merkle.IAVLProof
	err = wire.ReadBinaryBytes(resQuery.Proof, &proof)
	assert.Nil(t, err)

	// Post a packet
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketPostTx{
		FromChainID:     "test_chain",
		FromChainHeight: 999,
		Packet:          packet,
		Proof:           proof,
	}}))
	assert.Equal(t, abci.CodeType_OK, res.Code, res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()
}

func TestIBCPluginBadCommit(t *testing.T) {

	eyesClient := eyes.NewLocalClient("", 0)
	store := types.NewKVCache(eyesClient)
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

	// Successfully register a chain
	res := ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCRegisterChainTx{
		BlockchainGenesis{
			ChainID: "test_chain",
			Genesis: string(genDocJSON_1),
		},
	}}))
	assert.True(t, res.IsOK(), res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Construct a Header
	header := tm.Header{
		ChainID:        "test_chain",
		Height:         999,
		ValidatorsHash: []byte("must_exist"), // TODO make optional
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
			BlockID:          tm.BlockID{Hash: blockHash},
		}
		vote.Signature = privAcc.PrivKey.Sign(
			tm.SignBytes("test_chain", vote),
		)
		commit.Precommits[i] = vote
	}

	// Update a chain with a broken commit
	// Modify the first byte of the first signature
	sig := commit.Precommits[0].Signature.(crypto.SignatureEd25519)
	sig[0] += 1
	commit.Precommits[0].Signature = sig
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCUpdateChainTx{
		Header: header,
		Commit: commit,
	}}))
	assert.Equal(t, IBCCodeInvalidCommit, res.Code, res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

}

func TestIBCPluginBadProof(t *testing.T) {

	eyesClient := eyes.NewLocalClient("", 0)
	store := types.NewKVCache(eyesClient)
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

	// Successfully register a chain
	res := ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCRegisterChainTx{
		BlockchainGenesis{
			ChainID: "test_chain",
			Genesis: string(genDocJSON_1),
		},
	}}))
	assert.True(t, res.IsOK(), res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Create a new packet (for testing)
	packet := Packet{
		SrcChainID: "test_chain",
		DstChainID: "dst_chain",
		Sequence:   0,
		Type:       "data",
		Payload:    []byte("hello world"),
	}
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketCreateTx{
		Packet: packet,
	}}))
	assert.Equal(t, abci.CodeType_OK, res.Code, res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Construct a Header that includes the above packet.
	store.Sync()
	resCommit := eyesClient.CommitSync()
	appHash := resCommit.Data
	header := tm.Header{
		ChainID:        "test_chain",
		Height:         999,
		AppHash:        appHash,
		ValidatorsHash: []byte("must_exist"), // TODO make optional
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
			BlockID:          tm.BlockID{Hash: blockHash},
		}
		vote.Signature = privAcc.PrivKey.Sign(
			tm.SignBytes("test_chain", vote),
		)
		commit.Precommits[i] = vote
	}

	// Update a chain
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCUpdateChainTx{
		Header: header,
		Commit: commit,
	}}))
	assert.Equal(t, abci.CodeType_OK, res.Code, res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()

	// Get proof for the packet
	packetKey := toKey(_IBC, _EGRESS,
		packet.SrcChainID,
		packet.DstChainID,
		cmn.Fmt("%v", packet.Sequence),
	)
	resQuery, err := eyesClient.QuerySync(abci.RequestQuery{
		Path:  "/store",
		Data:  packetKey,
		Prove: true,
	})
	assert.Nil(t, err)
	var proof *merkle.IAVLProof
	err = wire.ReadBinaryBytes(resQuery.Proof, &proof)
	assert.Nil(t, err)

	// Mutate the proof
	proof.InnerNodes[0].Height += 1

	// Post a packet
	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketPostTx{
		FromChainID:     "test_chain",
		FromChainHeight: 999,
		Packet:          packet,
		Proof:           proof,
	}}))
	assert.Equal(t, IBCCodeInvalidProof, res.Code, res.Log)
	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
	store.ClearLogLines()
}
