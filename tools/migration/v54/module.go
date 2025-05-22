package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

var moduleUpdates = migration.GoModUpdate{
	"github.com/cosmos/cosmos-sdk": "f7601e5",
	"cosmossdk.io/store":           "f7601e5",
}
