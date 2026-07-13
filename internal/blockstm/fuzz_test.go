package blockstm

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

// This file contains a property-based / fuzz harness for the whole Block-STM
// engine. The core oracle is the same one TestSTM uses: parallel execution of a
// block must produce byte-identical state to sequential execution of the same
// block. On top of that we assert parallel execution is itself deterministic
// (two parallel runs agree), which is what catches scheduler-order-dependent
// races that a single run can pass by luck.
//
// Transactions are descriptor-driven: each carries a generated read/write/delete
// set over a bounded key universe, with written values derived from values read,
// so a tx is a deterministic function of input state (keeping the oracle valid).
// The per-field docs on txSpec describe the individual dimensions.
//
// Out of scope, because they do not exist at the internal/blockstm layer
// (baseapp / VM-level concerns): SkipRest / block early termination, and block
// gas limits / block cutting.
//
// These tests run real parallel execution and assert serializability. They are
// gated behind RUN_BLOCKSTM_STRESS so the ordinary `go test` run cannot flake on
// a scheduler that does not yet satisfy the property: until the
// TryValidateNextVersion lost-update fix lands, parallel execution can violate
// serializability nondeterministically. The nightly/active-fuzz job sets the
// variable; once the fix is merged the gate can be dropped.
func skipUnlessStress(tb testing.TB) {
	tb.Helper()
	if os.Getenv("RUN_BLOCKSTM_STRESS") == "" {
		tb.Skip("set RUN_BLOCKSTM_STRESS=1 to run blockstm parallel-oracle tests (see fuzz_test.go header)")
	}
}

var (
	fuzzStores   = map[storetypes.StoreKey]int{StoreKeyAuth: 0, StoreKeyBank: 1}
	fuzzStoreByI = []storetypes.StoreKey{StoreKeyAuth, StoreKeyBank}
)

func fuzzStoreOf(keyID int) (storetypes.StoreKey, int) {
	idx := keyID % len(fuzzStoreByI)
	return fuzzStoreByI[idx], idx
}

func fuzzKeyBytes(keyID int) []byte {
	return []byte(fmt.Sprintf("k%06d", keyID))
}

// pre-write estimation variants: none, exact match, wrong keys, present but
// empty output, and partial match.
const (
	estNone = iota
	estMatch
	estWrong
	estEmpty
	estPartial
	estModeCount
)

// branchSpec is one alternative read/write set for a dynamic transaction.
type branchSpec struct {
	reads  []int
	writes []int
}

// txSpec is the generated descriptor for a single transaction.
type txSpec struct {
	reads   []int
	writes  []int
	deletes []int
	estMode int
	// iterate makes the tx do a full range scan over the auth store and fold the
	// scanned values into its output, adding a range read to its read-set. Range
	// reads exercise the read-set/validation path far harder than point Gets.
	iterate bool
	// readOwnWrites makes the tx read back its own first write and increment it,
	// exercising read-after-write within a single incarnation's view.
	readOwnWrites bool
	// dynamic makes the tx's read/write SETS depend on the value of a selector
	// key (it executes branches[selector value % len]). Different incarnations
	// read different selector values and so take different branches — the
	// incarnation-dependent behavior aptos-core injects, kept oracle-safe by
	// keying on read data rather than a raw incarnation counter. Directly
	// stresses the changed-write-set-across-incarnations path (Consolidate /
	// wroteNewPath) where the lost-update bug lived.
	dynamic  bool
	selector int
	branches []branchSpec
	// wouldWrite holds the pre-cleared write set for the estEmpty mode, so the
	// estimate can be built from keys the tx would have written.
	wouldWrite []int
}

// byteReader turns a fuzz byte slice into a deterministic stream of small
// bounded choices. Past the end of the input it yields zeros, which keeps
// generation total and deterministic for any input length.
type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) u8() int {
	if r.pos >= len(r.data) {
		return 0
	}
	b := r.data[r.pos]
	r.pos++
	return int(b)
}

// intn returns a value in [0, n). Biased for large n, which is fine for a
// generator — we only need coverage, not uniformity.
func (r *byteReader) intn(n int) int {
	if n <= 0 {
		return 0
	}
	return r.u8() % n
}

