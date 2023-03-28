package integration_test

import (
	"testing"

	"gotest.tools/v3/assert"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func TestIntegrationTestExample(t *testing.T) {
	// in this example we are testing the integration of the following modules:
	// - mint, which directly depends on auth, bank and staking

	encodingCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, mint.AppModuleBasic{})
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, minttypes.StoreKey)
	authority := authtypes.NewModuleAddress("gov").String()

	accountKeeper := authkeeper.NewAccountKeeper(
		encodingCfg.Codec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		map[string][]string{minttypes.ModuleName: {authtypes.Minter}},
		"cosmos",
		authority,
	)

	// subspace is nil because we don't test params (which is legacy anyway)
	authModule := auth.NewAppModule(encodingCfg.Codec, accountKeeper, authsims.RandomGenesisAccounts, nil)

	// here bankkeeper and staking keeper is nil because we are not testing them
	// subspace is nil because we don't test params (which is legacy anyway)
	mintKeeper := mintkeeper.NewKeeper(encodingCfg.Codec, keys[minttypes.StoreKey], nil, accountKeeper, nil, authtypes.FeeCollectorName, authority)
	mintModule := mint.NewAppModule(encodingCfg.Codec, mintKeeper, accountKeeper, nil, nil)

	integrationApp := integration.NewIntegrationApp(t, keys, authModule, mintModule)

	// register the message and query servers
	authtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), authkeeper.NewMsgServerImpl(accountKeeper))
	minttypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), mintkeeper.NewMsgServerImpl(mintKeeper))
	minttypes.RegisterQueryServer(integrationApp.QueryHelper(), mintKeeper)

	// now we can use the application to test an mint message
	result, err := integrationApp.RunMsg(&minttypes.MsgUpdateParams{
		Authority: authority,
		Params:    minttypes.DefaultParams(),
	})
	assert.NilError(t, err)
	assert.Assert(t, result != nil)

	// we now check the result
	resp := minttypes.MsgUpdateParamsResponse{}
	err = encodingCfg.Codec.Unmarshal(result.Value, &resp)
	assert.NilError(t, err)
}
