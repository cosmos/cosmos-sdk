package ormtable

import (
	"cosmossdk.io/orm/types/kv"
)

type batchIndexCommitmentWriter struct {
	Backend
	commitmentWriter *batchStoreWriter
	indexWriter      *batchStoreWriter
}

func newBatchIndexCommitmentWriter(store Backend) *batchIndexCommitmentWriter {
	return &batchIndexCommitmentWriter{
		Backend: store,
		commitmentWriter: &batchStoreWriter{
			ReadonlyStore: store.CommitmentStoreReader(),
			curBuf:        make([]*batchWriterEntry, 0, capacity),
		},
		indexWriter: &batchStoreWriter{
			ReadonlyStore: store.IndexStoreReader(),
			curBuf:        make([]*batchWriterEntry, 0, capacity),
		},
	}
}

func (w *batchIndexCommitmentWriter) CommitmentStore() kv.Store {
	return w.commitmentWriter
}

func (w *batchIndexCommitmentWriter) IndexStore() kv.Store {
	return w.indexWriter
}

// Write flushes any pending writes.
func (w *batchIndexCommitmentWriter) Write() error {
	err := flushWrites(w.Backend.CommitmentStore(), w.commitmentWriter)
	if err != nil {
		return err
	}

	err = flushWrites(w.Backend.IndexStore(), w.indexWriter)
	if err != nil {
		return err
	}

	// clear writes
	w.Close()

	return err
}

func flushWrites(store kv.Store, writer *batchStoreWriter) error {
	for _, buf := range writer.prevBufs {
		err := flushBuf(store, buf)
		if err != nil {
			return err
		}
	}
	return flushBuf(store, writer.curBuf)
}

func flushBuf(store kv.Store, writes []*batchWriterEntry) error {
	for _, write := range writes {
		switch {
		case write.hookCall != nil:
			write.hookCall()
		case !write.delete:
			err := store.Set(write.key, write.value)
			if err != nil {
				return err
			}
		default:
			err := store.Delete(write.key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Close discards any pending writes and should generally be called using
// a defer statement.
func (w *batchIndexCommitmentWriter) Close() {
	w.commitmentWriter.prevBufs = nil
	w.commitmentWriter.curBuf = nil
	w.indexWriter.prevBufs = nil
	w.indexWriter.curBuf = nil
}

type batchWriterEntry struct {
	key, value []byte
	delete     bool
	hookCall   func()
}

type batchStoreWriter struct {
	kv.ReadonlyStore
	prevBufs [][]*batchWriterEntry
	curBuf   []*batchWriterEntry
}

const capacity = 16

func (b *batchStoreWriter) Set(key, value []byte) error {
	b.append(&batchWriterEntry{key: key, value: value})
	return nil
}

func (b *batchStoreWriter) Delete(key []byte) error {
	b.append(&batchWriterEntry{key: key, delete: true})
	return nil
}

func (w *batchIndexCommitmentWriter) enqueueHook(f func()) {
	w.indexWriter.append(&batchWriterEntry{hookCall: f})
}

func (b *batchStoreWriter) append(entry *batchWriterEntry) {
	if len(b.curBuf) == capacity {
		b.prevBufs = append(b.prevBufs, b.curBuf)
		b.curBuf = make([]*batchWriterEntry, 0, capacity)
	}

	b.curBuf = append(b.curBuf, entry)
}

var _ Backend = &batchIndexCommitmentWriter{}
