package testutil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DiffDecimalsMigration(
	ctx sdk.Context,
	storeKey *storetypes.KVStoreKey,
	iterations int,
	migrateDec func(int64),
	targetHash string,
) error {
	for i := int64(0); i < int64(iterations); i++ {
		migrateDec(i)
	}

	h := sha256.New()
	it := ctx.KVStore(storeKey).Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		h.Write(it.Key())
		h.Write(it.Value())
	}

	hash := h.Sum(nil)
	if hex.EncodeToString(hash) != targetHash {
		return fmt.Errorf("hashes don't match: %s != %s", hex.EncodeToString(hash), targetHash)
	}

	return nil
}
