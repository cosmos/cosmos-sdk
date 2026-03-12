package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// callUpdates defines function signature changes where arguments were added or removed.
//
// TODO: audit v54 breaking changes for function signature changes.
// Known areas to investigate:
// - bankKeeper initialization (new args for BlockSTM support)
// - store v2 API changes (IAVLx new store API)
// - any baseapp method signature changes
var callUpdates = []migration.FunctionArgUpdate{
	// Example format:
	// {
	// 	ImportPath:  "github.com/cosmos/cosmos-sdk/x/bank/keeper",
	// 	FuncName:    "NewKeeper",
	// 	OldArgCount: 6,
	// 	NewArgCount: 7,
	// },
}
