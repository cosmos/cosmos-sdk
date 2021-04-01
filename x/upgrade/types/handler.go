package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// UpgradeHandler specifies the type of function that is called when an upgrade
// is applied.
//
// `fromVM` is a VersionMap of moduleName to fromVersion (unit64), where
// fromVersion denotes the version from which we should migrate the module, the
// target version being the module's latest version in the return VersionMap,
// let's call it `toVM`.
//
// fromVM is retrieved from x/upgrade's store, whereas `toVM` is chosen
// arbitrarily by the app developer (and persisted to x/upgrade's store right
// after the upgrade handler runs). In general, `toVM` should map module names
// to their latest ConsensusVersion.
type UpgradeHandler func(ctx sdk.Context, plan Plan, fromVM module.VersionMap) (module.VersionMap, error)
