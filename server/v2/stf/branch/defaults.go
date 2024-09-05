package branch

import "cosmossdk.io/core/store"

func DefaultNewWriterMap(r store.ReaderMap) store.WriterMap {
	return NewWriterMap(r, func(readonlyState store.Reader) store.Writer {
		return NewStore(readonlyState)
	})
}
