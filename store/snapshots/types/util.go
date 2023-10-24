package types

import (
	"encoding/binary"

	protoio "github.com/cosmos/gogoproto/io"
)

// WriteExtensionPayload writes an extension payload for current extension snapshotter.
func WriteExtensionPayload(protoWriter protoio.Writer, payload []byte) error {
	return protoWriter.WriteMsg(&SnapshotItem{
		Item: &SnapshotItem_ExtensionPayload{
			ExtensionPayload: &SnapshotExtensionPayload{
				Payload: payload,
			},
		},
	})
}

// Uint64ToBigEndian - marshals uint64 to a big endian byte slice so it can be sorted
func Uint64ToBigEndian(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

// BigEndianToUint64 returns an uint64 from big endian encoded bytes. If encoding
// is empty, zero is returned.
func BigEndianToUint64(bz []byte) uint64 {
	if len(bz) == 0 {
		return 0
	}

	return binary.BigEndian.Uint64(bz)
}
