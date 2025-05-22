package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

var moduleUpdates = migration.GoModUpdate{
	"github.com/cosmos/cosmos-sdk": "v0.53.0-rc.4.0.20250522154455-f7601e5b81c2",
	"cosmossdk.io/store":           "v1.10.0-rc.1.0.20250522154455-f7601e5b81c2",
}
