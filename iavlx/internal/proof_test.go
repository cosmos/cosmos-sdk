package internal

import (
	"fmt"
	"testing"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// buildTestTree builds a tree from key-value pairs using MutationContext and SetRecursive.
// Returns the root NodePointer and the version used.
func buildTestTree(t require.TestingT, kvPairs [][2][]byte) (*NodePointer, uint32) {
	version := uint32(1)
	ctx := NewMutationContext(version, version)

	var root *NodePointer
	for _, kv := range kvPairs {
		key, value := kv[0], kv[1]
		leafNode := ctx.NewLeafNode(key, value)
		newRoot, _, err := SetRecursive(root, leafNode, ctx)
		require.NoError(t, err)
		root = newRoot
	}

	// Compute hashes on all nodes
	if root != nil {
		rootMem := root.Mem.Load()
		require.NotNil(t, rootMem, "root should be a MemNode")
		_, err := rootMem.ComputeHash(SyncHashScheduler{})
		require.NoError(t, err)
	}

	return root, version
}

func TestTree_MembershipProof(t *testing.T) {
	rapid.Check(t, testTreeMembershipProof)
}

func testTreeMembershipProof(t *rapid.T) {
	// Generate a random number of key-value pairs
	numKeys := rapid.IntRange(1, 200).Draw(t, "numKeys")
	keys := make([][]byte, 0, numKeys)
	keySet := make(map[string]bool) // Track keys to avoid duplicates
	kvPairs := make([][2][]byte, 0, numKeys)

	for i := range numKeys {
		// Generate random key and value with varying lengths (non-empty)
		keyLen := rapid.IntRange(1, 50).Draw(t, fmt.Sprintf("keyLen-%d", i))
		valueLen := rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("valueLen-%d", i))
		key := rapid.SliceOfN(rapid.Byte(), keyLen, keyLen).Draw(t, fmt.Sprintf("key-%d", i))
		value := rapid.SliceOfN(rapid.Byte(), valueLen, valueLen).Draw(t, fmt.Sprintf("value-%d", i))

		// Skip duplicates
		if keySet[string(key)] {
			continue
		}
		keySet[string(key)] = true

		kvPairs = append(kvPairs, [2][]byte{key, value})
		keys = append(keys, key)
	}

	// Build the tree
	root, version := buildTestTree(t, kvPairs)
	treeReader := NewTreeReader(version, root)

	// Property: For all keys that were inserted, membership proofs should be valid
	for _, key := range keys {
		proof, err := treeReader.GetMembershipProof(key)
		require.NoError(t, err, "failed to get membership proof for key %X", key)

		ok, err := treeReader.VerifyMembership(proof, key)
		require.NoError(t, err, "failed to verify membership for key %X", key)
		require.True(t, ok, "membership verification failed for key %X", key)
	}
}

func TestTree_NonMembershipProof(t *testing.T) {
	rapid.Check(t, testTreeNonMembershipProof)
}

func testTreeNonMembershipProof(t *rapid.T) {
	// Generate a random number of key-value pairs and track them
	numKeys := rapid.IntRange(1, 200).Draw(t, "numKeys")
	keySet := make(map[string]bool)
	kvPairs := make([][2][]byte, 0, numKeys)

	for i := range numKeys {
		// Generate random key and value with varying lengths (non-empty values)
		keyLen := rapid.IntRange(1, 50).Draw(t, fmt.Sprintf("keyLen-%d", i))
		valueLen := rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("valueLen-%d", i))
		key := rapid.SliceOfN(rapid.Byte(), keyLen, keyLen).Draw(t, fmt.Sprintf("key-%d", i))
		value := rapid.SliceOfN(rapid.Byte(), valueLen, valueLen).Draw(t, fmt.Sprintf("value-%d", i))

		keySet[string(key)] = true
		kvPairs = append(kvPairs, [2][]byte{key, value})
	}

	// Build the tree
	root, version := buildTestTree(t, kvPairs)
	treeReader := NewTreeReader(version, root)

	// Property: For keys that are NOT in the tree, non-membership proofs should be valid
	numNonMembershipTests := rapid.IntRange(1, 100).Draw(t, "numNonMembershipTests")
	for i := range numNonMembershipTests {
		// Generate random keys that are not in the keySet
		var nonExistentKey []byte
		maxAttempts := 100
		for attempt := 0; attempt < maxAttempts; attempt++ {
			keyLen := rapid.IntRange(1, 100).Draw(t, fmt.Sprintf("nonMemberKeyLen-%d-%d", i, attempt))
			candidate := rapid.SliceOfN(rapid.Byte(), keyLen, keyLen).Draw(t, fmt.Sprintf("nonMemberKey-%d-%d", i, attempt))

			// Check if this key is not in the tree
			if !keySet[string(candidate)] {
				nonExistentKey = candidate
				break
			}
		}

		// If we couldn't find a non-existent key after max attempts, skip this iteration
		if nonExistentKey == nil {
			continue
		}

		proof, err := treeReader.GetNonMembershipProof(nonExistentKey)
		require.NoError(t, err, "failed to get non-membership proof for key %X", nonExistentKey)
		require.NotNil(t, proof, "non-membership proof is nil for key %X", nonExistentKey)

		valid, err := treeReader.VerifyNonMembership(proof, nonExistentKey)
		require.NoError(t, err, "failed to verify non-membership for key %X", nonExistentKey)
		require.True(t, valid, "non-membership verification failed for key %X", nonExistentKey)
	}
}

func TestTreeReaderProofsNilRoot(t *testing.T) {
	treeReader := NewTreeReader(0, nil)

	_, err := treeReader.GetMembershipProof([]byte("key"))
	require.Error(t, err)

	_, err = treeReader.GetNonMembershipProof([]byte("key"))
	require.Error(t, err)

	_, err = treeReader.VerifyMembership(&ics23.CommitmentProof{}, []byte("key"))
	require.Error(t, err)

	_, err = treeReader.VerifyNonMembership(&ics23.CommitmentProof{}, []byte("key"))
	require.Error(t, err)
}
