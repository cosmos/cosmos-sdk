package tendermint

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	stypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

const chainid = "testchain"

func defaultComponents() (sdk.StoreKey, sdk.Context, stypes.CommitMultiStore, *codec.Codec) {
	key := sdk.NewKVStoreKey("test")
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	cdc := codec.New()
	return key, ctx, cms, cdc
}

type node struct {
	prevvalset MockValidators
	valset     MockValidators

	cms   sdk.CommitMultiStore
	store sdk.KVStore

	commits []tmtypes.SignedHeader
}

func NewNode(valset MockValidators) *node {
	key, ctx, cms, _ := defaultComponents()
	return &node{
		valset:  valset,
		cms:     cms,
		store:   ctx.KVStore(key),
		commits: nil,
	}
}

func (node *node) last() tmtypes.SignedHeader {
	if len(node.commits) == 0 {
		return tmtypes.SignedHeader{}
	}
	return node.commits[len(node.commits)-1]
}

func (node *node) Commit() tmtypes.SignedHeader {
	valsethash := node.valset.ValidatorSet().Hash()
	nextvalset := node.valset.Mutate()
	nextvalsethash := nextvalset.ValidatorSet().Hash()
	commitid := node.cms.Commit()

	header := tmtypes.Header{
		ChainID: chainid,
		Height:  int64(len(node.commits) + 1),
		LastBlockID: tmtypes.BlockID{
			Hash: node.last().Header.Hash(),
		},

		ValidatorsHash:     valsethash,
		NextValidatorsHash: nextvalsethash,
		AppHash:            commitid.Hash,
	}

	commit := node.valset.Sign(header)

	node.prevvalset = node.valset
	node.valset = nextvalset
	node.commits = append(node.commits, commit)

	return commit
}

type Verifier struct {
	ConsensusState
}

func NewVerifier(header tmtypes.SignedHeader, nextvalset MockValidators, root merkle.Root) *Verifier {
	return &Verifier{
		ConsensusState{
			ChainID:          chainid,
			Height:           uint64(header.Height),
			Root:             root.Update(header.AppHash),
			NextValidatorSet: nextvalset.ValidatorSet(),
		},
	}
}

func (v *Verifier) Validate(header tmtypes.SignedHeader, valset, nextvalset MockValidators) error {
	newcs, err := v.ConsensusState.Validate(
		Header{
			SignedHeader:     header,
			ValidatorSet:     valset.ValidatorSet(),
			NextValidatorSet: nextvalset.ValidatorSet(),
		},
	)
	if err != nil {
		return err
	}
	v.ConsensusState = newcs.(ConsensusState)

	return nil
}

func newRoot() merkle.Root {
	return merkle.NewRoot(nil, [][]byte{[]byte("test")}, []byte{0x12, 0x34})
}

func testUpdate(t *testing.T, interval int, ok bool) {
	node := NewNode(NewMockValidators(100, 10))

	node.Commit()

	root := newRoot()

	verifier := NewVerifier(node.last(), node.valset, root)

	for i := 0; i < 100; i++ {
		header := node.Commit()

		if i%interval == 0 {
			err := verifier.Validate(header, node.prevvalset, node.valset)
			if ok {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		}
	}
}

func TestEveryBlockUpdate(t *testing.T) {
	testUpdate(t, 1, true)
}

func TestEvenBlockUpdate(t *testing.T) {
	testUpdate(t, 2, true)
}

func TestSixthBlockUpdate(t *testing.T) {
	testUpdate(t, 6, true)
}

/*
// This should fail, since the amount of mutation is so large
// Commented out because it sometimes success
func TestTenthBlockUpdate(t *testing.T) {
	testUpdate(t, 10, false)
}
*/

func (node *node) query(t *testing.T, root merkle.Root, k []byte) ([]byte, commitment.Proof) {
	code, value, proof := root.Query(node.cms, k)
	require.Equal(t, uint32(0), code)
	return value, proof
}

func (node *node) Set(k, value []byte) {
	node.store.Set(newRoot().Key(k), value)
}

func testProof(t *testing.T) {
	node := NewNode(NewMockValidators(100, 10))

	node.Commit()

	kvps := cmn.KVPairs{}
	for h := 0; h < 20; h++ {
		for i := 0; i < 100; i++ {
			k := make([]byte, 32)
			v := make([]byte, 32)
			rand.Read(k)
			rand.Read(v)
			kvps = append(kvps, cmn.KVPair{Key: k, Value: v})
			node.Set(k, v)
		}
		header := node.Commit()
		proofs := []commitment.Proof{}
		root := newRoot().Update(header.AppHash)
		for _, kvp := range kvps {
			v, p := node.query(t, root.(merkle.Root), []byte(kvp.Key))
			require.Equal(t, kvp.Value, v)
			proofs = append(proofs, p)
		}
		cstore, err := commitment.NewStore(root, proofs)
		require.NoError(t, err)

		for _, kvp := range kvps {
			require.True(t, cstore.Prove(kvp.Key, kvp.Value))
		}
	}
}

func TestProofs(t *testing.T) {
	testProof(t)
}
