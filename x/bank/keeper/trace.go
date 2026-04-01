package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"unicode/utf8"

	"github.com/cosmos/cosmos-sdk/store/v2/cachekv"
	"github.com/cosmos/cosmos-sdk/store/v2/gaskv"
	iavlstore "github.com/cosmos/cosmos-sdk/store/v2/iavl"
	"github.com/cosmos/cosmos-sdk/store/v2/listenkv"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// encodeValue returns a human-readable string if the bytes are valid printable UTF-8,
// otherwise returns hex-encoded bytes with 0x prefix.
func encodeValue(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	if utf8.Valid(b) {
		printable := true
		for _, r := range string(b) {
			if r < 0x20 || r == 0x7f {
				printable = false
				break
			}
		}
		if printable {
			return string(b)
		}
	}
	return "0x" + hex.EncodeToString(b)
}

// decodeKey attempts to produce a human-readable description of a store key
// based on known prefix conventions for the bank and auth stores.
func decodeKey(storeName string, key []byte) string {
	if len(key) == 0 {
		return ""
	}
	prefix := key[0]
	rest := key[1:]

	switch storeName {
	case "bank":
		return decodeBankKey(prefix, rest)
	case "acc":
		return decodeAuthKey(prefix, rest)
	}
	return ""
}

// decodeBankKey decodes bank store keys by prefix.
//
//	00 | denom                          → Supply
//	01 | denom                          → DenomMetadata
//	02 | len | addr | denom             → Balances (addr non-terminal: length-prefixed)
//	03 | denom | 0x00 | len | addr      → DenomAddress index (addr length-prefixed)
//	04 | denom                          → SendEnabled
//	05 | ...                            → Params
func decodeBankKey(prefix byte, rest []byte) string {
	switch prefix {
	case 0x00:
		return fmt.Sprintf("supply/%s", string(rest))
	case 0x01:
		return fmt.Sprintf("denom_metadata/%s", string(rest))
	case 0x02:
		// Balances: len(addr) | addr_bytes | denom_string
		if len(rest) < 1 {
			return ""
		}
		addrLen := int(rest[0])
		if len(rest) < 1+addrLen {
			return ""
		}
		addr := sdk.AccAddress(rest[1 : 1+addrLen])
		denom := string(rest[1+addrLen:])
		return fmt.Sprintf("balances/%s/%s", addr, denom)
	case 0x03:
		// DenomAddress index: denom_string | 0x00 | len(addr) | addr_bytes
		nullIdx := bytes.IndexByte(rest, 0x00)
		if nullIdx < 0 {
			return ""
		}
		denom := string(rest[:nullIdx])
		after := rest[nullIdx+1:]
		if len(after) < 1 {
			return fmt.Sprintf("denom_index/%s", denom)
		}
		addrLen := int(after[0])
		if len(after) < 1+addrLen {
			return ""
		}
		addr := sdk.AccAddress(after[1 : 1+addrLen])
		return fmt.Sprintf("denom_index/%s/%s", denom, addr)
	case 0x04:
		return fmt.Sprintf("send_enabled/%s", string(rest))
	case 0x05:
		return "params"
	}
	return ""
}

// decodeAuthKey decodes auth store keys by prefix.
//
//	01 | addr_bytes (20 bytes) → Accounts
func decodeAuthKey(prefix byte, rest []byte) string {
	switch prefix {
	case 0x00:
		return "params"
	case 0x01:
		// Accounts: raw addr bytes (terminal, 20 bytes)
		if len(rest) == 20 {
			addr := sdk.AccAddress(rest)
			return fmt.Sprintf("account/%s", addr)
		}
		return fmt.Sprintf("account/0x%s", hex.EncodeToString(rest))
	case 0x02:
		return "global_account_number"
	}
	return ""
}

// ContextWithTrace stores a SendTrace in the context for annotation access by any keeper.
func ContextWithTrace(ctx context.Context, trace *SendTrace) context.Context {
	return sdk.ContextWithSendTrace(ctx, trace)
}

// AnnotateTrace sets a human-readable annotation on the current trace.
// Delegates to the SDK-level helper so auth and other keepers can also annotate.
func AnnotateTrace(ctx context.Context, annotation string) {
	sdk.AnnotateTrace(ctx, annotation)
}

// TraceEntry records a single store operation linked to a SendCoins call.
type TraceEntry struct {
	Seq        int    `json:"seq"`
	Annotation string `json:"annotation,omitempty"` // human-readable context for this operation
	Op         string `json:"op"`
	Store      string `json:"store"`
	Key        string `json:"key"`                   // hex-encoded raw key bytes
	KeyDecoded string `json:"key_decoded,omitempty"` // human-readable key (e.g. "balances/cosmos1.../uatom")
	Value      string `json:"value"`                 // printable or 0x-prefixed hex
	CacheHit   *bool  `json:"cache_hit"`             // true=cache hit, false=read-through, nil=write op
	CacheLayer *int   `json:"cache_layer,omitempty"` // 0=top cache, 1=parent cache, etc. nil if not a cache hit
	IAVLSource string `json:"iavl_source,omitempty"` // which IAVL layer answered (only for read-through GETs)
}

// SendTrace records all store operations for a single SendCoins invocation.
type SendTrace struct {
	Height int64        `json:"height"`
	From   string       `json:"from"`
	To     string       `json:"to"`
	Amount string       `json:"amount"`
	Ops    []TraceEntry `json:"ops"`

	mu         sync.Mutex
	seq        int
	annotation string // current human-readable annotation
}

// SetAnnotation sets the annotation that will be applied to all subsequent operations.
func (t *SendTrace) SetAnnotation(annotation string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.annotation = annotation
}

func (t *SendTrace) record(op, storeName string, key, value []byte, cacheHit *bool, cacheLayer *int, iavlSource string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Ops = append(t.Ops, TraceEntry{
		Seq:        t.seq,
		Annotation: t.annotation,
		Op:         op,
		Store:      storeName,
		Key:        hex.EncodeToString(key),
		KeyDecoded: decodeKey(storeName, key),
		Value:      encodeValue(value),
		CacheHit:   cacheHit,
		CacheLayer: cacheLayer,
		IAVLSource: iavlSource,
	})
	t.seq++
}

// BlockTraces holds all SendCoins traces for a single block.
type BlockTraces struct {
	Height int64        `json:"height"`
	Traces []*SendTrace `json:"traces"`
}

// TraceRecorder accumulates per-block traces and flushes to a file.
type TraceRecorder struct {
	mu       sync.Mutex
	current  *BlockTraces
	filePath string
}

// NewTraceRecorder creates a recorder that writes to the given file path.
func NewTraceRecorder(filePath string) *TraceRecorder {
	return &TraceRecorder{
		filePath: filePath,
		current:  &BlockTraces{},
	}
}

// NewSendTrace starts tracking a new SendCoins call.
func (r *TraceRecorder) NewSendTrace(height int64, from, to sdk.AccAddress, amt sdk.Coins) *SendTrace {
	trace := &SendTrace{
		Height: height,
		From:   from.String(),
		To:     to.String(),
		Amount: amt.String(),
	}
	r.mu.Lock()
	r.current.Height = height
	r.current.Traces = append(r.current.Traces, trace)
	r.mu.Unlock()
	return trace
}

// FlushAndReset writes the current block's traces to the file and resets for the next block.
func (r *TraceRecorder) FlushAndReset(newHeight int64) {
	r.mu.Lock()
	toWrite := r.current
	r.current = &BlockTraces{Height: newHeight}
	r.mu.Unlock()

	if len(toWrite.Traces) == 0 {
		return
	}

	data, err := json.MarshalIndent(toWrite, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "bank trace: marshal error: %v\n", err)
		return
	}
	if err := os.WriteFile(r.filePath, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "bank trace: write error: %v\n", err)
	}
}

// Tracer is the global trace recorder. Initialized in NewBaseKeeper.
var Tracer *TraceRecorder

// --- TracingMultiStore ---

// TracingMultiStore wraps a MultiStore and intercepts GetKVStore for traced store keys.
type TracingMultiStore struct {
	storetypes.MultiStore
	trace     *SendTrace
	traceKeys map[string]bool // store key names to trace
}

// NewTracingMultiStore creates a tracing wrapper that intercepts the given store key names.
func NewTracingMultiStore(parent storetypes.MultiStore, trace *SendTrace, storeKeyNames ...string) *TracingMultiStore {
	keys := make(map[string]bool, len(storeKeyNames))
	for _, name := range storeKeyNames {
		keys[name] = true
	}
	return &TracingMultiStore{
		MultiStore: parent,
		trace:      trace,
		traceKeys:  keys,
	}
}

// GetKVStore returns a TracingKVStore for traced keys, or the original store otherwise.
func (m *TracingMultiStore) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	inner := m.MultiStore.GetKVStore(key)
	if m.traceKeys[key.Name()] {
		return &TracingKVStore{
			KVStore:   inner,
			trace:     m.trace,
			storeName: key.Name(),
		}
	}
	return inner
}

