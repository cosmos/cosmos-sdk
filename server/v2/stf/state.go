package stf

import (
	"fmt"

	"cosmossdk.io/server/v2/core/store"
)

type cachedWritersMap struct {
	state                 store.ReaderMap
	branchedAccountsState map[string]store.Writer
	branch                func(state store.Reader) store.Writer
}

func (b cachedWritersMap) GetReader(actor []byte) (store.Reader, error) {
	return b.GetWriter(actor)
}

func (b cachedWritersMap) GetWriter(actor []byte) (store.Writer, error) {
	actorStr := string(actor)
	// this is the case in which we have already cached some branched state.
	actorState, ok := b.branchedAccountsState[actorStr]
	if ok {
		return actorState, nil
	}
	// this is the case in which it's the first time in the execution context
	// we were asked for this account's state, so we will fetch it from the state.
	accountState, err := b.state.GetReader(actor)
	if err != nil {
		return nil, err
	}

	actorState = b.branch(accountState)
	b.branchedAccountsState[actorStr] = actorState
	return actorState, nil
}

func (b cachedWritersMap) ApplyStateChanges(stateChanges []store.StateChanges) error {
	for _, sc := range stateChanges {
		err := b.applyStateChange(sc)
		if err != nil {
			return fmt.Errorf("unable to apply state change for actor %X: %w", sc.Actor, err)
		}
	}
	return nil
}

func (b cachedWritersMap) GetStateChanges() ([]store.StateChanges, error) {
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

func (b cachedWritersMap) applyStateChange(sc store.StateChanges) error {
	writableState, err := b.GetWriter(sc.Actor)
	if err != nil {
		return err
	}
	return writableState.ApplyChangeSets(sc.StateChanges)
}

func newBranchedAccountsState(
	state store.ReaderMap,
	branch func(readonlyState store.Reader) store.Writer) store.WriterMap {

	return cachedWritersMap{
		state:                 state,
		branchedAccountsState: make(map[string]store.Writer),
		branch:                branch,
	}
}
