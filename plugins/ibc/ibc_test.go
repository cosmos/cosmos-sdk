package ibc

// import (
// 	"bytes"
// 	"encoding/json"
// 	"sort"
// 	"strings"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	abci "github.com/tendermint/abci/types"
// 	crypto "github.com/tendermint/go-crypto"
// 	"github.com/tendermint/go-wire"
// 	eyes "github.com/tendermint/merkleeyes/client"
// 	"github.com/tendermint/merkleeyes/iavl"
// 	cmn "github.com/tendermint/tmlibs/common"

// 	"github.com/tendermint/basecoin/types"
// 	tm "github.com/tendermint/tendermint/types"
// )

// // NOTE: PrivAccounts are sorted by Address,
// // GenesisDoc, not necessarily.
// func genGenesisDoc(chainID string, numVals int) (*tm.GenesisDoc, []types.PrivAccount) {
// 	var privAccs []types.PrivAccount
// 	genDoc := &tm.GenesisDoc{
// 		ChainID:    chainID,
// 		Validators: nil,
// 	}

// 	for i := 0; i < numVals; i++ {
// 		name := cmn.Fmt("%v_val_%v", chainID, i)
// 		privAcc := types.PrivAccountFromSecret(name)
// 		genDoc.Validators = append(genDoc.Validators, tm.GenesisValidator{
// 			PubKey: privAcc.PubKey,
// 			Amount: 1,
// 			Name:   name,
// 		})
// 		privAccs = append(privAccs, privAcc)
// 	}

// 	// Sort PrivAccounts
// 	sort.Sort(PrivAccountsByAddress(privAccs))

// 	return genDoc, privAccs
// }

// //-------------------------------------
// // Implements sort for sorting PrivAccount by address.

// type PrivAccountsByAddress []types.PrivAccount

// func (pas PrivAccountsByAddress) Len() int {
// 	return len(pas)
// }

// func (pas PrivAccountsByAddress) Less(i, j int) bool {
// 	return bytes.Compare(pas[i].Account.PubKey.Address(), pas[j].Account.PubKey.Address()) == -1
// }

// func (pas PrivAccountsByAddress) Swap(i, j int) {
// 	it := pas[i]
// 	pas[i] = pas[j]
// 	pas[j] = it
// }

// //--------------------------------------------------------------------------------

// var testGenesisDoc = `{
//   "app_hash": "",
//   "chain_id": "test_chain_1",
//   "genesis_time": "0001-01-01T00:00:00.000Z",
//   "validators": [
//     {
//       "amount": 10,
//       "name": "",
//       "pub_key": {
//           "type": "ed25519",
//           "data":"D6EBB92440CF375054AA59BCF0C99D596DEEDFFB2543CAE1BA1908B72CF9676A"
//       }
//     }
//   ],
//   "app_options": {
//     "accounts": [
//       {
//         "pub_key": {
//           "type": "ed25519",
//           "data": "B3588BDC92015ED3CDB6F57A86379E8C79A7111063610B7E625487C76496F4DF"
//         },
//         "coins": [
//           {
//             "denom": "mycoin",
//             "amount": 9007199254740992
//           }
//         ]
//       }
//     ]
//   }
// }`

// func TestIBCGenesisFromString(t *testing.T) {
// 	eyesClient := eyes.NewLocalClient("", 0)
// 	store := types.NewKVCache(eyesClient)
// 	store.SetLogging() // Log all activity

// 	ibcPlugin := New()
// 	ctx := types.NewCallContext(nil, nil, types.Coins{})

// 	registerChain(t, ibcPlugin, store, ctx, "test_chain", testGenesisDoc)
// }

// //--------------------------------------------------------------------------------

// func TestIBCPluginRegister(t *testing.T) {
// 	require := require.New(t)

// 	eyesClient := eyes.NewLocalClient("", 0)
// 	store := types.NewKVCache(eyesClient)
// 	store.SetLogging() // Log all activity

// 	ibcPlugin := New()
// 	ctx := types.NewCallContext(nil, nil, types.Coins{})

// 	chainID_1 := "test_chain"
// 	genDoc_1, _ := genGenesisDoc(chainID_1, 4)
// 	genDocJSON_1, err := json.Marshal(genDoc_1)
// 	require.Nil(err)

// 	// Register a malformed chain
// 	res := ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCRegisterChainTx{
// 		BlockchainGenesis{
// 			ChainID: "test_chain",
// 			Genesis: "<THIS IS NOT JSON>",
// 		},
// 	}}))
// 	assertAndLog(t, store, res, IBCCodeEncodingError)

// 	// Successfully register a chain
// 	registerChain(t, ibcPlugin, store, ctx, "test_chain", string(genDocJSON_1))

