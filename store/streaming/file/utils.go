package file

import (
	"encoding/binary"
	"fmt"
)

// SegmentBytes returns all of the protobuf messages contained in the byte array as an array of byte arrays
// The messages have their length prefix removed
// The first byte array will be the abci request, the last the abci response, and in between are the KVPairs
func SegmentBytes(bz []byte) ([][]byte, error) {
	var err error
	segments := make([][]byte, 0)
	for len(bz) > 0 {
		var segment []byte
		segment, bz, err = getHeadSegment(bz)
		if err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}
	return segments, nil
}

// Returns the bytes for the leading protobuf object in the byte array (removing the length prefix) and returns the remainder of the byte array
func getHeadSegment(bz []byte) ([]byte, []byte, error) {
	size, prefixSize := binary.Uvarint(bz)
	if prefixSize < 0 {
		return nil, nil, fmt.Errorf("invalid number of bytes read from length-prefixed encoding: %d", prefixSize)
	}
	if size > uint64(len(bz)-prefixSize) {
		return nil, nil, fmt.Errorf("not enough bytes to read; want: %v, got: %v", size, len(bz)-prefixSize)
	}
	return bz[prefixSize:(uint64(prefixSize) + size)], bz[uint64(prefixSize)+size:], nil
}
