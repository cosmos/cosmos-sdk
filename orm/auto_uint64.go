package orm

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Indexable = &AutoUInt64TableBuilder{}

// NewAutoUInt64TableBuilder creates a builder to setup a AutoUInt64Table object.
func NewAutoUInt64TableBuilder(prefixData byte, prefixSeq byte, storeKey sdk.StoreKey, model codec.ProtoMarshaler, cdc codec.Codec) *AutoUInt64TableBuilder {
	if prefixData == prefixSeq {
		panic("prefixData and prefixSeq must be unique")
	}

	uInt64KeyCodec := FixLengthIndexKeys(EncodedSeqLength)
	return &AutoUInt64TableBuilder{
		TableBuilder: NewTableBuilder(prefixData, storeKey, model, uInt64KeyCodec, cdc),
		seq:          NewSequence(storeKey, prefixSeq),
	}
}

type AutoUInt64TableBuilder struct {
	*TableBuilder
	seq Sequence
}

// Build create the AutoUInt64Table object.
func (a AutoUInt64TableBuilder) Build() AutoUInt64Table {
	return AutoUInt64Table{
		table: a.TableBuilder.Build(),
		seq:   a.seq,
	}
}

var _ SequenceExportable = &AutoUInt64Table{}
var _ TableExportable = &AutoUInt64Table{}

// AutoUInt64Table is the table type which an auto incrementing ID.
type AutoUInt64Table struct {
	table Table
	seq   Sequence
}

// Create a new persistent object with an auto generated uint64 primary key. They key is returned.
// Create iterates though the registered callbacks and may add secondary index keys by them.
func (a AutoUInt64Table) Create(ctx HasKVStore, obj codec.ProtoMarshaler) (uint64, error) {
	autoIncID := a.seq.NextVal(ctx)
	err := a.table.Create(ctx, EncodeSequence(autoIncID), obj)
	if err != nil {
		return 0, err
	}
	return autoIncID, nil
}

// Save updates the given object under the rowID key. It expects the key to exists already
// and fails with an `ErrNotFound` otherwise. Any caller must therefore make sure that this contract
// is fulfilled. Parameters must not be nil.
//
// Save iterates though the registered callbacks and may add or remove secondary index keys by them.
func (a AutoUInt64Table) Save(ctx HasKVStore, rowID uint64, newValue codec.ProtoMarshaler) error {
	return a.table.Save(ctx, EncodeSequence(rowID), newValue)
}

// Delete removes the object under the rowID key. It expects the key to exists already
// and fails with a `ErrNotFound` otherwise. Any caller must therefore make sure that this contract
// is fulfilled.
//
// Delete iterates though the registered callbacks and removes secondary index keys by them.
func (a AutoUInt64Table) Delete(ctx HasKVStore, rowID uint64) error {
	return a.table.Delete(ctx, EncodeSequence(rowID))
}

// Has checks if a rowID exists.
func (a AutoUInt64Table) Has(ctx HasKVStore, rowID uint64) bool {
	return a.table.Has(ctx, EncodeSequence(rowID))
}

// GetOne load the object persisted for the given RowID into the dest parameter.
// If none exists `ErrNotFound` is returned instead. Parameters must not be nil.
func (a AutoUInt64Table) GetOne(ctx HasKVStore, rowID uint64, dest codec.ProtoMarshaler) (RowID, error) {
	rawRowID := EncodeSequence(rowID)
	if err := a.table.GetOne(ctx, rawRowID, dest); err != nil {
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
func (a AutoUInt64Table) PrefixScan(ctx HasKVStore, start, end uint64) (Iterator, error) {
	return a.table.PrefixScan(ctx, EncodeSequence(start), EncodeSequence(end))
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
func (a AutoUInt64Table) ReversePrefixScan(ctx HasKVStore, start uint64, end uint64) (Iterator, error) {
	return a.table.ReversePrefixScan(ctx, EncodeSequence(start), EncodeSequence(end))
}

// Sequence returns the sequence used by this table
func (a AutoUInt64Table) Sequence() Sequence {
	return a.seq
}

// Table satisfies the TableExportable interface and must not be used otherwise.
func (a AutoUInt64Table) Table() Table {
	return a.table
}
