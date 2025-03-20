package gen

import (
	"bytes"
	"testing"

	"github.com/cespare/xxhash/v2"
	"github.com/stretchr/testify/require"

	benchmarkmodulev1 "cosmossdk.io/api/cosmos/benchmark/module/v1"
)

func Test_Genesis(t *testing.T) {
	params := &benchmarkmodulev1.GeneratorParams{
		Seed:         34,
		BucketCount:  10,
		GenesisCount: 2_000_000,
		KeyMean:      64,
		KeyStdDev:    8,
		ValueMean:    1024,
		ValueStdDev:  256,
	}
	g := NewGenerator(Options{GeneratorParams: params})
	db := make(map[uint64]map[uint64]bool)
	for kv := range g.GenesisSet() {
		if _, ok := db[kv.StoreKey]; !ok {
			db[kv.StoreKey] = make(map[uint64]bool)
		}
		db[kv.StoreKey][kv.Key[0]] = true
	}

	g = NewGenerator(Options{
		GeneratorParams: params,
		InsertWeight:    0.25,
		DeleteWeight:    0.05,
		UpdateWeight:    0.50,
		GetWeight:       0.20,
	}, WithGenesis())
	for range 100_000 {
		sk, op, err := g.Next()
		require.NoError(t, err)
		switch {
		case op.Delete:
			require.True(t, db[sk][op.Seed])
			delete(db[sk], op.Seed)
		case op.ValueLength > 0:
			if op.Exists {
				// update
				require.True(t, db[sk][op.Seed])
			} else {
				// insert
				require.False(t, db[sk][op.Seed])
			}
			db[sk][op.Seed] = true
		case op.ValueLength == 0:
			// get
			require.True(t, db[sk][op.Seed])
		default:
			t.Fatalf("unexpected op: %v", op)
		}
	}

	// Test state Marshal/Unmarshal
	var buf bytes.Buffer
	require.NoError(t, g.state.Marshal(&buf))
	s := &State{}
	require.NoError(t, s.Unmarshal(bytes.NewReader(buf.Bytes())))
	require.Equal(t, len(g.state.Keys), len(s.Keys))
	for i := range g.state.Keys {
		require.Equal(t, len(g.state.Keys[i]), len(s.Keys[i]))
		for j := range g.state.Keys[i] {
			require.Equal(t, g.state.Keys[i][j], s.Keys[i][j])
		}
	}
}

func Test_Genesis_BytesKey(t *testing.T) {
	params := &benchmarkmodulev1.GeneratorParams{
		Seed:         34,
		BucketCount:  10,
		GenesisCount: 2_000_000,
		KeyMean:      64,
		KeyStdDev:    8,
		ValueMean:    1024,
		ValueStdDev:  256,
	}
	g := NewGenerator(Options{GeneratorParams: params})
	db := make(map[uint64]map[uint64]bool)
	for kv := range g.GenesisSet() {
		if _, ok := db[kv.StoreKey]; !ok {
			db[kv.StoreKey] = make(map[uint64]bool)
		}
		key := xxhash.Sum64(Bytes(kv.Key.Seed(), kv.Key.Length()))
		db[kv.StoreKey][key] = true
	}

	g = NewGenerator(Options{
		GeneratorParams: params,
		InsertWeight:    0.25,
		DeleteWeight:    0.05,
		UpdateWeight:    0.50,
		GetWeight:       0.20,
	}, WithGenesis())
	for range 1_000_000 {
		sk, op, err := g.Next()
		require.NoError(t, err)
		key := xxhash.Sum64(Bytes(op.Seed, op.KeyLength))
		switch {
		case op.Delete:
			require.True(t, db[sk][key])
			delete(db[sk], key)
		case op.ValueLength > 0:
			if op.Exists {
				// update
				require.True(t, db[sk][key])
			} else {
				// insert
				require.False(t, db[sk][key])
			}
			db[sk][key] = true
		case op.ValueLength == 0:
			// get
			require.True(t, db[sk][key])
		default:
			t.Fatalf("unexpected op: %v", op)
		}
	}
}

func Test_Bytes_Deterministic(t *testing.T) {
	seed := uint64(12345)
	length := uint64(53)
	expected := Bytes(seed, length)
	for i := 0; i < 100; i++ {
		result := Bytes(seed, length)
		require.Equal(t, expected, result, "Bytes() should be deterministic")
	}
}
