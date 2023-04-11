package v5

import "encoding/binary"

var HistoricalInfoKey = []byte{0x50} // prefix for the historical info

// GetHistoricalInfoKey returns a key prefix for indexing HistoricalInfo objects.
func GetHistoricalInfoKey(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, uint64(height))
	return append(HistoricalInfoKey, heightBytes...)
}
