package trace_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/kv/mem"
	"cosmossdk.io/store/v2/kv/trace"
)

const storeKey = "storeKey"

var kvPairs = store.KVPairs{
	{Key: []byte(fmt.Sprintf("key%0.8d", 1)), Value: []byte(fmt.Sprintf("value%0.8d", 1))},
	{Key: []byte(fmt.Sprintf("key%0.8d", 2)), Value: []byte(fmt.Sprintf("value%0.8d", 2))},
	{Key: []byte(fmt.Sprintf("key%0.8d", 3)), Value: []byte(fmt.Sprintf("value%0.8d", 3))},
}

func newTraceKVStore(w io.Writer) store.KVStore {
	store := newEmptyTraceKVStore(w)

	for _, kvPair := range kvPairs {
		store.Set(kvPair.Key, kvPair.Value)
	}

	return store
}

func newEmptyTraceKVStore(w io.Writer) store.KVStore {
	memKVStore := mem.New(storeKey)
	tc := store.TraceContext(map[string]any{"blockHeight": 64})

	return trace.New(memKVStore, w, tc)
}

func TestTraceKVStoreGet(t *testing.T) {
	testCases := []struct {
		key           []byte
		expectedValue []byte
		expectedOut   string
	}{
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
			key:         kvPairs[0].Key,
			value:       kvPairs[0].Value,
			expectedOut: "{\"operation\":\"write\",\"key\":\"a2V5MDAwMDAwMDE=\",\"value\":\"dmFsdWUwMDAwMDAwMQ==\",\"metadata\":{\"blockHeight\":64}}\n",
		},
		{
			key:         kvPairs[1].Key,
			value:       kvPairs[1].Value,
			expectedOut: "{\"operation\":\"write\",\"key\":\"a2V5MDAwMDAwMDI=\",\"value\":\"dmFsdWUwMDAwMDAwMg==\",\"metadata\":{\"blockHeight\":64}}\n",
		},
		{
			key:         kvPairs[2].Key,
			value:       kvPairs[2].Value,
			expectedOut: "{\"operation\":\"write\",\"key\":\"a2V5MDAwMDAwMDM=\",\"value\":\"dmFsdWUwMDAwMDAwMw==\",\"metadata\":{\"blockHeight\":64}}\n",
		},
	}

	for _, tc := range testCases {
		var buf bytes.Buffer

		store := newEmptyTraceKVStore(&buf)
		buf.Reset()
		store.Set(tc.key, tc.value)

		require.Equal(t, tc.expectedOut, buf.String())
	}

	var buf bytes.Buffer
	store := newEmptyTraceKVStore(&buf)
	require.Panics(t, func() { store.Set([]byte(""), []byte("value")) }, "setting an empty key should panic")
	require.Panics(t, func() { store.Set(nil, []byte("value")) }, "setting a nil key should panic")
}

func TestTraceKVStoreDelete(t *testing.T) {
	testCases := []struct {
		key         []byte
		expectedOut string
	}{
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
	require.Equal(t, []byte(nil), s)
	require.Equal(t, []byte(nil), e)

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
	require.False(t, iterator.Next())
}

func TestTestTraceKVStoreReverseIterator(t *testing.T) {
	var buf bytes.Buffer

	store := newTraceKVStore(&buf)
	iterator := store.ReverseIterator(nil, nil)

	s, e := iterator.Domain()
	require.Equal(t, []byte(nil), s)
	require.Equal(t, []byte(nil), e)

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
	require.False(t, iterator.Next())
}

func TestTraceKVStoreGetStoreType(t *testing.T) {
	traceKVStore := newEmptyTraceKVStore(nil)
	require.Equal(t, store.StoreTypeTrace, traceKVStore.GetStoreType())
}
