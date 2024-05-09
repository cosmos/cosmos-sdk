package nft

import (
	_ "cosmossdk.io/x/accounts"       // import as blank for app wiring
	_ "cosmossdk.io/x/auth"           // import as blank for app wiring
	_ "cosmossdk.io/x/auth/tx/config" // import as blank for app wiring
	_ "cosmossdk.io/x/bank"           // import as blank for app wiring
	_ "cosmossdk.io/x/consensus"      // import as blank for app wiring
	_ "cosmossdk.io/x/mint"           // import as blank for app wiring
	_ "cosmossdk.io/x/nft/module"     // import as blank for app wiring
	_ "cosmossdk.io/x/staking"        // import as blank for app wiring

	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	_ "github.com/cosmos/cosmos-sdk/x/genutil" // import as blank for app wiring
)

var AppConfig = configurator.NewAppConfig(
	configurator.AccountsModule(),
	configurator.AuthModule(),
	configurator.BankModule(),
	configurator.StakingModule(),
	configurator.TxModule(),
	configurator.ConsensusModule(),
	configurator.GenutilModule(),
	configurator.MintModule(),
	configurator.NFTModule(),
)
