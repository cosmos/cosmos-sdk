package gaskv_test

import (
	"fmt"
	"sort"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/v2/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/v2/gaskv"
	"github.com/cosmos/cosmos-sdk/store/v2/types"
)

func bz(s string) []byte { return []byte(s) }

func keyFmt(i int) []byte { return bz(fmt.Sprintf("key%0.8d", i)) }
func valFmt(i int) []byte { return bz(fmt.Sprintf("value%0.8d", i)) }

func TestGasKVStoreBasic(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	meter := types.NewGasMeter(10000)
	st := gaskv.NewStore(mem, meter, types.KVGasConfig())

	require.Equal(t, types.StoreTypeDB, st.GetStoreType())
	require.Panics(t, func() { st.CacheWrap() })

	require.Panics(t, func() { st.Set(nil, []byte("value")) }, "setting a nil key should panic")
	require.Panics(t, func() { st.Set([]byte(""), []byte("value")) }, "setting an empty key should panic")

	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	st.Set(keyFmt(1), valFmt(1))
	require.Equal(t, valFmt(1), st.Get(keyFmt(1)))
	st.Delete(keyFmt(1))
	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	require.Equal(t, meter.GasConsumed(), types.Gas(6858))
}

func TestGasKVStoreIterator(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	meter := types.NewGasMeter(100000)
	st := gaskv.NewStore(mem, meter, types.KVGasConfig())
	require.False(t, st.Has(keyFmt(1)))
	require.Empty(t, st.Get(keyFmt(1)), "Expected `key1` to be empty")
	require.Empty(t, st.Get(keyFmt(2)), "Expected `key2` to be empty")
	require.Empty(t, st.Get(keyFmt(3)), "Expected `key3` to be empty")

	st.Set(keyFmt(1), valFmt(1))
	require.True(t, st.Has(keyFmt(1)))
	st.Set(keyFmt(2), valFmt(2))
	require.True(t, st.Has(keyFmt(2)))
	st.Set(keyFmt(3), valFmt(0))

	iterator := st.Iterator(nil, nil)
	start, end := iterator.Domain()
	require.Nil(t, start)
	require.Nil(t, end)
	require.NoError(t, iterator.Error())

	t.Cleanup(func() {
		if err := iterator.Close(); err != nil {
			t.Fatal(err)
		}
	})
	ka := iterator.Key()
	require.Equal(t, ka, keyFmt(1))
	va := iterator.Value()
	require.Equal(t, va, valFmt(1))
	iterator.Next()
	kb := iterator.Key()
	require.Equal(t, kb, keyFmt(2))
	vb := iterator.Value()
	require.Equal(t, vb, valFmt(2))
	iterator.Next()
	require.Equal(t, types.Gas(14565), meter.GasConsumed())
	kc := iterator.Key()
	require.Equal(t, kc, keyFmt(3))
	vc := iterator.Value()
	require.Equal(t, vc, valFmt(0))
	iterator.Next()
	require.Equal(t, types.Gas(14595), meter.GasConsumed())
	require.False(t, iterator.Valid())
	require.Panics(t, iterator.Next)
	require.Equal(t, types.Gas(14595), meter.GasConsumed())
	require.NoError(t, iterator.Error())

	reverseIterator := st.ReverseIterator(nil, nil)
	t.Cleanup(func() {
		if err := reverseIterator.Close(); err != nil {
			t.Fatal(err)
		}
	})
	require.Equal(t, reverseIterator.Key(), keyFmt(3))
	reverseIterator.Next()
	require.Equal(t, reverseIterator.Key(), keyFmt(2))
	reverseIterator.Next()
	require.Equal(t, reverseIterator.Key(), keyFmt(1))
	reverseIterator.Next()
	require.False(t, reverseIterator.Valid())
	require.Panics(t, reverseIterator.Next)
	require.Equal(t, types.Gas(14931), meter.GasConsumed())
}

