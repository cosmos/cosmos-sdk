package branch

import (
	"fmt"
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
}

func (b WriterMap) GetReader(actor []byte) (store.Reader, error) {
	return b.GetWriter(actor)
}

func (b WriterMap) GetWriter(actor []byte) (store.Writer, error) {
	// Simplify and optimize state retrieval
	if actorState, ok := b.branchedWriterState[unsafeString(actor)]; ok {
		return actorState, nil
	} else if writerState, err := b.state.GetReader(actor); err != nil {
		return nil, err
	} else {
		actorState = b.branch(writerState)
		b.branchedWriterState[string(actor)] = actorState
		return actorState, nil
	}
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
