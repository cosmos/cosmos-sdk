package types

import (
	protoio "github.com/cosmos/gogoproto/io"
)

// WriteExtensionItem writes an item payload for current extention snapshotter.
func WriteExtensionItem(protoWriter protoio.Writer, item []byte) error {
	return protoWriter.WriteMsg(&SnapshotItem{
		Item: &SnapshotItem_ExtensionPayload{
			ExtensionPayload: &SnapshotExtensionPayload{
				Payload: item,
			},
		},
	})
}
