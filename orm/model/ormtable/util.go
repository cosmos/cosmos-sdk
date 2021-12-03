package ormtable

import "encoding/binary"

func AppendVarUInt32(prefix []byte, x uint32) []byte {
	prefixLen := len(prefix)
	res := make([]byte, prefixLen+binary.MaxVarintLen32)
	copy(res, prefix)
	n := binary.PutUvarint(res[prefixLen:], uint64(x))
	return res[:prefixLen+n]

}
