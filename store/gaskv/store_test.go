package gaskv_test

import (
	"fmt"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/dbadapter"
	"cosmossdk.io/store/gaskv"
	"cosmossdk.io/store/types"
)

func bz(s string) []byte { return []byte(s) }

func keyFmt(i int) []byte { return bz(fmt.Sprintf("key%0.8d", i)) }
func valFmt(i int) []byte { return bz(fmt.Sprintf("value%0.8d", i)) }

// TestSafeMul tests the safeMul function for various scenarios
func TestSafeMul(t *testing.T) {
	// Test normal cases
	t.Run("normal cases", func(t *testing.T) {
		// Test basic multiplication
		result, err := gaskv.SafeMul(10, 5)
		require.NoError(t, err)
		require.Equal(t, types.Gas(50), result)

		// Test with zero cost
		result, err = gaskv.SafeMul(0, 1000)
		require.NoError(t, err)
		require.Equal(t, types.Gas(0), result)

		// Test with zero length
		result, err = gaskv.SafeMul(1000, 0)
		require.NoError(t, err)
		require.Equal(t, types.Gas(0), result)

		// Test with both zero
		result, err = gaskv.SafeMul(0, 0)
		require.NoError(t, err)
		require.Equal(t, types.Gas(0), result)

		// Test large but safe values
		result, err = gaskv.SafeMul(1000000, 1000000)
		require.NoError(t, err)
		require.Equal(t, types.Gas(1000000000000), result)
	})

	// Test edge cases
	t.Run("edge cases", func(t *testing.T) {
		// Test maximum uint64 values that don't overflow
		maxUint64 := types.Gas(^uint64(0))
		result, err := gaskv.SafeMul(maxUint64, 1)
		require.NoError(t, err)
		require.Equal(t, maxUint64, result)

		// Test with 1 and a large but safe value
		// Use a value that's safe to convert to int
		safeLargeValue := 1000000000 // 1 billion, safe for int
		result, err = gaskv.SafeMul(1, safeLargeValue)
		require.NoError(t, err)
		require.Equal(t, types.Gas(safeLargeValue), result)
	})

	// Test overflow cases
	t.Run("overflow cases", func(t *testing.T) {
		maxUint64 := types.Gas(^uint64(0))

		// Test overflow: maxUint64 * 2 should overflow
		result, err := gaskv.SafeMul(maxUint64, 2)
		require.Error(t, err)
		require.Contains(t, err.Error(), "gas calculation overflow")
		require.Equal(t, types.Gas(0), result)

		// Test overflow: large values that multiply to overflow
		result, err = gaskv.SafeMul(maxUint64/2+1, 2)
		require.Error(t, err)
		require.Contains(t, err.Error(), "gas calculation overflow")
		require.Equal(t, types.Gas(0), result)

		// Test overflow: choose length that is safely representable as int on 32/64-bit,
		// and a cost that guarantees overflow when multiplied by length.
		// length = 1<<30 is safe for 32-bit; cost = floor(MaxUint64/length) + 1 ensures overflow.
		length := 1 << 30
		overflowCost := types.Gas((^uint64(0))/uint64(length)) + 1
		result, err = gaskv.SafeMul(overflowCost, length)
		require.Error(t, err)
		require.Contains(t, err.Error(), "gas calculation overflow")
		require.Equal(t, types.Gas(0), result)
	})

	// Test negative length
	t.Run("negative length", func(t *testing.T) {
		result, err := gaskv.SafeMul(100, -1)
		require.Error(t, err)
		require.Contains(t, err.Error(), "negative length")
		require.Equal(t, types.Gas(0), result)

		result, err = gaskv.SafeMul(0, -100)
		require.Error(t, err)
		require.Contains(t, err.Error(), "negative length")
		require.Equal(t, types.Gas(0), result)
	})

	// Test boundary cases
	t.Run("boundary cases", func(t *testing.T) {
		maxUint64 := types.Gas(^uint64(0))

		// Test exactly at the boundary (should not overflow)
		// Find a value that when multiplied by 2 equals maxUint64
		boundaryValue := maxUint64 / 2
		result, err := gaskv.SafeMul(boundaryValue, 2)
		require.NoError(t, err)
		// The issue is that maxUint64 is odd, so dividing by 2 loses 1
		// We need to handle this case properly
		if maxUint64%2 == 1 {
			// If maxUint64 is odd, boundaryValue * 2 will be maxUint64 - 1
			require.Equal(t, maxUint64-1, result)
		} else {
			require.Equal(t, maxUint64, result)
		}

		// Test just over the boundary (should overflow)
		// Use a value that's guaranteed to overflow when multiplied by 2
		overflowValue := maxUint64/2 + 1
		result, err = gaskv.SafeMul(overflowValue, 2)
		require.Error(t, err)
		require.Contains(t, err.Error(), "gas calculation overflow")
		require.Equal(t, types.Gas(0), result)
	})
}

// TestSafeMulIntegration tests that safeMul works correctly in actual gas calculations
func TestSafeMulIntegration(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	meter := types.NewGasMeter(1000000)
	st := gaskv.NewStore(mem, meter, types.KVGasConfig())

	// Test with normal sized data
	normalKey := []byte("normal_key")
	normalValue := []byte("normal_value")
	st.Set(normalKey, normalValue)
	value := st.Get(normalKey)
	require.Equal(t, normalValue, value)

	// Test with large data (but not too large to avoid key size limits)
	largeKey := make([]byte, 10000)   // 10KB key
	largeValue := make([]byte, 10000) // 10KB value
	for i := range largeKey {
		largeKey[i] = byte(i % 256)
		largeValue[i] = byte(i % 256)
	}

	// This should work without overflow
	st.Set(largeKey, largeValue)
	retrievedValue := st.Get(largeKey)
	require.Equal(t, largeValue, retrievedValue)

	// Verify gas was consumed (should be a large amount but not overflow)
	gasConsumed := meter.GasConsumed()
	require.Greater(t, gasConsumed, types.Gas(0))
	require.Less(t, gasConsumed, types.Gas(^uint64(0))) // Should not be max uint64
}

func TestGasKVStoreBasic(t *testing.T) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	meter := types.NewGasMeter(10000)
	st := gaskv.NewStore(mem, meter, types.KVGasConfig())

	require.Equal(t, types.StoreTypeDB, st.GetStoreType())
	require.Panics(t, func() { st.CacheWrap() })
	require.Panics(t, func() { st.CacheWrapWithTrace(nil, nil) })

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
	require.Equal(t, types.Gas(14667), meter.GasConsumed())
	require.False(t, iterator.Valid())
	require.Panics(t, iterator.Next)
	require.Equal(t, types.Gas(14697), meter.GasConsumed())
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
	require.Equal(t, types.Gas(15135), meter.GasConsumed())
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
