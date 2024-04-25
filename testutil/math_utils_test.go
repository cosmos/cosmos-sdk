package testutil_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	math "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestDiffDecimalsMigration(t *testing.T) {
	key := storetypes.NewKVStoreKey("test")
	ctx := testutil.DefaultContext(key, storetypes.NewTransientStoreKey("transient"))
	for i := int64(0); i < 5; i++ {
		legacyDec := math.LegacyNewDec(i)
		dec := math.NewDecFromInt64(i)
		ctx.KVStore(key).Set([]byte(fmt.Sprintf("legacy_%d", i)), []byte(legacyDec.String()))
		ctx.KVStore(key).Set([]byte(fmt.Sprintf("new_%d", i)), []byte(dec.String()))
	}

	hashLegacy := computeHash(ctx, key, "legacy_")
	hashNew := computeHash(ctx, key, "new_")
	require.Equal(t, hashLegacy, hashNew, "Hashes do not match")
}

func computeHash(ctx sdk.Context, key storetypes.StoreKey, prefix string) string {
	h := sha256.New()
	start, end := prefixRange(prefix)
	it := ctx.KVStore(key).Iterator(start, end)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		h.Write(it.Key())
		h.Write(it.Value())
	}
	return hex.EncodeToString(h.Sum(nil))
}

func prefixRange(prefix string) (start, end []byte) {
	return []byte(prefix), append([]byte(prefix), 0xFF)
}
