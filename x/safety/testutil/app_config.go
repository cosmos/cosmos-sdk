package testutil

import (
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	safetymodulev1 "cosmossdk.io/api/cosmos/safety/module/v1"
	"cosmossdk.io/core/appconfig"

	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	_ "github.com/cosmos/cosmos-sdk/x/auth"           // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/bank"           // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/consensus"      // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/safety"         // import as blank for app wiring
	"github.com/cosmos/cosmos-sdk/x/safety/types"
	_ "github.com/cosmos/cosmos-sdk/x/staking" // import as blank for app wiring
)

// SafetyModule returns a configurator option that wires x/safety. authority
// is written into the default genesis; empty means the gov module.
func SafetyModule(authority string) configurator.ModuleOption {
	return func(config *configurator.Config) {
		config.ModuleConfigs[types.ModuleName] = &appv1alpha1.ModuleConfig{
			Name:   types.ModuleName,
			Config: appconfig.WrapAny(&safetymodulev1.Module{Authority: authority}),
		}
	}
}
