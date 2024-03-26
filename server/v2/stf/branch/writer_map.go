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
	actorState, ok := b.branchedWriterState[unsafeString(actor)]
	if ok {
		return actorState, nil
	}

	writerState, err := b.state.GetReader(actor)
	if err != nil {
		return nil, err
	}

	actorState = b.branch(writerState)
	b.branchedWriterState[string(actor)] = actorState

	return actorState, nil
}

func (b WriterMap) ApplyStateChanges(stateChanges []store.StateChanges) error {
	for _, sc := range stateChanges {
		err := b.applyStateChange(sc)
		if err != nil {
			return fmt.Errorf("unable to apply state change for actor %X: %w", sc.Actor, err)
		}
	}
	return nil
}

func (b WriterMap) GetStateChanges() ([]store.StateChanges, error) {
	sc := make([]store.StateChanges, len(b.branchedWriterState))
	for account, stateChange := range b.branchedWriterState {
		kvChanges, err := stateChange.ChangeSets()
		if err != nil {
			return nil, err
		}
		sc = append(sc, store.StateChanges{
			Actor:        []byte(account),
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
