package testutil

import (
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/depinject/appconfig"

	nftmodulev1 "github.com/cosmos/cosmos-sdk/contrib/api/cosmos/nft/module/v1"
	_ "github.com/cosmos/cosmos-sdk/contrib/x/nft/module" // import as blank for app wiring
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	_ "github.com/cosmos/cosmos-sdk/x/auth"           // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/bank"           // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/consensus"      // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/mint"           // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/params"         // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/staking"        // import as blank for app wiring
)

func nftModule() configurator.ModuleOption {
	return func(config *configurator.Config) {
		config.ModuleConfigs["nft"] = &appv1alpha1.ModuleConfig{
			Name:   "nft",
			Config: appconfig.WrapAny(&nftmodulev1.Module{}),
		}
	}
}

var AppConfig = configurator.NewAppConfig(
	configurator.AuthModule(),
	configurator.BankModule(),
	configurator.StakingModule(),
	configurator.TxModule(),
	configurator.ConsensusModule(),
	configurator.ParamsModule(),
	configurator.GenutilModule(),
	configurator.MintModule(),
	nftModule(),
)