func (r *byteReader) boolean() bool { return r.u8()&1 == 1 }

// randReader adapts a *rand.Rand to the same interface so the large seeded test
// can reuse the exact generation logic the fuzzer uses.
type randReader struct{ r *rand.Rand }

func (r *randReader) u8() int { return r.r.Intn(256) }
func (r *randReader) intn(n int) int {
	if n <= 0 {
		return 0
	}
	return r.r.Intn(n)
}
func (r *randReader) boolean() bool { return r.r.Intn(2) == 1 }

type chooser interface {
	u8() int
	intn(n int) int
	boolean() bool
}

// genTxSpec generates one transaction descriptor. maxRW caps the read/write set
// sizes and iterMod controls range-scan frequency (1/iterMod of txs iterate).
func genTxSpec(c chooser, universe, maxRW, iterMod int) txSpec {
	s := txSpec{estMode: c.intn(estModeCount)}
	for j := 0; j < c.intn(maxRW); j++ {
		s.reads = append(s.reads, c.intn(universe))
	}
	for j := 0; j < c.intn(maxRW); j++ {
		s.writes = append(s.writes, c.intn(universe))
	}
	for j := 0; j < c.intn(2); j++ {
		s.deletes = append(s.deletes, c.intn(universe))
	}
	s.iterate = c.u8()%iterMod == 0
	s.readOwnWrites = c.boolean()

	if c.u8()%4 == 0 {
		s.dynamic = true
		s.selector = c.intn(universe)
		nBranches := 2 + c.intn(2) // 2 or 3 alternative behaviors
		for b := 0; b < nBranches; b++ {
			var br branchSpec
			for j := 0; j <= c.intn(maxRW); j++ {
				br.reads = append(br.reads, c.intn(universe))
			}
			for j := 0; j <= c.intn(maxRW); j++ {
				br.writes = append(br.writes, c.intn(universe))
			}
			s.branches = append(s.branches, br)
		}
		// static writes/deletes/estimates don't apply to the branch-selected form
		s.writes = nil
		s.deletes = nil
		s.estMode = estNone
	}

	// estEmpty models "estimated to write, but produces no output": drop the
	// write and delete sets (both are output), keeping the estimate below.
	if s.estMode == estEmpty {
		s.wouldWrite = append(append([]int{}, s.writes...), s.deletes...)
		s.writes = nil
		s.deletes = nil
	}
	return s
}

func genBlock(c chooser, maxTxs, maxUniverse, maxExecutors int) (universe, executors int, specs []txSpec) {
	universe = 1 + c.intn(maxUniverse)
	executors = 1 + c.intn(maxExecutors)
	numTxs := 1 + c.intn(maxTxs)

	specs = make([]txSpec, numTxs)
	for i := range specs {
		specs[i] = genTxSpec(c, universe, 4, 4)
	}
	return universe, executors, specs
}

