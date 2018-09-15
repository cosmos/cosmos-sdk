package trace

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/store/utils"
)

var kvPairs = []types.KVPair{
	{Key: utils.KeyFmt(1), Value: utils.ValFmt(1)},
	{Key: utils.KeyFmt(2), Value: utils.ValFmt(2)},
	{Key: utils.KeyFmt(3), Value: utils.ValFmt(3)},
}

func newStore(w io.Writer) *Store {
	store := newEmptyStore(w)

	for _, kvPair := range kvPairs {
		store.Set(kvPair.Key, kvPair.Value)
	}

	return store
}

func newEmptyStore(w io.Writer) *Store {
	memDB := dbadapter.NewStore(dbm.NewMemDB())
	tc := types.TraceContext(map[string]interface{}{"blockHeight": 64})

	return NewStore(memDB, &types.Tracer{w, tc})
}

func TestStoreGet(t *testing.T) {
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

		store := newStore(&buf)
		buf.Reset()
		value := store.Get(tc.key)

		require.Equal(t, tc.expectedValue, value)
		require.Equal(t, tc.expectedOut, buf.String())
	}
}

func TestStoreSet(t *testing.T) {
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

		store := newEmptyStore(&buf)
		buf.Reset()
		store.Set(tc.key, tc.value)

		require.Equal(t, tc.expectedOut, buf.String())
	}
}

func TestStoreDelete(t *testing.T) {
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

		store := newStore(&buf)
		buf.Reset()
		store.Delete(tc.key)

		require.Equal(t, tc.expectedOut, buf.String())
	}
}

func TestStoreHas(t *testing.T) {
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

		store := newStore(&buf)
		buf.Reset()
		ok := store.Has(tc.key)

		require.Equal(t, tc.expected, ok)
	}
}

func TestTestStoreIterator(t *testing.T) {
	var buf bytes.Buffer

	store := newStore(&buf)
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
	require.Panics(t, iterator.Next)
	require.NotPanics(t, iterator.Close)
}

func TestTestStoreReverseIterator(t *testing.T) {
	var buf bytes.Buffer

	store := newStore(&buf)
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
	require.Panics(t, iterator.Next)
	require.NotPanics(t, iterator.Close)
}

func TestStoreGetStoreType(t *testing.T) {
	memDB := dbadapter.NewStore(dbm.NewMemDB())
	store := NewEmptyStore(nil)
	require.Equal(t, memDB.GetStoreType(), store.GetStoreType())
}

func TestStoreCacheWrap(t *testing.T) {
	store := newEmptyStore(nil)
	require.Panics(t, func() { store.CacheWrap() })
}
func TestStoreCacheWrapWithTrace(t *testing.T) {
	store := newEmptyStore(nil)
	require.Panics(t, func() { store.CacheWrapWithTrace(nil, nil) })
}
