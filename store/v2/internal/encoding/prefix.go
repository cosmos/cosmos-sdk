package encoding

import "encoding/binary"

const separator = '/'

// BuildPrefixWithVersion returns a byte slice with the given prefix and BigEndian encoded version.
// It is mainly used to represent the removed store key at the metadata store.
func BuildPrefixWithVersion(prefix string, version uint64) []byte {
	n := len(prefix)
	buf := make([]byte, n+8+1)
	copy(buf, prefix)
	binary.BigEndian.PutUint64(buf[n:], version)
	buf[n+8] = separator
	return buf
}
