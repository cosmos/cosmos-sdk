package types

import (
	"io"

	"github.com/cosmos/cosmos-sdk/codec"
)

// WriteListener interface for streaming data out from a listenkv.Store
type WriteListener interface {
	// if value is nil then it was deleted
	// storeKey indicates the source KVStore, to facilitate using the the same WriteListener across separate KVStores
	// delete bool indicates if it was a delete; true: delete, false: set
	OnWrite(storeKey StoreKey, key []byte, value []byte, delete bool) error
}

// StoreKVPairWriteListener is used to configure listening to a KVStore by writing out length-prefixed
// protobuf encoded StoreKVPairs to an underlying io.Writer
type StoreKVPairWriteListener struct {
	writer     io.Writer
	marshaller codec.BinaryCodec
}

// NewStoreKVPairWriteListener wraps creates a StoreKVPairWriteListener with a provdied io.Writer and codec.BinaryCodec
func NewStoreKVPairWriteListener(w io.Writer, m codec.BinaryCodec) *StoreKVPairWriteListener {
	return &StoreKVPairWriteListener{
		writer:     w,
		marshaller: m,
	}
}

// OnWrite satisfies the WriteListener interface by writing length-prefixed protobuf encoded StoreKVPairs
func (wl *StoreKVPairWriteListener) OnWrite(storeKey StoreKey, key []byte, value []byte, delete bool) error {
	kvPair := new(StoreKVPair)
	kvPair.StoreKey = storeKey.Name()
	kvPair.Delete = delete
	kvPair.Key = key
	kvPair.Value = value
	by, err := wl.marshaller.MarshalLengthPrefixed(kvPair)
	if err != nil {
		return err
	}
	if _, err := wl.writer.Write(by); err != nil {
		return err
	}
	return nil
}