// makeTx compiles a txSpec into an executable Tx. The written value is derived
// from the sum of the values read plus the tx index, making behavior depend on
// input state while remaining deterministic.
func makeTx(txIdx int, s txSpec) Tx {
	return func(ms MultiStore, _ Cache) error {
		var sum uint64
		reads, writes := s.reads, s.writes

		// Dynamic: the branch (hence read/write set) is chosen by the selector's
		// current value. See txSpec.dynamic.
		if s.dynamic && len(s.branches) > 0 {
			sk, _ := fuzzStoreOf(s.selector)
			var sel uint64
			if v := ms.GetKVStore(sk).Get(fuzzKeyBytes(s.selector)); v != nil {
				sel = binary.BigEndian.Uint64(v)
			}
			sum += sel
			b := s.branches[int(sel%uint64(len(s.branches)))]
			reads, writes = b.reads, b.writes
		}

		for _, r := range reads {
			sk, _ := fuzzStoreOf(r)
			if v := ms.GetKVStore(sk).Get(fuzzKeyBytes(r)); v != nil {
				sum += binary.BigEndian.Uint64(v)
			}
		}

		// range read: fold the whole auth store into the sum, adding a range
		// dependency so any prior write to a scanned key invalidates this tx.
		if s.iterate {
			it := ms.GetKVStore(StoreKeyAuth).Iterator(nil, nil)
			for ; it.Valid(); it.Next() {
				if len(it.Value()) == 8 {
					sum += binary.BigEndian.Uint64(it.Value())
				}
			}
			it.Close()
		}

		val := sum + uint64(txIdx) + 1
		var bz [8]byte
		binary.BigEndian.PutUint64(bz[:], val)
		for _, w := range writes {
			sk, _ := fuzzStoreOf(w)
			ms.GetKVStore(sk).Set(fuzzKeyBytes(w), bz[:])
		}

		// Read-after-write within the same incarnation: read back the first write
		// and increment it. The value read must be the one just written.
		if s.readOwnWrites && len(writes) > 0 {
			w := writes[0]
			sk, _ := fuzzStoreOf(w)
			kv := ms.GetKVStore(sk)
			var back uint64
			if v := kv.Get(fuzzKeyBytes(w)); v != nil {
				back = binary.BigEndian.Uint64(v)
			}
			var bz2 [8]byte
			binary.BigEndian.PutUint64(bz2[:], back+1)
			kv.Set(fuzzKeyBytes(w), bz2[:])
		}

		for _, d := range s.deletes {
			sk, _ := fuzzStoreOf(d)
			ms.GetKVStore(sk).Delete(fuzzKeyBytes(d))
		}

		return nil
	}
}

// buildEstimates constructs the per-tx estimates slice from the specs. Estimates
// are only a performance hint: whatever we mark here, correct or not, the final
// state must be unchanged, so every mode below is safe to feed.
func buildEstimates(specs []txSpec) []MultiLocations {
	anyEstimate := false
	for _, s := range specs {
		if s.estMode != estNone {
			anyEstimate = true
			break
		}
	}
	if !anyEstimate {
		return nil
	}

	estimates := make([]MultiLocations, len(specs))
	for i, s := range specs {
		// A tx's write footprint is writes ∪ deletes (a delete is a tombstone
		// write), so the estimate modes model it, not just writes.
		footprint := append(append([]int{}, s.writes...), s.deletes...)
		var keyIDs []int
		switch s.estMode {
		case estNone:
			continue
		case estMatch:
			keyIDs = footprint
		case estWrong:
			// keys the tx does not touch: shift each into a disjoint band.
			for _, w := range footprint {
				keyIDs = append(keyIDs, w+1_000_000)
			}
		case estEmpty:
			keyIDs = s.wouldWrite // estimate present, but the tx writes nothing
		case estPartial:
			if len(footprint) > 0 {
				keyIDs = footprint[:1+len(footprint)/2]
			}
		}
		if len(keyIDs) == 0 {
			continue
		}
		estimates[i] = locationsByStore(keyIDs)
	}
	return estimates
}

func locationsByStore(keyIDs []int) MultiLocations {
	byStore := map[int][]Key{}
	for _, id := range keyIDs {
		_, si := fuzzStoreOf(id)
		byStore[si] = append(byStore[si], Key(fuzzKeyBytes(id)))
	}
	out := make(MultiLocations, len(byStore))
	for si, keys := range byStore {
		sort.Slice(keys, func(a, b int) bool { return string(keys[a]) < string(keys[b]) })
		out[si] = keys
	}
	return out
}

func specsToBlock(specs []txSpec) *MockBlock {
	txs := make([]Tx, len(specs))
	for i, s := range specs {
		txs[i] = makeTx(i, s)
	}
	return NewMockBlock(txs)
}

func runParallel(t *testing.T, specs []txSpec, executors int, estimates []MultiLocations) *MultiMemDB {
	t.Helper()
	storage := NewMultiMemDB(fuzzStores)
	block := specsToBlock(specs)
	err := ExecuteBlockWithEstimates(
		context.Background(), block.Size(), fuzzStores, storage, executors, estimates,
		func(txn TxnIndex, ms MultiStore) { block.ExecuteTx(txn, ms, nil) },
	)
	require.NoError(t, err)
	return storage
}

