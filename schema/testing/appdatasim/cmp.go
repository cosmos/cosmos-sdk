package appdatasim

import (
	"fmt"

	"cosmossdk.io/schema/testing/statesim"
)

// HasAppData defines an interface for things that hold app data include app state.
// If an indexer implements this then DiffAppData can be used to compare it with
// the Simulator state which also implements this.
type HasAppData interface {
	// AppState returns the app state.
	AppState() statesim.AppState

	// BlockNum returns the latest block number.
	BlockNum() uint64
}

// DiffAppData compares the app data of two objects that implement HasAppData.
// This can be used by indexer to compare their state with the Simulator state
// if the indexer implements HasAppData.
// It returns a human-readable diff if the app data differs and the empty string
// if they are the same.
func DiffAppData(expected, actual HasAppData) string {
	res := ""

	if stateDiff := statesim.DiffAppStates(expected.AppState(), actual.AppState()); stateDiff != "" {
		res += "App State Diff:\n"
		res += stateDiff
	}

	if expected.BlockNum() != actual.BlockNum() {
		res += fmt.Sprintf("BlockNum: expected %d, got %d\n", expected.BlockNum(), actual.BlockNum())
	}

	return res
}
