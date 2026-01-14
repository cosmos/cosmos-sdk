package internal

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/tidwall/btree"
)

type TreeStoreOptions struct {
	ChangesetRolloverSize int
	CheckpointInterval    int
}

type TreeStore struct {
	opts TreeStoreOptions

	version    atomic.Uint32
	savedLayer atomic.Uint32

	currentWriter *ChangesetWriter

	root *NodePointer

	changesetsByVersion *btree.Map[uint32, *Changeset]
	changesetsLock      sync.RWMutex
	checkpointMgr       *checkpointMgr

	latestRoots     *btree.Map[uint32, *NodePointer]
	latestRootsLock sync.RWMutex
}

func NewTreeStore(dir string, opts TreeStoreOptions) (*TreeStore, error) {
	ts := &TreeStore{opts: opts}
	writer, err := NewChangesetWriter(dir, 1, 1, ts)
	if err != nil {
		return nil, err
	}
	ts.currentWriter = writer
	return ts, nil
}

func (ts *TreeStore) StagedVersion() uint32 {
	return ts.version.Load() + 1
}

func (ts *TreeStore) SaveRoot(newRoot *NodePointer) error {
	ts.root = newRoot
	ts.version.Add(1)
	return nil
}

func (ts *TreeStore) WriteWALUpdates(updates []KVUpdate, fsync bool) error {
	walWriter := ts.currentWriter.WALWriter()
	err := walWriter.WriteWALUpdates(updates)
	if err != nil {
		return err
	}

	err = walWriter.WriteWALCommit(uint64(ts.StagedVersion()))
	if err != nil {
		return err
	}

	if fsync {
		err = walWriter.Sync()
		if err != nil {
			return err
		}
	}

	// TODO if we are at rollover size, create new changeset writer
	if walWriter.Size() >= ts.opts.ChangesetRolloverSize {
		ts.checkpointMgr.reqChan <- checkpointReq{
			newWriter: ts.currentWriter,
			root:      ts.root,
		}
	}

	return nil
}

func (ts *TreeStore) ForceToDisk() error {
	//if ts.root == nil || ts.root.Mem.Load() == nil {
	//	return nil
	//}
	//layer, err := ts.currentWriter.SaveLayer(ts.root)
	//if err != nil {
	//	return err
	//}
	//ts.savedLayer.Store(layer)
	//err = ts.currentWriter.CreatedSharedReader()
	//if err != nil {
	//	return err
	//}
	//ts.root.Mem.Store(nil) // flush in-memory node
	return nil
}

func (ts *TreeStore) Latest() *NodePointer {
	return ts.root
}

func (ts *TreeStore) GetChangesetForLayer(layer uint32) (*ChangesetReader, Pin) {
	panic("not implemented")
}

type checkpointMgr struct {
	lastLayer         atomic.Uint32
	reqChan           chan checkpointReq
	changesetsByLayer *btree.Map[uint32, *Changeset]
}

func newCheckpointProc() *checkpointMgr {
	return &checkpointMgr{
		reqChan: make(chan checkpointReq),
	}
}

func (cp *checkpointMgr) start(doneChan chan error) {
	go func() {
		err := cp.proc()
		doneChan <- err
	}()
}

func (cp *checkpointMgr) proc() error {
	var curWriter *ChangesetWriter
	for req := range cp.reqChan {
		layer := cp.lastLayer.Load() + 1
		if req.newWriter != nil {
			if curWriter != nil {
				err := curWriter.Seal()
				if err != nil {
					return err
				}
			}
			curWriter = req.newWriter
			cp.changesetsByLayer.Set(layer, curWriter.Changeset())
		}
		if curWriter == nil {
			return fmt.Errorf("checkpointMgr: no current writer")
		}
		err := curWriter.SaveLayer(layer, req.root)
		if err != nil {
			return err
		}
		err = curWriter.Flush()
		if err != nil {
			return err
		}
		cp.lastLayer.Store(layer)
	}
}

type checkpointReq struct {
	newWriter *ChangesetWriter
	root      *NodePointer
}
