package encoding

import "encoding/binary"

const separator = '/'

func BuildPrefixWithVersion(prefix string, version uint64) []byte {
	n := len(prefix)
	buf := make([]byte, n+8+1)
	copy(buf, prefix)
	binary.BigEndian.PutUint64(buf[n:], version)
	buf[n+8] = separator
	return buf
}
