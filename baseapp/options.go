package baseapp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// File for storing in-package BaseApp optional functions,
// for options that need access to non-exported fields of the BaseApp

// SetPruning sets a pruning option on the multistore associated with the app
func SetPruning(pruning string) func(*BaseApp) {
	var pruningEnum sdk.PruningStrategy
	switch pruning {
	case "nothing":
		pruningEnum = sdk.PruneNothing
	case "everything":
		pruningEnum = sdk.PruneEverything
	case "syncable":
		pruningEnum = sdk.PruneSyncable
	default:
		panic(fmt.Sprintf("invalid pruning strategy: %s", pruning))
	}
	return func(bap *BaseApp) {
		bap.cms.SetPruning(pruningEnum)
	}
}

// SetMinimumFees returns an option that sets the minimum fees on the app.
func SetMinimumFees(minFees string) func(*BaseApp) {
	fees, err := sdk.ParseCoins(minFees)
	if err != nil {
		panic(fmt.Sprintf("invalid minimum fees: %v", err))
	}
	return func(bap *BaseApp) { bap.SetMinimumFees(fees) }
}
