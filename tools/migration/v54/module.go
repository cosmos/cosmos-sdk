package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

var moduleUpdates = migration.GoModUpdate{
	"github.com/cosmos/cosmos-sdk": "v0.54.0-alpha.0.0.20250602195229-601ab15623c5",
	"cosmossdk.io/store":           "v1.10.0-rc.1.0.20250602195229-601ab15623c5",
	// "github.com/cometbft/cometbft": "v1.0.1",
	"google.golang.org/grpc": "v1.72.1",
	"cosmossdk.io/x/tx":      "v1.2.0-alpha.0.0.20250602195229-601ab15623c5",
	"cosmossdk.io/client/v2": "v2.0.0-beta.10.0.20250602195229-601ab15623c5",
	"cosmossdk.io/core":      "v1.1.0-alpha.1.0.20250602195229-601ab15623c5",
	"cosmossdk.io/simapp":    "v0.0.0-20250602195229-601ab15623c5",
}

var replacements = []migration.GoModReplacement{
	{
		Module:      "github.com/cometbft/cometbft",
		Replacement: "github.com/zrbecker/cometbft",
		Version:     "v0.0.0-20250531072547-03a237b0b780",
	},
	{
		Module:      "github.com/cometbft/cometbft/api",
		Replacement: "github.com/zrbecker/cometbft/api",
		Version:     "v0.0.0-20250531072547-03a237b0b780",
	},
}

var removals = []string{
	"cosmossdk.io/x/circuit",
	"cosmossdk.io/x/evidence",
	"cosmossdk.io/x/upgrade",
	"cosmossdk.io/x/nft",
	"cosmossdk.io/x/feegrant",
}
