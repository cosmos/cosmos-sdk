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
	// Simplify and optimize state retrieval
	actorKey := unsafeString(actor)

	// acquire a lock to ensure thread-safe access to the branchedWriterState map
	b.mu.Lock()
	defer b.mu.Unlock()

	// check if the writer for the given actor already exists in the map
	actorState, ok := b.branchedWriterState[actorKey]
	if ok {
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
