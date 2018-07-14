package store

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
)

var kvPairs = []KVPair{
	{Key: keyFmt(1), Value: valFmt(1)},
	{Key: keyFmt(2), Value: valFmt(2)},
	{Key: keyFmt(3), Value: valFmt(3)},
}

func newTraceKVStore(w io.Writer) *TraceKVStore {
	store := newEmptyTraceKVStore(w)

	for _, kvPair := range kvPairs {
		store.Set(kvPair.Key, kvPair.Value)
	}

	return store
}

func newEmptyTraceKVStore(w io.Writer) *TraceKVStore {
	memDB := dbStoreAdapter{dbm.NewMemDB()}
	tc := TraceContext(map[string]interface{}{"blockHeight": 64})

	return NewTraceKVStore(memDB, w, tc)
}

func TestTraceKVStoreGet(t *testing.T) {
	testCases := []struct {
		key           []byte
		expectedValue []byte
		expectedOut   string
	}{
		{
			key:           []byte{},
			expectedValue: nil,
			expectedOut:   "{\"operation\":\"read\",\"key\":\"\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n",
		},
		{
			key:           kvPairs[0].Key,
			expectedValue: kvPairs[0].Value,
			expectedOut:   "{\"operation\":\"read\",\"key\":\"a2V5MDAwMDAwMDE=\",\"value\":\"dmFsdWUwMDAwMDAwMQ==\",\"metadata\":{\"blockHeight\":64}}\n",
		},
		{
			key:           []byte("does-not-exist"),
			expectedValue: nil,
			expectedOut:   "{\"operation\":\"read\",\"key\":\"ZG9lcy1ub3QtZXhpc3Q=\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n",
		},
	}

	for _, tc := range testCases {
		var buf bytes.Buffer

		store := newTraceKVStore(&buf)
		buf.Reset()
		value := store.Get(tc.key)

		require.Equal(t, tc.expectedValue, value)
		require.Equal(t, tc.expectedOut, buf.String())
	}
}

func TestTraceKVStoreSet(t *testing.T) {
	testCases := []struct {
		key         []byte
		value       []byte
		expectedOut string
	}{
		{
			key:         []byte{},
			value:       nil,
			expectedOut: "{\"operation\":\"write\",\"key\":\"\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n",
		},
		{
			key:         kvPairs[0].Key,
			value:       kvPairs[0].Value,
			expectedOut: "{\"operation\":\"write\",\"key\":\"a2V5MDAwMDAwMDE=\",\"value\":\"dmFsdWUwMDAwMDAwMQ==\",\"metadata\":{\"blockHeight\":64}}\n",
		},
	}

	for _, tc := range testCases {
		var buf bytes.Buffer

		store := newEmptyTraceKVStore(&buf)
		buf.Reset()
		store.Set(tc.key, tc.value)

		require.Equal(t, tc.expectedOut, buf.String())
	}
}

func TestTraceKVStoreDelete(t *testing.T) {
	testCases := []struct {
		key         []byte
		expectedOut string
	}{
		{
			key:         []byte{},
			expectedOut: "{\"operation\":\"delete\",\"key\":\"\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n",
		},
		{
			key:         kvPairs[0].Key,
			expectedOut: "{\"operation\":\"delete\",\"key\":\"a2V5MDAwMDAwMDE=\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n",
		},
	}

	for _, tc := range testCases {
		var buf bytes.Buffer

		store := newTraceKVStore(&buf)
		buf.Reset()
		store.Delete(tc.key)

		require.Equal(t, tc.expectedOut, buf.String())
	}
}

func TestTraceKVStoreHas(t *testing.T) {
	testCases := []struct {
		key      []byte
		expected bool
	}{
		{
			key:      []byte{},
			expected: false,
		},
		{
			key:      kvPairs[0].Key,
			expected: true,
		},
	}

	for _, tc := range testCases {
		var buf bytes.Buffer

		store := newTraceKVStore(&buf)
		buf.Reset()
		ok := store.Has(tc.key)

		require.Equal(t, tc.expected, ok)
	}
}

func TestTestTraceKVStoreIterator(t *testing.T) {
	var buf bytes.Buffer

	store := newTraceKVStore(&buf)
	iterator := store.Iterator(nil, nil)

	s, e := iterator.Domain()
	require.Equal(t, []uint8([]byte(nil)), s)
	require.Equal(t, []uint8([]byte(nil)), e)

	testCases := []struct {
		expectedKey      []byte
		expectedValue    []byte
		expectedKeyOut   string
		expectedvalueOut string
	}{
		{
			expectedKey:      kvPairs[0].Key,
			expectedValue:    kvPairs[0].Value,
			expectedKeyOut:   "{\"operation\":\"iterKey\",\"key\":\"a2V5MDAwMDAwMDE=\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n",
			expectedvalueOut: "{\"operation\":\"iterValue\",\"key\":\"\",\"value\":\"dmFsdWUwMDAwMDAwMQ==\",\"metadata\":{\"blockHeight\":64}}\n",
		},
		{
			expectedKey:      kvPairs[1].Key,
			expectedValue:    kvPairs[1].Value,
			expectedKeyOut:   "{\"operation\":\"iterKey\",\"key\":\"a2V5MDAwMDAwMDI=\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n",
			expectedvalueOut: "{\"operation\":\"iterValue\",\"key\":\"\",\"value\":\"dmFsdWUwMDAwMDAwMg==\",\"metadata\":{\"blockHeight\":64}}\n",
		},
		{
			expectedKey:      kvPairs[2].Key,
			expectedValue:    kvPairs[2].Value,
			expectedKeyOut:   "{\"operation\":\"iterKey\",\"key\":\"a2V5MDAwMDAwMDM=\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n",
			expectedvalueOut: "{\"operation\":\"iterValue\",\"key\":\"\",\"value\":\"dmFsdWUwMDAwMDAwMw==\",\"metadata\":{\"blockHeight\":64}}\n",
		},
	}

	for _, tc := range testCases {
		buf.Reset()
		ka := iterator.Key()
		require.Equal(t, tc.expectedKeyOut, buf.String())

		buf.Reset()
		va := iterator.Value()
		require.Equal(t, tc.expectedvalueOut, buf.String())

		require.Equal(t, tc.expectedKey, ka)
		require.Equal(t, tc.expectedValue, va)

		iterator.Next()
	}

	require.False(t, iterator.Valid())
	require.Panics(t, iterator.Next)
	require.NotPanics(t, iterator.Close)
}

func TestTestTraceKVStoreReverseIterator(t *testing.T) {
	var buf bytes.Buffer

	store := newTraceKVStore(&buf)
	iterator := store.ReverseIterator(nil, nil)

	s, e := iterator.Domain()
	require.Equal(t, []uint8([]byte(nil)), s)
	require.Equal(t, []uint8([]byte(nil)), e)

	testCases := []struct {
		expectedKey      []byte
		expectedValue    []byte
		expectedKeyOut   string
		expectedvalueOut string
	}{
		{
			expectedKey:      kvPairs[2].Key,
			expectedValue:    kvPairs[2].Value,
			expectedKeyOut:   "{\"operation\":\"iterKey\",\"key\":\"a2V5MDAwMDAwMDM=\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n",
			expectedvalueOut: "{\"operation\":\"iterValue\",\"key\":\"\",\"value\":\"dmFsdWUwMDAwMDAwMw==\",\"metadata\":{\"blockHeight\":64}}\n",
		},
		{
			expectedKey:      kvPairs[1].Key,
			expectedValue:    kvPairs[1].Value,
			expectedKeyOut:   "{\"operation\":\"iterKey\",\"key\":\"a2V5MDAwMDAwMDI=\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n",
			expectedvalueOut: "{\"operation\":\"iterValue\",\"key\":\"\",\"value\":\"dmFsdWUwMDAwMDAwMg==\",\"metadata\":{\"blockHeight\":64}}\n",
		},
		{
			expectedKey:      kvPairs[0].Key,
			expectedValue:    kvPairs[0].Value,
			expectedKeyOut:   "{\"operation\":\"iterKey\",\"key\":\"a2V5MDAwMDAwMDE=\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n",
			expectedvalueOut: "{\"operation\":\"iterValue\",\"key\":\"\",\"value\":\"dmFsdWUwMDAwMDAwMQ==\",\"metadata\":{\"blockHeight\":64}}\n",
		},
	}

	for _, tc := range testCases {
		buf.Reset()
		ka := iterator.Key()
		require.Equal(t, tc.expectedKeyOut, buf.String())

		buf.Reset()
		va := iterator.Value()
		require.Equal(t, tc.expectedvalueOut, buf.String())

		require.Equal(t, tc.expectedKey, ka)
		require.Equal(t, tc.expectedValue, va)

		iterator.Next()
	}

	require.False(t, iterator.Valid())
	require.Panics(t, iterator.Next)
	require.NotPanics(t, iterator.Close)
}

func TestTraceKVStorePrefix(t *testing.T) {
	store := newEmptyTraceKVStore(nil)
	pStore := store.Prefix([]byte("trace_prefix"))
	require.IsType(t, prefixStore{}, pStore)
}

func TestTraceKVStoreGetStoreType(t *testing.T) {
	memDB := dbStoreAdapter{dbm.NewMemDB()}
	store := newEmptyTraceKVStore(nil)
	require.Equal(t, memDB.GetStoreType(), store.GetStoreType())
}

func TestTraceKVStoreCacheWrap(t *testing.T) {
	store := newEmptyTraceKVStore(nil)
	require.Panics(t, func() { store.CacheWrap() })
}
func TestTraceKVStoreCacheWrapWithTrace(t *testing.T) {
	store := newEmptyTraceKVStore(nil)
	require.Panics(t, func() { store.CacheWrapWithTrace(nil, nil) })
}
