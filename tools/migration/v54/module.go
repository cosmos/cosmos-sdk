package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

var moduleUpdates = migration.GoModUpdate{
	"github.com/cosmos/cosmos-sdk":     "v0.54.0-alpha.0.0.20250604161429-8c61b74a1806",
	"cosmossdk.io/store":               "v1.10.0-rc.1.0.20250604161429-8c61b74a1806",
	"github.com/cometbft/cometbft":     "v2.0.0-20250604002332-f4d33abd2469",
	"github.com/cometbft/cometbft/api": "v0.0.0-20250604002332-f4d33abd2469",
	"google.golang.org/grpc":           "v1.72.1",
	"cosmossdk.io/x/tx":                "v1.2.0-alpha.0.0.20250604161429-8c61b74a1806",
	"cosmossdk.io/client/v2":           "v2.0.0-beta.10.0.20250604161429-8c61b74a1806",
	"cosmossdk.io/core":                "v1.1.0-alpha.1.0.20250604161429-8c61b74a1806",
	"cosmossdk.io/simapp":              "v0.0.0-20250602195229-601ab15623c5",
	"cosmossdk.io/api":                 "v1.0.0-alpha.0.0.20250604161429-8c61b74a1806",
}

var removals = []string{
	"cosmossdk.io/x/circuit",
	"cosmossdk.io/x/evidence",
	"cosmossdk.io/x/upgrade",
	"cosmossdk.io/x/nft",
	"cosmossdk.io/x/feegrant",
}
