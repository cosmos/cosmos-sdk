package orm

// Max255DynamicLengthIndexKeyCodec works with up to 255 byte dynamic size RowIDs.
// They are encoded as `concat(searchableKey, rowID, len(rowID)[0])` and can be used
// with PrimaryKey or external Key tables for example.
type Max255DynamicLengthIndexKeyCodec struct{}

// BuildIndexKey builds the index key by appending searchableKey with rowID and length int.
// The RowID length must not be greater than 255.
func (Max255DynamicLengthIndexKeyCodec) BuildIndexKey(searchableKey []byte, rowID RowID) []byte {
	rowIDLen := len(rowID)
	switch {
	case rowIDLen == 0:
		panic("Empty RowID")
	case rowIDLen > 255:
		panic("RowID exceeds max size")
	}

	searchableKeyLen := len(searchableKey)
	res := make([]byte, searchableKeyLen+rowIDLen+1)
	copy(res, searchableKey)
	copy(res[searchableKeyLen:], rowID)
	res[searchableKeyLen+rowIDLen] = byte(rowIDLen)
	return res
}

// StripRowID returns the RowID from the combined persistentIndexKey. It is the reverse operation to BuildIndexKey
// but with the searchableKey and length int dropped.
func (Max255DynamicLengthIndexKeyCodec) StripRowID(persistentIndexKey []byte) RowID {
	n := len(persistentIndexKey)
	searchableKeyLen := persistentIndexKey[n-1]
	return persistentIndexKey[n-int(searchableKeyLen)-1 : n-1]
}

// FixLengthIndexKeyCodec expects the RowID to always have the same length with all entries.
// They are encoded as `concat(searchableKey, rowID)` and can be used
// with AutoUint64Tables and length EncodedSeqLength for example.
type FixLengthIndexKeyCodec struct {
	rowIDLength int
}

// FixLengthIndexKeys is a constructor for FixLengthIndexKeyCodec.
func FixLengthIndexKeys(rowIDLength int) *FixLengthIndexKeyCodec {
	return &FixLengthIndexKeyCodec{rowIDLength: rowIDLength}
}

// BuildIndexKey builds the index key by appending searchableKey with rowID.
// The RowID length must not be greater than what is defined by rowIDLength in construction.
func (c FixLengthIndexKeyCodec) BuildIndexKey(searchableKey []byte, rowID RowID) []byte {
	switch n := len(rowID); {
	case n == 0:
		panic("Empty RowID")
	case n > c.rowIDLength:
		panic("RowID exceeds max size")
	}
	n := len(searchableKey)
	res := make([]byte, n+c.rowIDLength)
	copy(res, searchableKey)
	copy(res[n:], rowID)
	return res
}

// StripRowID returns the RowID from the combined persistentIndexKey. It is the reverse operation to BuildIndexKey
// but with the searchableKey dropped.
func (c FixLengthIndexKeyCodec) StripRowID(persistentIndexKey []byte) RowID {
	n := len(persistentIndexKey)
	return persistentIndexKey[n-c.rowIDLength:]
}
