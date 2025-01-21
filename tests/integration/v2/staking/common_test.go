package staking

import (
	"context"
	"math/big"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	_ "cosmossdk.io/x/accounts" // import as blank for app wiring
	_ "cosmossdk.io/x/bank"     // import as blank for app wiring
	bankkeeper "cosmossdk.io/x/bank/keeper"
	_ "cosmossdk.io/x/consensus" // import as blank for app wiring
	consensuskeeper "cosmossdk.io/x/consensus/keeper"
	_ "cosmossdk.io/x/mint"         // import as blank for app wiring
	_ "cosmossdk.io/x/protocolpool" // import as blank for app wiring
	_ "cosmossdk.io/x/slashing"     // import as blank for app wiring
	slashingkeeper "cosmossdk.io/x/slashing/keeper"
	_ "cosmossdk.io/x/staking" // import as blank for app wiring
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/testutil"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth" // import as blank for app wiring
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
)

var (
	PKs = simtestutil.CreateTestPubKeys(500)

	mockStakingHook = types.StakingHooksWrapper{}
)

type fixture struct {
	app *integration.App

	ctx context.Context
	cdc codec.Codec

	queryClient stakingkeeper.Querier

	accountKeeper   authkeeper.AccountKeeper
	bankKeeper      bankkeeper.Keeper
	stakingKeeper   *stakingkeeper.Keeper
	slashKeeper     slashingkeeper.Keeper
	consensusKeeper consensuskeeper.Keeper
}

func init() {
	sdk.DefaultPowerReduction = math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
}

// intended to be used with require/assert:  require.True(ValEq(...))
func ValEq(t *testing.T, exp, got types.Validator) (*testing.T, bool, string, types.Validator, types.Validator) {
	t.Helper()
	return t, exp.MinEqual(&got), "expected:\n%v\ngot:\n%v", exp, got
}

// generateAddresses generates numAddrs of normal AccAddrs and ValAddrs
func generateAddresses(f *fixture, numAddrs int) ([]sdk.AccAddress, []sdk.ValAddress) {
	addrDels := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, f.ctx, numAddrs, math.NewInt(10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	return addrDels, addrVals
}

func createValidators(
	t *testing.T,
	f *fixture,
	powers []int64,
) ([]sdk.AccAddress, []sdk.ValAddress, []types.Validator) {
	t.Helper()
	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, f.ctx, 5, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 300))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	pks := simtestutil.CreateTestPubKeys(5)

	val1 := testutil.NewValidator(t, valAddrs[0], pks[0])
	val2 := testutil.NewValidator(t, valAddrs[1], pks[1])
	vals := []types.Validator{val1, val2}

	assert.NilError(t, f.stakingKeeper.SetValidator(f.ctx, val1))
	assert.NilError(t, f.stakingKeeper.SetValidator(f.ctx, val2))
	assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.ctx, val1))
	assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.ctx, val2))
	assert.NilError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.ctx, val1))
	assert.NilError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.ctx, val2))

	for _, addr := range addrs {
		acc := f.accountKeeper.NewAccountWithAddress(f.ctx, addr)
		f.accountKeeper.SetAccount(f.ctx, acc)
	}

	_, err := f.stakingKeeper.Delegate(f.ctx, addrs[0], f.stakingKeeper.TokensFromConsensusPower(f.ctx, powers[0]), types.Unbonded, val1, true)
	assert.NilError(t, err)
	_, err = f.stakingKeeper.Delegate(f.ctx, addrs[1], f.stakingKeeper.TokensFromConsensusPower(f.ctx, powers[1]), types.Unbonded, val2, true)
	assert.NilError(t, err)
	_, err = f.stakingKeeper.Delegate(f.ctx, addrs[0], f.stakingKeeper.TokensFromConsensusPower(f.ctx, powers[2]), types.Unbonded, val2, true)
	assert.NilError(t, err)
	applyValidatorSetUpdates(t, f.ctx, f.stakingKeeper, -1)

	return addrs, valAddrs, vals
}

func ProvideMockStakingHook() types.StakingHooksWrapper {
	return mockStakingHook
}

func initFixture(tb testing.TB, isGenesisSkip bool, stakingHooks ...types.StakingHooksWrapper) *fixture {
	tb.Helper()

	res := fixture{}

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.StakingModule(),
		configurator.SlashingModule(),
		configurator.TxModule(),
		configurator.ValidateModule(),
		configurator.ConsensusModule(),
		configurator.GenutilModule(),
		configurator.MintModule(),
		configurator.ProtocolPoolModule(),
	}

	configs := []depinject.Config{
		configurator.NewAppV2Config(moduleConfigs...),
		depinject.Supply(log.NewNopLogger()),
	}

	// add mock staking hooks if given
	if len(stakingHooks) != 0 {
		mockStakingHook = stakingHooks[0]
		configs = append(configs, depinject.ProvideInModule(
			"mock", ProvideMockStakingHook,
		))
	}

	var err error
	startupCfg := integration.DefaultStartUpConfig(tb)

	startupCfg.BranchService = &integration.BranchService{}
	startupCfg.HeaderService = &integration.HeaderService{}
	if isGenesisSkip {
		startupCfg.GenesisBehavior = integration.Genesis_SKIP
	}

	res.app, err = integration.NewApp(
		depinject.Configs(configs...),
		startupCfg,
		&res.bankKeeper, &res.accountKeeper, &res.stakingKeeper,
		&res.slashKeeper, &res.consensusKeeper, &res.cdc)
	assert.NilError(tb, err)

	res.ctx = res.app.StateLatestContext(tb)

	// set default staking params
	assert.NilError(tb, res.stakingKeeper.Params.Set(res.ctx, types.DefaultParams()))

	res.queryClient = stakingkeeper.NewQuerier(res.stakingKeeper)

	return &res
}
