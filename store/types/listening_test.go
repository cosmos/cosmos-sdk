package types

import (
	"bytes"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"
)

func TestNewStoreKVPairWriteListener(t *testing.T) {
	testWriter := new(bytes.Buffer)
	interfaceRegistry := types.NewInterfaceRegistry()
	testMarshaller := codec.NewProtoCodec(interfaceRegistry)

	wl := NewStoreKVPairWriteListener(testWriter, testMarshaller)

	require.IsType(t, &StoreKVPairWriteListener{}, wl)
	require.Equal(t, testWriter, wl.writer)
	require.Equal(t, testMarshaller, wl.marshaller)
}

func TestOnWrite(t *testing.T) {
	testWriter := new(bytes.Buffer)
	interfaceRegistry := types.NewInterfaceRegistry()
	testMarshaller := codec.NewProtoCodec(interfaceRegistry)

	wl := NewStoreKVPairWriteListener(testWriter, testMarshaller)

	testStoreKey := NewKVStoreKey("test_key")
	testKey := []byte("testing123")
	testValue := []byte("testing321")
	err := wl.OnWrite(testStoreKey, true, testKey, testValue)
	require.Nil(t, err)

	outputBytes := testWriter.Bytes()
	outputKVPair := new(StoreKVPair)
	expectedOutputKVPair := &StoreKVPair{
		Key:      testKey,
		Value:    testValue,
		StoreKey: testStoreKey.Name(),
		Set:      true,
	}
	testMarshaller.UnmarshalBinaryLengthPrefixed(outputBytes, outputKVPair)
	require.EqualValues(t, expectedOutputKVPair, outputKVPair)
}
