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

// GetStateChanges returns the state changes for all actors in the WriterMap, including all direct
// ancesotors from which this WriterMap was derived.
// See WriterMap.recurseStateChanges for more details.
// Subject to possible renaming to ensure a developer can retrieve only changes in *this* branch
// context (not ancestors) if that is desired.
// see: https://github.com/cosmos/cosmos-sdk/pull/20412#discussion_r1618771230
func (b WriterMap) GetStateChanges() ([]store.StateChanges, error) {
	var (
		changes = make(map[string][]store.KVPair)
		sc      []store.StateChanges
	)
	if err := b.recurseStateChanges(changes); err != nil {
		return nil, err
	}

	for account, kvPairs := range changes {
		sc = append(sc, store.StateChanges{
			Actor:        []byte(account),
			StateChanges: kvPairs,
		})
	}
	return sc, nil
}

// recurseStateChanges will recursively collect state changes from the tree of
// WriterMap's and write them to the `changes` map.
func (b WriterMap) recurseStateChanges(changes map[string][]store.KVPair) error {
	// depth first
	if wr, ok := b.state.(WriterMap); ok {
		if err := wr.recurseStateChanges(changes); err != nil {
			return err
		}
	}
	for account, stateChange := range b.branchedWriterState {
		kvChanges, err := stateChange.ChangeSets()
		if err != nil {
			return err
		}
		changes[account] = append(changes[account], kvChanges...)
	}
	return nil
}

func (b WriterMap) applyStateChange(sc store.StateChanges) error {
	writableState, err := b.GetWriter(sc.Actor)
	if err != nil {
		return err
	}
	return writableState.ApplyChangeSets(sc.StateChanges)
}

func unsafeString(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }
