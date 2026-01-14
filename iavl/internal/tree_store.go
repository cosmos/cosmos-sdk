package internal

import "sync/atomic"

type TreeStore struct {
	version    uint32
	savedLayer atomic.Uint32
	writer     *ChangesetWriter
	root       *NodePointer
}

func NewTreeStore(dir string) (*TreeStore, error) {
	ts := &TreeStore{}
	writer, err := NewChangesetWriter(dir, 1, ts)
	if err != nil {
		return nil, err
	}
	err = writer.kvlog.WriteStartWAL(uint64(ts.StagedVersion()))
	if err != nil {
		return nil, err
	}
	ts.writer = writer
	return ts, nil
}

func (ts *TreeStore) StagedVersion() uint32 {
	return ts.version + 1
}

func (ts *TreeStore) SaveRoot(newRoot *NodePointer) error {
	ts.root = newRoot
	ts.version++
	return nil
}

func (ts *TreeStore) WriteWALUpdates(updates []KVUpdate, fsync bool) error {
	err := ts.writer.kvlog.WriteWALUpdates(updates)
	if err != nil {
		return err
	}

	err = ts.writer.kvlog.WriteWALCommit(uint64(ts.StagedVersion()))
	if err != nil {
		return err
	}

	if fsync {
		err = ts.writer.kvlog.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ts *TreeStore) ForceToDisk() error {
	if ts.root == nil || ts.root.Mem.Load() == nil {
		return nil
	}
	layer, err := ts.writer.SaveLayer(ts.root)
	if err != nil {
		return err
	}
	ts.savedLayer.Store(layer)
	err = ts.writer.CreatedSharedReader()
	if err != nil {
		return err
	}
	ts.root.Mem.Store(nil) // flush in-memory node
	return nil
}

func (ts *TreeStore) Latest() *NodePointer {
	return ts.root
}

func (ts *TreeStore) GetChangesetForLayer(layer uint32) (*ChangesetReader, Pin) {
	panic("not implemented")
}
