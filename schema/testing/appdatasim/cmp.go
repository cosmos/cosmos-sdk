package appdatasim

import "cosmossdk.io/schema/testing/statesim"

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
		res += "BlockNum: expected %d, got %d\n"
	}

	return res
}