// 	// Duplicate request fails
// 	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCRegisterChainTx{
// 		BlockchainGenesis{
// 			ChainID: "test_chain",
// 			Genesis: string(genDocJSON_1),
// 		},
// 	}}))
// 	assertAndLog(t, store, res, IBCCodeChainAlreadyExists)
// }

// func TestIBCPluginPost(t *testing.T) {
// 	require := require.New(t)

// 	eyesClient := eyes.NewLocalClient("", 0)
// 	store := types.NewKVCache(eyesClient)
// 	store.SetLogging() // Log all activity

// 	ibcPlugin := New()
// 	ctx := types.NewCallContext(nil, nil, types.Coins{})

// 	chainID_1 := "test_chain"
// 	genDoc_1, _ := genGenesisDoc(chainID_1, 4)
// 	genDocJSON_1, err := json.Marshal(genDoc_1)
// 	require.Nil(err)

// 	// Register a chain
// 	registerChain(t, ibcPlugin, store, ctx, "test_chain", string(genDocJSON_1))

// 	// Create a new packet (for testing)
// 	packet := NewPacket("test_chain", "dst_chain", 0, DataPayload([]byte("hello world")))
// 	res := ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketCreateTx{
// 		Packet: packet,
// 	}}))
// 	assertAndLog(t, store, res, abci.CodeType_OK)

// 	// Post a duplicate packet
// 	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketCreateTx{
// 		Packet: packet,
// 	}}))
// 	assertAndLog(t, store, res, IBCCodePacketAlreadyExists)
// }

// func TestIBCPluginPayloadBytes(t *testing.T) {
// 	assert := assert.New(t)
// 	require := require.New(t)

// 	eyesClient := eyes.NewLocalClient("", 0)
// 	store := types.NewKVCache(eyesClient)
// 	store.SetLogging() // Log all activity

// 	ibcPlugin := New()
// 	ctx := types.NewCallContext(nil, nil, types.Coins{})

// 	chainID_1 := "test_chain"
// 	genDoc_1, privAccs_1 := genGenesisDoc(chainID_1, 4)
// 	genDocJSON_1, err := json.Marshal(genDoc_1)
// 	require.Nil(err)

// 	// Register a chain
// 	registerChain(t, ibcPlugin, store, ctx, "test_chain", string(genDocJSON_1))

// 	// Create a new packet (for testing)
// 	packet := NewPacket("test_chain", "dst_chain", 0, DataPayload([]byte("hello world")))
// 	res := ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketCreateTx{
// 		Packet: packet,
// 	}}))
// 	assertAndLog(t, store, res, abci.CodeType_OK)

// 	// Construct a Header that includes the above packet.
// 	store.Sync()
// 	resCommit := eyesClient.CommitSync()
// 	appHash := resCommit.Data
// 	header := newHeader("test_chain", 999, appHash, []byte("must_exist"))

// 	// Construct a Commit that signs above header
// 	commit := constructCommit(privAccs_1, header)

// 	// Update a chain
// 	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCUpdateChainTx{
// 		Header: header,
// 		Commit: commit,
// 	}}))
// 	assertAndLog(t, store, res, abci.CodeType_OK)

// 	// Get proof for the packet
// 	packetKey := toKey(_IBC, _EGRESS,
// 		packet.SrcChainID,
// 		packet.DstChainID,
// 		cmn.Fmt("%v", packet.Sequence),
// 	)
// 	resQuery, err := eyesClient.QuerySync(abci.RequestQuery{
// 		Path:  "/store",
// 		Data:  packetKey,
// 		Prove: true,
// 	})
// 	assert.Nil(err)
// 	var proof *iavl.IAVLProof
// 	err = wire.ReadBinaryBytes(resQuery.Proof, &proof)
// 	assert.Nil(err)

// 	// Post a packet
// 	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketPostTx{
// 		FromChainID:     "test_chain",
// 		FromChainHeight: 999,
// 		Packet:          packet,
// 		Proof:           proof,
// 	}}))
// 	assertAndLog(t, store, res, abci.CodeType_OK)
// }

// func TestIBCPluginPayloadCoins(t *testing.T) {
// 	assert := assert.New(t)
// 	require := require.New(t)

// 	eyesClient := eyes.NewLocalClient("", 0)
// 	store := types.NewKVCache(eyesClient)
// 	store.SetLogging() // Log all activity

// 	ibcPlugin := New()
// 	coins := types.Coins{
// 		types.Coin{
// 			Denom:  "mycoin",
// 			Amount: 100,
// 		},
// 	}
// 	ctx := types.NewCallContext(nil, nil, coins)

