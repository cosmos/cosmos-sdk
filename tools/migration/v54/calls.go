package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

var callUpdates = []migration.FunctionArgUpdate{
	{
		ImportPath:  "github.com/cometbft/cometbft/rpc/client/http",
		FuncName:    "New",
		OldArgCount: 2,
		NewArgCount: 1,
	},
}
