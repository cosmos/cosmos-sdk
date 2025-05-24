package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

var moduleUpdates = migration.GoModUpdate{
	"github.com/cosmos/cosmos-sdk": "v0.53.0-rc.4.0.20250522154455-f7601e5b81c2",
	"cosmossdk.io/store":           "v1.10.0-rc.1.0.20250522154455-f7601e5b81c2",
	"github.com/cometbft/cometbft": "v1.0.1",
	"google.golang.org/grpc":       "v1.72.1",
	"cosmossdk.io/x/tx":            "v1.1.1-0.20250515174933-df4c1c3edf16",
	"cosmossdk.io/client/v2":       "v2.0.0-beta.8",
}

var removals = []string{
	"cosmossdk.io/x/circuit",
	"cosmossdk.io/x/evidence",
	"cosmossdk.io/x/upgrade",
	"cosmossdk.io/x/nft",
	"cosmossdk.io/x/feegrant",
}
