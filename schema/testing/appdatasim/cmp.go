package appdatasim

import (
	"fmt"

	"cosmossdk.io/schema/testing/statesim"
)

type HasAppData interface {
	AppState() statesim.AppState
	BlockNum() uint64
}

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
