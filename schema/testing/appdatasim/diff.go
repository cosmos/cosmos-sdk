package appdatasim

import (
	"fmt"

	"cosmossdk.io/schema/testing/statesim"
	"cosmossdk.io/schema/view"
)

// DiffAppData compares the app data of two objects that implement HasAppData.
// This can be used by indexer to compare their state with the Simulator state
// if the indexer implements HasAppData.
// It returns a human-readable diff if the app data differs and the empty string
// if they are the same.
func DiffAppData(expected, actual view.AppData) string {
	res := ""

	if stateDiff := statesim.DiffAppStates(expected.AppState(), actual.AppState()); stateDiff != "" {
		res += "App State Diff:\n"
		res += stateDiff
	}

	expectedBlock, err := expected.BlockNum()
	if err != nil {
		res += fmt.Sprintf("ERROR getting expected block num: %s\n", err)
		return res
	}

	actualBlock, err := actual.BlockNum()
	if err != nil {
		res += fmt.Sprintf("ERROR getting actual block num: %s\n", err)
		return res
	}

	if expectedBlock != actualBlock {
		res += fmt.Sprintf("BlockNum: expected %d, got %d\n", expectedBlock, actualBlock)
	}

	return res
}