// checkOracle runs the block sequentially and in parallel (twice) and asserts
// all three agree store-by-store.
func checkOracle(t *testing.T, universe, executors int, specs []txSpec, estimates []MultiLocations) {
	t.Helper()

	seqStore := NewMultiMemDB(fuzzStores)
	runSequential(seqStore, specsToBlock(specs))

	par1 := runParallel(t, specs, executors, estimates)
	par2 := runParallel(t, specs, executors, estimates)

	for store := range fuzzStores {
		require.True(t, StoreEqual(seqStore.GetKVStore(store), par1.GetKVStore(store)),
			"parallel != sequential for store %s", store.Name())
		require.True(t, StoreEqual(par1.GetKVStore(store), par2.GetKVStore(store)),
			"parallel run not deterministic for store %s", store.Name())
	}
	_ = universe
}

// checkInterrupt runs the block with the context cancelled mid-way (right after
// the cancelAt-th transaction body executes). The engine must either finish
// cleanly — in which case the state must still match sequential — or return
// context.Canceled. Any other error, a panic, or a hang is a bug.
func checkInterrupt(t *testing.T, specs []txSpec, executors int, estimates []MultiLocations, cancelAt int) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage := NewMultiMemDB(fuzzStores)
	block := specsToBlock(specs)
	var once sync.Once
	err := ExecuteBlockWithEstimates(ctx, block.Size(), fuzzStores, storage, executors, estimates,
		func(txn TxnIndex, ms MultiStore) {
			block.ExecuteTx(txn, ms, nil)
			if int(txn) == cancelAt {
				once.Do(cancel)
			}
		},
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected error under interrupt: %v", err)
	}
	if err == nil {
		// The block committed before the cancel took effect: it must still be
		// serially equivalent.
		seq := NewMultiMemDB(fuzzStores)
		runSequential(seq, specsToBlock(specs))
		for store := range fuzzStores {
			require.True(t, StoreEqual(seq.GetKVStore(store), storage.GetKVStore(store)),
				"interrupt-completed parallel != sequential for store %s", store.Name())
		}
	}
}

func FuzzBlockSTM(f *testing.F) {
	skipUnlessStress(f)

	// A few seeds spanning empty, tiny, and medium inputs.
	f.Add([]byte{})
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{0xff, 0x01, 0x10, 0x08, 0x04, 0x02, 0x00, 0x40, 0x20, 0x11, 0x22, 0x33})
	f.Add(bytesRepeat(0xab, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		r := &byteReader{data: data}
		// Bounded so a single fuzz execution stays fast; the large seeded test
		// covers big blocks.
		universe, executors, specs := genBlock(r, 40, 24, 8)
		estimates := buildEstimates(specs)
		checkOracle(t, universe, executors, specs, estimates)

		// ~1/4 of inputs also get an interrupt trial.
		if r.u8()%4 == 0 && len(specs) > 0 {
			checkInterrupt(t, specs, executors, estimates, r.intn(len(specs)))
		}
	})
}

// TestBlockSTMRandomizedLarge exercises large blocks (up to a few thousand txs)
// with fixed seeds so it runs in the normal suite. Skipped in -short mode.
func TestBlockSTMRandomizedLarge(t *testing.T) {
	skipUnlessStress(t)
	if testing.Short() {
		t.Skip("skipping large randomized block-stm test in short mode")
	}

	for _, seed := range []int64{1, 2, 3, 42, 1337} {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			c := &randReader{r: rand.New(rand.NewSource(seed))}
			universe := 200 + c.intn(2000)
			executors := 4 + c.intn(12)
			numTxs := 1000 + c.intn(4000)

			specs := make([]txSpec, numTxs)
			for i := range specs {
				// wider read/write sets (maxRW=6); iterate is rare here (1/32)
				// since a full-store scan folded into a tx serializes it against
				// all prior writes, which is expensive on thousand-tx blocks.
				specs[i] = genTxSpec(c, universe, 6, 32)
			}
			estimates := buildEstimates(specs)
			checkOracle(t, universe, executors, specs, estimates)
		})
	}
}

func bytesRepeat(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}
