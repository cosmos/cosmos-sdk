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

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
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

type Node struct {
	PrevValset MockValidators
	Valset     MockValidators

	Cms   sdk.CommitMultiStore
	Key   sdk.StoreKey
	Store sdk.KVStore

	Commits []tmtypes.SignedHeader

	Root merkle.Root
}

func NewNode(valset MockValidators) *Node {
	key, ctx, cms, _ := defaultComponents()
	return &Node{
		Valset:  valset,
		Cms:     cms,
		Key:     key,
		Store:   ctx.KVStore(key),
		Commits: nil,
	}
}

func (node *Node) Last() tmtypes.SignedHeader {
	if len(node.Commits) == 0 {
		return tmtypes.SignedHeader{}
	}
	return node.Commits[len(node.Commits)-1]
}

func (node *Node) Commit() tmtypes.SignedHeader {
	valsethash := node.Valset.ValidatorSet().Hash()
	nextvalset := node.Valset.Mutate()
	nextvalsethash := nextvalset.ValidatorSet().Hash()
	commitid := node.Cms.Commit()

	header := tmtypes.Header{
		ChainID: chainid,
		Height:  int64(len(node.Commits) + 1),
		LastBlockID: tmtypes.BlockID{
			Hash: node.Last().Header.Hash(),
		},

		ValidatorsHash:     valsethash,
		NextValidatorsHash: nextvalsethash,
		AppHash:            commitid.Hash,
	}

	commit := node.Valset.Sign(header)

	node.PrevValset = node.Valset
	node.Valset = nextvalset
	node.Commits = append(node.Commits, commit)

	return commit
}

func (node *Node) LastStateVerifier(root merkle.Root) *Verifier {
	return NewVerifier(node.Last(), node.Valset, root)
}

type Verifier struct {
	client.ConsensusState
}

func NewVerifier(header tmtypes.SignedHeader, nextvalset MockValidators, root merkle.Root) *Verifier {
	return &Verifier{
		tendermint.ConsensusState{
			ChainID:          chainid,
			Height:           uint64(header.Height),
			Root:             root.Update(header.AppHash),
			NextValidatorSet: nextvalset.ValidatorSet(),
		},
	}
}

func (v *Verifier) Validate(header tmtypes.SignedHeader, valset, nextvalset MockValidators) error {
	newcs, err := v.ConsensusState.Validate(
		tendermint.Header{
			SignedHeader:     header,
			ValidatorSet:     valset.ValidatorSet(),
			NextValidatorSet: nextvalset.ValidatorSet(),
		},
	)
	if err != nil {
		return err
	}
	v.ConsensusState = newcs.(tendermint.ConsensusState)

	return nil
}

func (node *Node) Query(t *testing.T, root merkle.Root, k []byte) ([]byte, commitment.Proof) {
	code, value, proof := root.QueryMultiStore(node.Cms, k)
	require.Equal(t, uint32(0), code)
	return value, proof
}

func (node *Node) Set(k, value []byte) {
	node.Store.Set(newRoot().Key(k), value)
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
			v, p := node.Query(t, root.(merkle.Root), []byte(kvp.Key))
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
