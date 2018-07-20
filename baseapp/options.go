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
		panic(fmt.Sprintf("Invalid pruning strategy: %s", pruning))
	}
	return func(bap *BaseApp) {
		bap.cms.SetPruning(pruningEnum)
	}
}
