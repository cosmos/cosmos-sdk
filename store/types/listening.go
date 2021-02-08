package types

import (
	"io"

	"github.com/cosmos/cosmos-sdk/codec"
)

// WriteListener interface for streaming data out from a listenkv.Store
type WriteListener interface {
	// if value is nil then it was deleted
	// storeKey indicates the source KVStore, to facilitate using the the same WriteListener across separate KVStores
	// set bool indicates if it was a set; true: set, false: delete
	OnWrite(storeKey StoreKey, set bool, key []byte, value []byte)
}

// StoreKVPairWriteListener is used to configure listening to a KVStore by writing out length-prefixed
// protobuf encoded StoreKVPairs to an underlying io.Writer
type StoreKVPairWriteListener struct {
	writer     io.Writer
	marshaller codec.BinaryMarshaler
}

// NewStoreKVPairWriteListener wraps creates a StoreKVPairWriteListener with a provdied io.Writer and codec.BinaryMarshaler
func NewStoreKVPairWriteListener(w io.Writer, m codec.BinaryMarshaler) *StoreKVPairWriteListener {
	return &StoreKVPairWriteListener{
		writer:     w,
		marshaller: m,
	}
}

// OnWrite satisfies the WriteListener interface by writing length-prefixed protobuf encoded StoreKVPairs
func (wl *StoreKVPairWriteListener) OnWrite(storeKey StoreKey, set bool, key []byte, value []byte) {
	kvPair := new(StoreKVPair)
	kvPair.StoreKey = storeKey.Name()
	kvPair.Set = set
	kvPair.Key = key
	kvPair.Value = value
	if by, err := wl.marshaller.MarshalBinaryLengthPrefixed(kvPair); err == nil {
		wl.writer.Write(by)
	}
}
