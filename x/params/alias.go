package params

import (
	"github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	StoreKey     = types.StoreKey
	TStoreKey    = types.TStoreKey
	ModuleName   = types.ModuleName
	QuerierRoute = types.QuerierRoute
)

var (
	// functions aliases
	NewKeeper                 = keeper.NewKeeper
	NewParamSetPair           = types.NewParamSetPair
	NewKeyTable               = types.NewKeyTable
	NewQuerySubspaceParams    = types.NewQuerySubspaceParams
	NewQuerier                = keeper.NewQuerier
	NewSubspaceParamsResponse = types.NewSubspaceParamsResponse
)

type (
	Keeper                 = keeper.Keeper
	ParamSetPair           = types.ParamSetPair
	ParamSetPairs          = types.ParamSetPairs
	ParamSet               = types.ParamSet
	Subspace               = types.Subspace
	ReadOnlySubspace       = types.ReadOnlySubspace
	KeyTable               = types.KeyTable
	QuerySubspaceParams    = types.QuerySubspaceParams
	SubspaceParamsResponse = types.SubspaceParamsResponse
)
