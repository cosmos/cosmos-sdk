package orm

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Indexable = &AutoUInt64Table{}

// AutoUInt64Table is the table type with an auto incrementing ID.
type AutoUInt64Table struct {
	*table
	seq Sequence
}

// NewAutoUInt64Table creates a new AutoUInt64Table.
func NewAutoUInt64Table(prefixData [2]byte, prefixSeq byte, model codec.ProtoMarshaler, cdc codec.Codec) (*AutoUInt64Table, error) {
	table, err := newTable(prefixData, model, cdc)
	if err != nil {
		return nil, err
	}
	return &AutoUInt64Table{
		table: table,
		seq:   NewSequence(prefixSeq),
	}, nil
}

// Create a new persistent object with an auto generated uint64 primary key. The
// key is returned.
//
// Create iterates through the registered callbacks that may add secondary index
// keys.
func (a AutoUInt64Table) Create(store sdk.KVStore, obj codec.ProtoMarshaler) (uint64, error) {
	autoIncID := a.seq.NextVal(store)
	err := a.table.Create(store, EncodeSequence(autoIncID), obj)
	if err != nil {
		return 0, err
	}
	return autoIncID, nil
}

// Update updates the given object under the rowID key. It expects the key to
// exists already and fails with an `ErrNotFound` otherwise. Any caller must
// therefore make sure that this contract is fulfilled. Parameters must not be
// nil.
//
// Update iterates through the registered callbacks that may add or remove
// secondary index keys.
func (a AutoUInt64Table) Update(store sdk.KVStore, rowID uint64, newValue codec.ProtoMarshaler) error {
	return a.table.Update(store, EncodeSequence(rowID), newValue)
}

// Delete removes the object under the rowID key. It expects the key to exists already
// and fails with a `ErrNotFound` otherwise. Any caller must therefore make sure that this contract
// is fulfilled.
//
// Delete iterates though the registered callbacks and removes secondary index keys by them.
func (a AutoUInt64Table) Delete(store sdk.KVStore, rowID uint64) error {
	return a.table.Delete(store, EncodeSequence(rowID))
}

// Has checks if a rowID exists.
func (a AutoUInt64Table) Has(store sdk.KVStore, rowID uint64) bool {
	return a.table.Has(store, EncodeSequence(rowID))
}

// GetOne load the object persisted for the given RowID into the dest parameter.
// If none exists `ErrNotFound` is returned instead. Parameters must not be nil.
func (a AutoUInt64Table) GetOne(store sdk.KVStore, rowID uint64, dest codec.ProtoMarshaler) (RowID, error) {
	rawRowID := EncodeSequence(rowID)
	if err := a.table.GetOne(store, rawRowID, dest); err != nil {
		return nil, err
	}
	return rawRowID, nil
}

// Sequence returns the sequence used by this table
func (a AutoUInt64Table) Sequence() Sequence {
	return a.seq
}
