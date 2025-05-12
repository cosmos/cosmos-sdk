package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

var (
	importReplacements = []migration.ImportReplacement{
		{Old: "github.com/cometbft/cometbft/proto/tendermint/types", New: "github.com/cometbft/cometbft/api/cometbft/types/v1"},
		{Old: "github.com/cometbft/cometbft/proto/tendermint/crypto", New: "github.com/cometbft/cometbft/api/cometbft/crypto/v1"},
		{Old: "github.com/cometbft/cometbft/proto/tendermint/state", New: "github.com/cometbft/cometbft/api/cometbft/state/v1"},
	}
)