// CacheMultiStore preserves tracing through cache layers.
func (m *TracingMultiStore) CacheMultiStore() storetypes.CacheMultiStore {
	return m.MultiStore.CacheMultiStore()
}

// --- TracingKVStore ---

// TracingKVStore wraps a KVStore and records all Get/Set/Delete operations.
type TracingKVStore struct {
	storetypes.KVStore
	trace     *SendTrace
	storeName string
}

func (s *TracingKVStore) Get(key []byte) []byte {
	// Check cache hit before the actual Get (which may populate the cache)
	cacheHit, cacheLayer := s.checkCacheHit(key)

	// If this will be a read-through to IAVL, probe which IAVL layer answers
	var iavlSource string
	if cacheHit != nil && !*cacheHit {
		if iavlSt := s.findIAVLStore(); iavlSt != nil {
			_, src := iavlSt.GetWithSource(key)
			iavlSource = string(src)
		}
	}

	value := s.KVStore.Get(key)
	s.trace.record("GET", s.storeName, key, value, cacheHit, cacheLayer, iavlSource)
	return value
}

func (s *TracingKVStore) Set(key, value []byte) {
	s.trace.record("SET", s.storeName, key, value, nil, nil, "")
	s.KVStore.Set(key, value)
}

func (s *TracingKVStore) Delete(key []byte) {
	s.trace.record("DELETE", s.storeName, key, nil, nil, nil, "")
	s.KVStore.Delete(key)
}

func (s *TracingKVStore) Has(key []byte) bool {
	cacheHit, cacheLayer := s.checkCacheHit(key)
	result := s.KVStore.Has(key)
	value := []byte{}
	if result {
		value = []byte{1}
	}
	s.trace.record("HAS", s.storeName, key, value, cacheHit, cacheLayer, "")
	return result
}

// checkCacheHit walks the cachekv chain to find which layer (if any) has the key cached.
// Returns (*bool, *int): cache_hit flag and the layer index (0=top, 1=parent, etc.).
// Layer nil means either not a cache hit or can't determine.
func (s *TracingKVStore) checkCacheHit(key []byte) (*bool, *int) {
	store := s.KVStore
	layer := 0

	for {
		ckvStore, ok := store.(*cachekv.GStore[[]byte])
		if !ok {
			break
		}
		if ckvStore.HasCached(key) {
			hit := true
			return &hit, &layer
		}
		// Walk to the parent cache layer
		parent := ckvStore.Parent()
		if parent == nil {
			break
		}
		// The parent might be another cachekv, a listenkv, or an IAVL store
		parentKV, ok := parent.(storetypes.KVStore)
		if !ok {
			break
		}
		store = parentKV
		layer++
	}

	// Key not found in any cache layer — will read through to IAVL
	miss := false
	return &miss, nil
}

// findIAVLStore walks through the cache/gas wrapper chain to find the underlying IAVL store.
func (s *TracingKVStore) findIAVLStore() *iavlstore.Store {
	var store storetypes.KVStore = s.KVStore
	for {
		switch st := store.(type) {
		case *cachekv.GStore[[]byte]:
			parent := st.Parent()
			if parent == nil {
				return nil
			}
			parentKV, ok := parent.(storetypes.KVStore)
			if !ok {
				return nil
			}
			store = parentKV
		case *gaskv.GStore[[]byte]:
			parent := st.Inner()
			if parent == nil {
				return nil
			}
			parentKV, ok := parent.(storetypes.KVStore)
			if !ok {
				return nil
			}
			store = parentKV
		case *listenkv.Store:
			store = st.Inner()
		case *iavlstore.Store:
			return st
		default:
			return nil
		}
	}
}

// Delegate iterator methods without tracing (not needed for this debug)
func (s *TracingKVStore) Iterator(start, end []byte) storetypes.Iterator {
	return s.KVStore.Iterator(start, end)
}

func (s *TracingKVStore) ReverseIterator(start, end []byte) storetypes.Iterator {
	return s.KVStore.ReverseIterator(start, end)
}

// CacheWrap delegates to the underlying store.
func (s *TracingKVStore) CacheWrap() storetypes.CacheWrap {
	return s.KVStore.CacheWrap()
}

// GetStoreType delegates to the underlying store.
func (s *TracingKVStore) GetStoreType() storetypes.StoreType {
	return s.KVStore.GetStoreType()
}
