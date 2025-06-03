package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

var importReplacements = []migration.ImportReplacement{
	{Old: "github.com/cometbft/cometbft/proto/tendermint/types", New: "github.com/cometbft/cometbft/api/cometbft/types/v2"},
	{Old: "github.com/cometbft/cometbft/proto/tendermint/crypto", New: "github.com/cometbft/cometbft/api/cometbft/crypto/v1"},
	{Old: "github.com/cometbft/cometbft/proto/tendermint/state", New: "github.com/cometbft/cometbft/api/cometbft/state/v2"},
	{Old: "github.com/cometbft/cometbft/proto/tendermint/version", New: "github.com/cometbft/cometbft/api/cometbft/version/v1"},
	{Old: "github.com/cometbft/cometbft/proto/tendermint/p2p", New: "github.com/cometbft/cometbft/api/cometbft/p2p/v1"},
	{Old: "cosmossdk.io/x/feegrant", New: "github.com/cosmos/cosmos-sdk/x/feegrant", AllPackages: true},
	{Old: "cosmossdk.io/x/circuit", New: "github.com/cosmos/cosmos-sdk/x/circuit", AllPackages: true},
	{Old: "cosmossdk.io/x/nft", New: "github.com/cosmos/cosmos-sdk/x/nft", AllPackages: true},
	{Old: "cosmossdk.io/x/evidence", New: "github.com/cosmos/cosmos-sdk/x/evidence", AllPackages: true},
	{Old: "cosmossdk.io/x/upgrade", New: "github.com/cosmos/cosmos-sdk/x/upgrade", AllPackages: true},
}
