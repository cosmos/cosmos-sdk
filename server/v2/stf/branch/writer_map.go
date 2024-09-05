package branch

import (
	"fmt"
	"sync"
	"unsafe"

	"cosmossdk.io/core/store"
)

func NewWriterMap(
	state store.ReaderMap,
	branch func(readonlyState store.Reader) store.Writer,
) store.WriterMap {
	return WriterMap{
		state:               state,
		branchedWriterState: make(map[string]store.Writer),
		branch:              branch,
	}
}

// WriterMap implements a branched version of the store.WriterMap.
// After the firs time the actor's branched Store is created, it is
// memoized in the WriterMap.
type WriterMap struct {
	state               store.ReaderMap
	branchedWriterState map[string]store.Writer
	branch              func(state store.Reader) store.Writer
	mu                  sync.RWMutex // mutex to protect branchedWriterState
}

func (b WriterMap) GetReader(actor []byte) (store.Reader, error) {
	return b.GetWriter(actor)
}

func (b WriterMap) GetWriter(actor []byte) (store.Writer, error) {
	actorKey := unsafeString(actor)

	// attempt to read the map with a read lock
	b.mu.RLock()
	actorState, ok := b.branchedWriterState[actorKey]
	b.mu.RUnlock()

	if ok {
		// if the actorState is found, return it
		return actorState, nil
	}

	// if not found, proceed with acquiring a write lock to update the map
	b.mu.Lock()
	defer b.mu.Unlock()

	// ensure that the actorState wasn't created by another goroutine while waiting for the write lock
	if actorState, ok = b.branchedWriterState[actorKey]; ok {
		return actorState, nil
	}

	// if still not found, create the actorState and update the map
	writerState, err := b.state.GetReader(actor)
	if err != nil {
		return nil, err
	}

	actorState = b.branch(writerState)
	b.branchedWriterState[actorKey] = actorState

	return actorState, nil
}

func (b WriterMap) ApplyStateChanges(stateChanges []store.StateChanges) error {
	for _, sc := range stateChanges {
		if err := b.applyStateChange(sc); err != nil {
			return fmt.Errorf("unable to apply state change for actor %X: %w", sc.Actor, err)
		}
	}
	return nil
}

// GetStateChanges returns the state changes for all actors in the WriterMap.
func (b WriterMap) GetStateChanges() ([]store.StateChanges, error) {
	sc := make([]store.StateChanges, 0, len(b.branchedWriterState))
	for acc, w := range b.branchedWriterState {
		accBytes := []byte(acc)
		kvChanges, err := w.ChangeSets()
		if err != nil {
			return nil, fmt.Errorf("unable to get actor writer changes %x: %w", accBytes, err)
		}
		sc = append(sc, store.StateChanges{
			Actor:        accBytes,
			StateChanges: kvChanges,
		})
	}

	return sc, nil
}

func (b WriterMap) applyStateChange(sc store.StateChanges) error {
	writableState, err := b.GetWriter(sc.Actor)
	if err != nil {
		return err
	}
	return writableState.ApplyChangeSets(sc.StateChanges)
}

func unsafeString(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }
