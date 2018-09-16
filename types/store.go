package types

import (
	"github.com/cosmos/cosmos-sdk/store/types"
)

// nolint: reexport
type (
	KVStore    = types.KVStore
	MultiStore = types.MultiStore
	StoreKey   = types.StoreKey
	CommitID   = types.CommitID
	Gas        = types.Gas
	GasTank    = types.GasTank
	GasConfig  = types.GasConfig
	GasMeter   = types.GasMeter
)

// nolint: reexport
func NewGasTank(limit Gas, config GasConfig) *GasTank {
	return types.NewGasTank(limit, config)
}
func NewInfiniteGasTank(config GasConfig) *GasTank {
	return types.NewInfiniteGasTank(config)
}
