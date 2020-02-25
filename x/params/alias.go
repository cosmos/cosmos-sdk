package params

// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	StoreKey  = types.StoreKey
	TStoreKey = types.TStoreKey
)

var (
	// functions aliases
	NewKeeper = keeper.NewKeeper
)

type (
	Keeper           = keeper.Keeper
	ParamSetPair     = types.ParamSetPair
	ParamSetPairs    = types.ParamSetPairs
	ParamSet         = types.ParamSet
	Subspace         = types.Subspace
	ReadOnlySubspace = types.ReadOnlySubspace
	KeyTable         = types.KeyTable
)
