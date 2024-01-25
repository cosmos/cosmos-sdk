package branch

import (
	"fmt"
	"unsafe"

	"cosmossdk.io/server/v2/core/store"
)

func NewWritersMap(
	state store.ReaderMap,
	branch func(readonlyState store.Reader) store.Writer) store.WriterMap {

	return WritersMap{
		state:                 state,
		branchedAccountsState: make(map[string]store.Writer),
		branch:                branch,
	}
}

type WritersMap struct {
	state                 store.ReaderMap
	branchedAccountsState map[string]store.Writer
	branch                func(state store.Reader) store.Writer
}

func (b WritersMap) GetReader(actor []byte) (store.Reader, error) {
	return b.GetWriter(actor)
}

func (b WritersMap) GetWriter(actor []byte) (store.Writer, error) {
	actorState, ok := b.branchedAccountsState[unsafeString(actor)]
	if ok {
		return actorState, nil
	}

	accountState, err := b.state.GetReader(actor)
	if err != nil {
		return nil, err
	}

	actorState = b.branch(accountState)
	b.branchedAccountsState[string(actor)] = actorState

	return actorState, nil
}

func (b WritersMap) ApplyStateChanges(stateChanges []store.StateChanges) error {
	for _, sc := range stateChanges {
		err := b.applyStateChange(sc)
		if err != nil {
			return fmt.Errorf("unable to apply state change for actor %X: %w", sc.Actor, err)
		}
	}
	return nil
}

func (b WritersMap) GetStateChanges() ([]store.StateChanges, error) {
	sc := make([]store.StateChanges, len(b.branchedAccountsState))
	for account, stateChange := range b.branchedAccountsState {
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

func (b WritersMap) applyStateChange(sc store.StateChanges) error {
	writableState, err := b.GetWriter(sc.Actor)
	if err != nil {
		return err
	}
	return writableState.ApplyChangeSets(sc.StateChanges)
}

func unsafeString(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }
