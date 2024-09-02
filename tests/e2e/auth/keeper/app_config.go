package keeper

import (
	_ "cosmossdk.io/x/accounts"       // import as blank for app wiring
	_ "cosmossdk.io/x/auth"           // import as blank for app wiring
	_ "cosmossdk.io/x/auth/tx/config" // import as blank for app wiring``
	_ "cosmossdk.io/x/auth/vesting"   // import as blank for app wiring
	_ "cosmossdk.io/x/bank"           // import as blank for app wiring
	_ "cosmossdk.io/x/comet"          // import as blank for app wiring
	_ "cosmossdk.io/x/staking"        // import as blank for app wiring

	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	_ "github.com/cosmos/cosmos-sdk/x/genutil" // import as blank for app wiring
)

var AppConfig = configurator.NewAppConfig(
	configurator.AccountsModule(),
	configurator.AuthModule(),
	configurator.BankModule(),
	configurator.VestingModule(),
	configurator.StakingModule(),
	configurator.TxModule(),
	configurator.ConsensusModule(),
	configurator.GenutilModule(),
)
