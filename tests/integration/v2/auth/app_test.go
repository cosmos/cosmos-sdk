package auth

import (
	"testing"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/x/accounts"
	_ "cosmossdk.io/x/accounts" // import as blank for app wiring
	"cosmossdk.io/x/accounts/accountstd"
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	_ "cosmossdk.io/x/bank" // import as blank for app wiring
	bankkeeper "cosmossdk.io/x/bank/keeper"
	_ "cosmossdk.io/x/consensus" // import as blank for app wiring
	_ "cosmossdk.io/x/staking"   // import as blank for app wirings

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth" // import as blank for app wiring
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring``
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
	"github.com/stretchr/testify/require"
)

type suite struct {
	app *integration.App

	cdc codec.Codec
	ctx sdk.Context

	authKeeper     authkeeper.AccountKeeper
	accountsKeeper accounts.Keeper
	bankKeeper     bankkeeper.Keeper
}

func (s suite) mustAddr(address []byte) string {
	str, _ := s.authKeeper.AddressCodec().BytesToString(address)
	return str
}

func createTestSuite(t *testing.T, extraAccs map[string]accountstd.Interface) *suite {
	t.Helper()
	res := suite{}

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.VestingModule(),
		configurator.StakingModule(),
		configurator.TxModule(),
		configurator.ValidateModule(),
		configurator.ConsensusModule(),
		configurator.GenutilModule(),
	}

	var err error
	startupCfg := integration.DefaultStartUpConfig(t)

	var accs []accountstd.DepinjectAccount
	for name, acc := range extraAccs {
		d := accountstd.DepinjectAccount{MakeAccount: newMockAccountsModKeeper(name, acc)}
		accs = append(accs, d)
	}

	res.app, err = integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Provide(
			// inject desired account types:
			basedepinject.ProvideAccount,

			// provide base account options
			basedepinject.ProvideSecp256K1PubKey,

			// provide extra accounts
			accs,
		), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&res.bankKeeper, &res.accountsKeeper)
	require.NoError(t, err)

	return &res
}
