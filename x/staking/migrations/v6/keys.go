package v6

import "cosmossdk.io/collections"

var (
	UnbondingTypeKey    = collections.NewPrefix(57) // prefix for an index containing the type of unbonding operations
	UnbondingIDKey      = collections.NewPrefix(55) // key for the counter for the incrementing id for UnbondingOperations
	UnbondingIndexKey   = collections.NewPrefix(56) // prefix for an index for looking up unbonding operations by their IDs
	ValidatorUpdatesKey = collections.NewPrefix(97)
	HistoricalInfoKey   = collections.NewPrefix(80) // prefix for the historical info
)
