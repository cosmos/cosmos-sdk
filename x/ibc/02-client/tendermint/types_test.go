package tendermint

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	stypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const chainid = "testchain"

func defaultComponents() (sdk.StoreKey, sdk.Context, stypes.CommitMultiStore, *codec.Codec) {
	key := sdk.NewKVStoreKey("ibc")
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
	valset MockValidators

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
	nextvalset := node.valset.Mutate(false)
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

	node.commits = append(node.commits, commit)

	return commit
}

func (node *node) Set(key, value string) {
	node.store.Set(append([]byte{0x00}, []byte(key)...), []byte(value))
}

type Verifier struct {
	ConsensusState
}

func NewVerifier(header tmtypes.SignedHeader, nextvalset MockValidators) *Verifier {
	return &Verifier{
		ConsensusState{
			ChainID:          chainid,
			Height:           uint64(header.Height),
			Root:             header.AppHash,
			NextValidatorSet: nextvalset.ValidatorSet(),
		},
	}
}

func (v *Verifier) Validate(header tmtypes.SignedHeader, nextvalset MockValidators) error {
	newcs, err := v.ConsensusState.Validate(
		Header{
			SignedHeader:     header,
			NextValidatorSet: nextvalset.ValidatorSet(),
		},
	)
	if err != nil {
		return err
	}
	v.ConsensusState = newcs.(ConsensusState)

	return nil
}

func TestUpdate(t *testing.T) {
	node := NewNode(NewMockValidators(100, 10))

	node.Commit()

	verifier := NewVerifier(node.last(), node.valset)

	header := node.Commit()

	err := verifier.Validate(header, node.valset)
	require.NoError(t, err)
}