// TestGasKVStoreIteratorChargesEachElementOnce ensures iterating a KVStore meters gas
// proportional to each key/value pair exactly once: the first pair is charged when the
// iterator is created, and each subsequent pair is charged by the Next() call that lands
// on it. Regression test for #15854, where the first pair was charged twice and the
// iterator never metered the pair it advanced onto.
func TestGasKVStoreIteratorChargesEachElementOnce(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	meter := types.NewGasMeter(1_000_000)
	st := gaskv.NewStore(mem, meter, types.KVGasConfig())

	// Distinct key/value lengths so that a test using a single uniform size could
	// not distinguish "charge the right pair at the right step" from "charge every
	// pair with the first pair's size" or "charge the right total at the wrong step".
	type entry struct {
		key []byte
		val []byte
	}
	entries := []entry{
		{bz("k1"), bz("v")},
		{bz("k2"), bz("vverylongvalue")},
		{bz("k3"), bz("vmedium")},
		{bz("k4"), bz("v")},
		{bz("k5"), bz("valonggervalue")},
	}
	for _, e := range entries {
		st.Set(e.key, e.val)
	}

	gasConfig := types.KVGasConfig()
	pairCost := func(e entry) types.Gas {
		return gasConfig.ReadCostPerByte * types.Gas(len(e.key)+len(e.val))
	}
	// entries sorted ascending by key (the order a forward iterator visits them).
	sorted := make([]entry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool { return string(sorted[i].key) < string(sorted[j].key) })

	// totalCost is the byte cost of metering every pair once; identical for a full
	// forward and a full reverse scan since the same pairs are visited.
	totalCost := types.Gas(0)
	for _, e := range sorted {
		totalCost += pairCost(e)
	}

	// Forward: the iterator is created already positioned on the first valid pair
	// (seek #1 charges that pair), and each Next() that lands on a valid pair
	// charges the newly reached pair. Only the final Next(), which lands past the
	// end, charges just the flat cost. Every element is therefore metered exactly
	// once, at the step that presents it.
	gasBefore := meter.GasConsumed()
	iter := st.Iterator(nil, nil)
	t.Cleanup(func() { _ = iter.Close() })

	// Seek #1 (creation) meters pair 0.
	require.True(t, iter.Valid())
	require.Equal(t, types.Gas(1)*gasConfig.IterNextCostFlat+pairCost(sorted[0]), meter.GasConsumed()-gasBefore)

	cumulative := types.Gas(1)*gasConfig.IterNextCostFlat + pairCost(sorted[0])
	seen := 1
	for iter.Valid() {
		seen++
		idx := seen - 1 // pair we are about to land on
		iter.Next()
		if iter.Valid() {
			cumulative += gasConfig.IterNextCostFlat + pairCost(sorted[idx])
		} else {
			cumulative += gasConfig.IterNextCostFlat // final seek lands past the end: flat only
		}
		require.Equal(t, cumulative, meter.GasConsumed()-gasBefore, "forward gas after step %d", idx)
	}
	require.Equal(t, len(entries), seen-1)
	// Full forward scan: n pairs metered once each, plus the flat cost of the n+1
	// seeks (creation + one Next() per pair, the last landing past the end).
	require.Equal(t, types.Gas(len(entries)+1)*gasConfig.IterNextCostFlat+totalCost, meter.GasConsumed()-gasBefore)

	// Early stop: charge exactly the pairs that were presented, no more.
	gasBefore = meter.GasConsumed()
	early := st.Iterator(nil, nil)
	t.Cleanup(func() { _ = early.Close() })
	expect := types.Gas(1)*gasConfig.IterNextCostFlat + pairCost(sorted[0])
	steps := 0
	for early.Valid() && steps < 2 {
		_ = early.Key()
		_ = early.Value()
		early.Next()
		steps++
		if early.Valid() {
			expect += gasConfig.IterNextCostFlat + pairCost(sorted[steps])
		} else {
			expect += gasConfig.IterNextCostFlat
		}
	}
	require.Equal(t, expect, meter.GasConsumed()-gasBefore, "early-stop gas")

	// Reverse: identical accounting, walking the pairs in descending key order.
	desc := make([]entry, len(sorted))
	for i, e := range sorted {
		desc[len(sorted)-1-i] = e
	}
	gasBefore = meter.GasConsumed()
	revIter := st.ReverseIterator(nil, nil)
	t.Cleanup(func() { _ = revIter.Close() })
	require.True(t, revIter.Valid())
	require.Equal(t, types.Gas(1)*gasConfig.IterNextCostFlat+pairCost(desc[0]), meter.GasConsumed()-gasBefore)

	cumulative = types.Gas(1)*gasConfig.IterNextCostFlat + pairCost(desc[0])
	seen = 1
	for revIter.Valid() {
		seen++
		idx := seen - 1
		revIter.Next()
		if revIter.Valid() {
			cumulative += gasConfig.IterNextCostFlat + pairCost(desc[idx])
		} else {
			cumulative += gasConfig.IterNextCostFlat
		}
		require.Equal(t, cumulative, meter.GasConsumed()-gasBefore, "reverse gas after step %d", idx)
	}
	require.Equal(t, len(entries), seen-1)
	require.Equal(t, types.Gas(len(entries)+1)*gasConfig.IterNextCostFlat+totalCost, meter.GasConsumed()-gasBefore)
}

func TestGasKVStoreOutOfGasSet(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	meter := types.NewGasMeter(0)
	st := gaskv.NewStore(mem, meter, types.KVGasConfig())
	require.Panics(t, func() { st.Set(keyFmt(1), valFmt(1)) }, "Expected out-of-gas")
}

func TestGasKVStoreOutOfGasIterator(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	meter := types.NewGasMeter(20000)
	st := gaskv.NewStore(mem, meter, types.KVGasConfig())
	st.Set(keyFmt(1), valFmt(1))
	iterator := st.Iterator(nil, nil)
	iterator.Next()
	require.Panics(t, func() { iterator.Value() }, "Expected out-of-gas")
}
