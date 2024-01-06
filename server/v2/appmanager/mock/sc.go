package mock

import (
	"crypto/sha256"

	storev2 "cosmossdk.io/store/v2"
	"github.com/celestiaorg/smt"
	ics23 "github.com/cosmos/ics23/go"
)

var _ storev2.Committer = (*StateCommitment)(nil)

func NewStateCommitment() StateCommitment {
	ns, vs := smt.NewSimpleMap(), smt.NewSimpleMap()
	tree := smt.NewSparseMerkleTree(ns, vs, sha256.New())
	return StateCommitment{tree: tree}
}

type StateCommitment struct {
	tree *smt.SparseMerkleTree
}

func (s StateCommitment) WriteBatch(_cs *storev2.Changeset) error {
	cs, ok := _cs.Pairs[""]
	if !ok {
		panic("mis implementation, we only expect one store key")
	}
	for _, change := range cs {
		if change.Value == nil {
			_, err := s.tree.Delete(change.Value)
			if err != nil {
				return err
			}
		} else {
			_, err := s.tree.Update(change.Key, change.Value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s StateCommitment) WorkingStoreInfos(version uint64) []storev2.StoreInfo {
	// TODO implement me
	panic("implement me")
}

func (s StateCommitment) GetLatestVersion() (uint64, error) {
	// TODO implement me
	panic("implement me")
}

func (s StateCommitment) LoadVersion(targetVersion uint64) error {
	// TODO implement me
	panic("implement me")
}

func (s StateCommitment) Commit() ([]storev2.StoreInfo, error) {
	return []storev2.StoreInfo{}, nil
}

func (s StateCommitment) SetInitialVersion(version uint64) error {
	// TODO implement me
	panic("implement me")
}

func (s StateCommitment) GetProof(storeKey string, version uint64, key []byte) (*ics23.CommitmentProof, error) {
	// TODO implement me
	panic("implement me")
}

func (s StateCommitment) Prune(version uint64) error {
	// TODO implement me
	panic("implement me")
}

func (s StateCommitment) Close() error {
	// TODO implement me
	panic("implement me")
}
