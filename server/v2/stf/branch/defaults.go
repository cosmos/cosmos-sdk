package branch

import "cosmossdk.io/server/v2/core/store"

func DefaultNewWriterMap(r store.ReaderMap) store.WriterMap {
	return NewWritersMap(r, func(readonlyState store.Reader) store.Writer {
		return NewStore(readonlyState)
	})
}