// 	chainID_1 := "test_chain"
// 	genDoc_1, privAccs_1 := genGenesisDoc(chainID_1, 4)
// 	genDocJSON_1, err := json.Marshal(genDoc_1)
// 	require.Nil(err)

// 	// Register a chain
// 	registerChain(t, ibcPlugin, store, ctx, "test_chain", string(genDocJSON_1))

// 	// send coins to this addr on the other chain
// 	destinationAddr := []byte("some address")
// 	coinsBad := types.Coins{types.Coin{"mycoin", 200}}
// 	coinsGood := types.Coins{types.Coin{"mycoin", 1}}

// 	// Try to send too many coins
// 	packet := NewPacket("test_chain", "dst_chain", 0, CoinsPayload{
// 		Address: destinationAddr,
// 		Coins:   coinsBad,
// 	})
// 	res := ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketCreateTx{
// 		Packet: packet,
// 	}}))
// 	assertAndLog(t, store, res, abci.CodeType_InsufficientFunds)

// 	// Send a small enough number of coins
// 	packet = NewPacket("test_chain", "dst_chain", 0, CoinsPayload{
// 		Address: destinationAddr,
// 		Coins:   coinsGood,
// 	})
// 	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketCreateTx{
// 		Packet: packet,
// 	}}))
// 	assertAndLog(t, store, res, abci.CodeType_OK)

// 	// Construct a Header that includes the above packet.
// 	store.Sync()
// 	resCommit := eyesClient.CommitSync()
// 	appHash := resCommit.Data
// 	header := newHeader("test_chain", 999, appHash, []byte("must_exist"))

// 	// Construct a Commit that signs above header
// 	commit := constructCommit(privAccs_1, header)

// 	// Update a chain
// 	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCUpdateChainTx{
// 		Header: header,
// 		Commit: commit,
// 	}}))
// 	assertAndLog(t, store, res, abci.CodeType_OK)

// 	// Get proof for the packet
// 	packetKey := toKey(_IBC, _EGRESS,
// 		packet.SrcChainID,
// 		packet.DstChainID,
// 		cmn.Fmt("%v", packet.Sequence),
// 	)
// 	resQuery, err := eyesClient.QuerySync(abci.RequestQuery{
// 		Path:  "/store",
// 		Data:  packetKey,
// 		Prove: true,
// 	})
// 	assert.Nil(err)
// 	var proof *iavl.IAVLProof
// 	err = wire.ReadBinaryBytes(resQuery.Proof, &proof)
// 	assert.Nil(err)

// 	// Account should be empty before the tx
// 	acc := types.GetAccount(store, destinationAddr)
// 	assert.Nil(acc)

// 	// Post a packet
// 	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketPostTx{
// 		FromChainID:     "test_chain",
// 		FromChainHeight: 999,
// 		Packet:          packet,
// 		Proof:           proof,
// 	}}))
// 	assertAndLog(t, store, res, abci.CodeType_OK)

// 	// Account should now have some coins
// 	acc = types.GetAccount(store, destinationAddr)
// 	assert.Equal(acc.Balance, coinsGood)
// }

// func TestIBCPluginBadCommit(t *testing.T) {
// 	require := require.New(t)

// 	eyesClient := eyes.NewLocalClient("", 0)
// 	store := types.NewKVCache(eyesClient)
// 	store.SetLogging() // Log all activity

// 	ibcPlugin := New()
// 	ctx := types.NewCallContext(nil, nil, types.Coins{})

// 	chainID_1 := "test_chain"
// 	genDoc_1, privAccs_1 := genGenesisDoc(chainID_1, 4)
// 	genDocJSON_1, err := json.Marshal(genDoc_1)
// 	require.Nil(err)

// 	// Successfully register a chain
// 	registerChain(t, ibcPlugin, store, ctx, "test_chain", string(genDocJSON_1))

// 	// Construct a Header
// 	header := newHeader("test_chain", 999, nil, []byte("must_exist"))

// 	// Construct a Commit that signs above header
// 	commit := constructCommit(privAccs_1, header)

// 	// Update a chain with a broken commit
// 	// Modify the first byte of the first signature
// 	sig := commit.Precommits[0].Signature.Unwrap().(crypto.SignatureEd25519)
// 	sig[0] += 1
// 	commit.Precommits[0].Signature = sig.Wrap()
// 	res := ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCUpdateChainTx{
// 		Header: header,
// 		Commit: commit,
// 	}}))
// 	assertAndLog(t, store, res, IBCCodeInvalidCommit)

// }

// func TestIBCPluginBadProof(t *testing.T) {
// 	assert := assert.New(t)
// 	require := require.New(t)

