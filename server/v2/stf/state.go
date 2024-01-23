package stf

import (
	"fmt"

	"cosmossdk.io/server/v2/core/store"
)

type branchedState struct {
	state                 store.GetReader
	branchedAccountsState map[string]store.Writer
	branch                func(state store.Reader) store.Writer
}

func (b branchedState) GetAccountReader(address []byte) (store.Reader, error) {
	return b.GetAccountWriter(address)
}

func (b branchedState) GetAccountWriter(address []byte) (store.Writer, error) {
	addressStr := string(address)
	// this is the case in which we have already cached some branched state.
	branchedState, ok := b.branchedAccountsState[addressStr]
	if ok {
		return branchedState, nil
	}
	// this is the case in which it's the first time in the execution context
	// we were asked for this account's state, so we will fetch it from the state.
	accountState, err := b.state.GetAccountReader(address)
	if err != nil {
		return nil, err
	}

	branchedState = b.branch(accountState)
	b.branchedAccountsState[addressStr] = branchedState
	return branchedState, nil
}

func (b branchedState) ApplyStateChanges(stateChanges []store.StateChanges) error {
	for _, sc := range stateChanges {
		err := b.applyStateChange(sc)
		if err != nil {
			return fmt.Errorf("unable to apply state change for address %X: %w", sc.Account, err)
		}
	}
	return nil
}

func (b branchedState) GetStateChanges() ([]store.StateChanges, error) {
	sc := make([]store.StateChanges, len(b.branchedAccountsState))
	for account, stateChange := range b.branchedAccountsState {
		kvChanges, err := stateChange.ChangeSets()
		if err != nil {
			return nil, err
		}
		sc = append(sc, store.StateChanges{
			Account:      []byte(account),
			StateChanges: kvChanges,
		})
	}
	return sc, nil
}

func (b branchedState) applyStateChange(sc store.StateChanges) error {
	writableState, err := b.GetAccountWriter(sc.Account)
	if err != nil {
		return err
	}
	return writableState.ApplyChangeSets(sc.StateChanges)
}

func newBranchedAccountsState(
	state store.GetReader,
	branch func(readonlyState store.Reader) store.Writer) store.GetWriter {

	return branchedState{
		state:                 state,
		branchedAccountsState: make(map[string]store.Writer),
		branch:                branch,
	}
}
