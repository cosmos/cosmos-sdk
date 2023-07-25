package testutil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DiffCollectionsMigration is meant to aid in the migration from the previous store
// solution to collections. It takes a few steps to use it:
//  1. Write a test function that writes elements using the old store solution.
//  2. Run the test function and copy the hash that it outputs (run it a couple
//     of times to make sure no non-deterministic data is being used).
//  3. Change the code to use collections and run the test function again to make
//     sure the hash didn't change.
func DiffCollectionsMigration(
	ctx sdk.Context,
	storeKey *storetypes.KVStoreKey,
	iterations int,
	writeElem func(int64),
	targetHash string,
) error {
	for i := int64(0); i < int64(iterations); i++ {
		writeElem(i)
	}

	allkvs := []byte{}
	it := ctx.KVStore(storeKey).Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		kv := append(it.Key(), it.Value()...)
		allkvs = append(allkvs, kv...)
	}

	hash := sha256.Sum256(allkvs)
	if hex.EncodeToString(hash[:]) != targetHash {
		return fmt.Errorf("hashes don't match: %s != %s", hex.EncodeToString(hash[:]), targetHash)
	}

	return nil
}
