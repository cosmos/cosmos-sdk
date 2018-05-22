package merkle

import (
	"bytes"

	cmn "github.com/tendermint/tmlibs/common"
	. "github.com/tendermint/tmlibs/test"

	"testing"
)

type testItem []byte

func (tI testItem) Hash() []byte {
	return []byte(tI)
}

func TestSimpleProof(t *testing.T) {

	total := 100

	items := make([]Hasher, total)
	for i := 0; i < total; i++ {
		items[i] = testItem(cmn.RandBytes(32))
	}

	rootHash := SimpleHashFromHashers(items)

	rootHash2, proofs := SimpleProofsFromHashers(items)

	if !bytes.Equal(rootHash, rootHash2) {
		t.Errorf("Unmatched root hashes: %X vs %X", rootHash, rootHash2)
	}

	// For each item, check the trail.
	for i, item := range items {
		itemHash := item.Hash()
		proof := proofs[i]

		// Verify success
		ok := proof.Verify(i, total, itemHash, rootHash)
		if !ok {
			t.Errorf("Verification failed for index %v.", i)
		}

		// Wrong item index should make it fail
		{
			ok = proof.Verify((i+1)%total, total, itemHash, rootHash)
			if ok {
				t.Errorf("Expected verification to fail for wrong index %v.", i)
			}
		}

		// Trail too long should make it fail
		origAunts := proof.Aunts
		proof.Aunts = append(proof.Aunts, cmn.RandBytes(32))
		{
			ok = proof.Verify(i, total, itemHash, rootHash)
			if ok {
				t.Errorf("Expected verification to fail for wrong trail length.")
			}
		}
		proof.Aunts = origAunts

		// Trail too short should make it fail
		proof.Aunts = proof.Aunts[0 : len(proof.Aunts)-1]
		{
			ok = proof.Verify(i, total, itemHash, rootHash)
			if ok {
				t.Errorf("Expected verification to fail for wrong trail length.")
			}
		}
		proof.Aunts = origAunts

		// Mutating the itemHash should make it fail.
		ok = proof.Verify(i, total, MutateByteSlice(itemHash), rootHash)
		if ok {
			t.Errorf("Expected verification to fail for mutated leaf hash")
		}

		// Mutating the rootHash should make it fail.
		ok = proof.Verify(i, total, itemHash, MutateByteSlice(rootHash))
		if ok {
			t.Errorf("Expected verification to fail for mutated root hash")
		}
	}
}
