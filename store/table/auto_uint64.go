package table

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Indexable = &AutoUInt64TableBuilder{}

// NewAutoUInt64TableBuilder creates a builder to setup a AutoUInt64Table object.
func NewAutoUInt64TableBuilder(prefixData byte, prefixSeq byte, model codec.ProtoMarshaler, cdc codec.Codec) *AutoUInt64TableBuilder {
	if prefixData == prefixSeq {
		panic("prefixData and prefixSeq must be unique")
	}

	uInt64KeyCodec := FixLengthIndexKeys(EncodedSeqLength)
	return &AutoUInt64TableBuilder{
		Builder: NewTableBuilder(prefixData, model, uInt64KeyCodec, cdc),
		seq:     NewSequence(prefixSeq),
	}
}

type AutoUInt64TableBuilder struct {
	*Builder
	seq Sequence
}

// Build create the AutoUInt64Table object.
func (a AutoUInt64TableBuilder) Build() AutoUInt64Table {
	return AutoUInt64Table{
		table: a.Builder.Build(),
		seq:   a.seq,
	}
}

// AutoUInt64Table is the table type which an auto incrementing ID.
type AutoUInt64Table struct {
	table Table
	seq   Sequence
}

// Create a new persistent object with an auto generated uint64 primary key. They key is returned.
// Create iterates though the registered callbacks and may add secondary index keys by them.
func (a AutoUInt64Table) Create(store sdk.KVStore, obj codec.ProtoMarshaler) (uint64, error) {
	autoIncID := a.seq.NextVal(store)
	err := a.table.Create(store, EncodeSequence(autoIncID), obj)
	if err != nil {
		return 0, err
	}
	return autoIncID, nil
}

// Update updates the given object under the rowID key. It expects the key to exists already
// and fails with an `ErrNotFound` otherwise. Any caller must therefore make sure that this contract
// is fulfilled. Parameters must not be nil.
//
// Update iterates though the registered callbacks and may add or remove secondary index keys by them.
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

// PrefixScan returns an Iterator over a domain of keys in ascending order. End is exclusive.
// Start is an MultiKeyIndex key or prefix. It must be less than end, or the Iterator is invalid and error is returned.
// Iterator must be closed by caller.
// To iterate over entire domain, use PrefixScan(nil, nil)
//
// WARNING: The use of a PrefixScan can be very expensive in terms of Gas. Please make sure you do not expose
// this as an endpoint to the public without further limits.
// Example:
//			it, err := idx.PrefixScan(ctx, start, end)
//			if err !=nil {
//				return err
//			}
//			const defaultLimit = 20
//			it = LimitIterator(it, defaultLimit)
//
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
func (a AutoUInt64Table) PrefixScan(store sdk.KVStore, start, end uint64) (Iterator, error) {
	return a.table.PrefixScan(store, EncodeSequence(start), EncodeSequence(end))
}

// ReversePrefixScan returns an Iterator over a domain of keys in descending order. End is exclusive.
// Start is an MultiKeyIndex key or prefix. It must be less than end, or the Iterator is invalid  and error is returned.
// Iterator must be closed by caller.
// To iterate over entire domain, use PrefixScan(nil, nil)
//
// WARNING: The use of a ReversePrefixScan can be very expensive in terms of Gas. Please make sure you do not expose
// this as an endpoint to the public without further limits. See `LimitIterator`
//
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
func (a AutoUInt64Table) ReversePrefixScan(store sdk.KVStore, start uint64, end uint64) (Iterator, error) {
	return a.table.ReversePrefixScan(store, EncodeSequence(start), EncodeSequence(end))
}