// 	eyesClient := eyes.NewLocalClient("", 0)
// 	store := types.NewKVCache(eyesClient)
// 	store.SetLogging() // Log all activity

// 	ibcPlugin := New()
// 	ctx := types.NewCallContext(nil, nil, types.Coins{})

// 	chainID_1 := "test_chain"
// 	genDoc_1, privAccs_1 := genGenesisDoc(chainID_1, 4)
// 	genDocJSON_1, err := json.Marshal(genDoc_1)
// 	require.Nil(err)

// 	// Successfully register a chain
// 	registerChain(t, ibcPlugin, store, ctx, "test_chain", string(genDocJSON_1))

// 	// Create a new packet (for testing)
// 	packet := NewPacket("test_chain", "dst_chain", 0, DataPayload([]byte("hello world")))
// 	res := ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketCreateTx{
// 		Packet: packet,
// 	}}))
// 	assertAndLog(t, store, res, abci.CodeType_OK)

// 	// Construct a Header that includes the above packet.
// 	store.Sync()
// 	resCommit := eyesClient.CommitSync()
// 	appHash := resCommit.Data
// 	header := newHeader("test_chain", 999, appHash, []byte("must_exist"))

// 	// Construct a Commit that signs above header
// 	commit := constructCommit(privAccs_1, header)

// 	// Update a chain
// 	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCUpdateChainTx{
// 		Header: header,
// 		Commit: commit,
// 	}}))
// 	assertAndLog(t, store, res, abci.CodeType_OK)

// 	// Get proof for the packet
// 	packetKey := toKey(_IBC, _EGRESS,
// 		packet.SrcChainID,
// 		packet.DstChainID,
// 		cmn.Fmt("%v", packet.Sequence),
// 	)
// 	resQuery, err := eyesClient.QuerySync(abci.RequestQuery{
// 		Path:  "/store",
// 		Data:  packetKey,
// 		Prove: true,
// 	})
// 	assert.Nil(err)
// 	var proof *iavl.IAVLProof
// 	err = wire.ReadBinaryBytes(resQuery.Proof, &proof)
// 	assert.Nil(err)

// 	// Mutate the proof
// 	proof.InnerNodes[0].Height += 1

// 	// Post a packet
// 	res = ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCPacketPostTx{
// 		FromChainID:     "test_chain",
// 		FromChainHeight: 999,
// 		Packet:          packet,
// 		Proof:           proof,
// 	}}))
// 	assertAndLog(t, store, res, IBCCodeInvalidProof)
// }

// //-------------------------------------
// // utils

// func assertAndLog(t *testing.T, store *types.KVCache, res abci.Result, codeExpected abci.CodeType) {
// 	assert := assert.New(t)
// 	assert.Equal(codeExpected, res.Code, res.Log)
// 	t.Log(">>", strings.Join(store.GetLogLines(), "\n"))
// 	store.ClearLogLines()
// }

// func newHeader(chainID string, height int, appHash, valHash []byte) tm.Header {
// 	return tm.Header{
// 		ChainID:        chainID,
// 		Height:         height,
// 		AppHash:        appHash,
// 		ValidatorsHash: valHash,
// 	}
// }

// func registerChain(t *testing.T, ibcPlugin *IBCPlugin, store *types.KVCache, ctx types.CallContext, chainID, genDoc string) {
// 	res := ibcPlugin.RunTx(store, ctx, wire.BinaryBytes(struct{ IBCTx }{IBCRegisterChainTx{
// 		BlockchainGenesis{
// 			ChainID: chainID,
// 			Genesis: genDoc,
// 		},
// 	}}))
// 	assertAndLog(t, store, res, abci.CodeType_OK)
// }

// func constructCommit(privAccs []types.PrivAccount, header tm.Header) tm.Commit {
// 	blockHash := header.Hash()
// 	blockID := tm.BlockID{Hash: blockHash}
// 	commit := tm.Commit{
// 		BlockID:    blockID,
// 		Precommits: make([]*tm.Vote, len(privAccs)),
// 	}
// 	for i, privAcc := range privAccs {
// 		vote := &tm.Vote{
// 			ValidatorAddress: privAcc.Account.PubKey.Address(),
// 			ValidatorIndex:   i,
// 			Height:           999,
// 			Round:            0,
// 			Type:             tm.VoteTypePrecommit,
// 			BlockID:          tm.BlockID{Hash: blockHash},
// 		}
// 		vote.Signature = privAcc.PrivKey.Sign(
// 			tm.SignBytes("test_chain", vote),
// 		)
// 		commit.Precommits[i] = vote
// 	}
// 	return commit
// }
