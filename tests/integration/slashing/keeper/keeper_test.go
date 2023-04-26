package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type fixture struct {
	app *integration.App

	ctx sdk.Context

	bankKeeper     bankkeeper.Keeper
	slashingKeeper slashingkeeper.Keeper
	stakingKeeper  *stakingkeeper.Keeper

	addrDels []sdk.AccAddress
}

func initFixture(t testing.TB) *fixture {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, slashingtypes.StoreKey, stakingtypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, distribution.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, cmtproto.Header{}, true, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddresses,
		authority.String(),
		log.NewNopLogger(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(cdc, keys[stakingtypes.StoreKey], accountKeeper, bankKeeper, authority.String())

	slashingKeeper := slashingkeeper.NewKeeper(cdc, &codec.LegacyAmino{}, keys[slashingtypes.StoreKey], stakingKeeper, authority.String())

	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)
	slashingModule := slashing.NewAppModule(cdc, slashingKeeper, accountKeeper, bankKeeper, stakingKeeper, nil, cdc.InterfaceRegistry())

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, bankModule, stakingModule, slashingModule)

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	// Register MsgServer and QueryServer
	slashingtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), slashingkeeper.NewMsgServerImpl(slashingKeeper))
	slashingtypes.RegisterQueryServer(integrationApp.QueryHelper(), slashingkeeper.NewQuerier(slashingKeeper))

	// set default staking params
	stakingKeeper.SetParams(sdkCtx, stakingtypes.DefaultParams())

	// TestParams set the SignedBlocksWindow to 1000 and MaxMissedBlocksPerWindow to 500
	slashingKeeper.SetParams(sdkCtx, testutil.TestParams())
	addrDels := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, sdkCtx, 6, stakingKeeper.TokensFromConsensusPower(sdkCtx, 200))

	info1 := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[0]), int64(4), int64(3), time.Unix(2, 0), false, int64(10))
	info2 := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[1]), int64(5), int64(4), time.Unix(2, 0), false, int64(10))

	slashingKeeper.SetValidatorSigningInfo(sdkCtx, sdk.ConsAddress(addrDels[0]), info1)
	slashingKeeper.SetValidatorSigningInfo(sdkCtx, sdk.ConsAddress(addrDels[1]), info2)

	return &fixture{
		app:            integrationApp,
		ctx:            sdkCtx,
		bankKeeper:     bankKeeper,
		slashingKeeper: slashingKeeper,
		stakingKeeper:  stakingKeeper,
		addrDels:       addrDels,
	}
}

func TestUnJailNotBonded(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	p := f.stakingKeeper.GetParams(f.ctx)
	p.MaxValidators = 5
	f.stakingKeeper.SetParams(f.ctx, p)

	valAddrs := simtestutil.ConvertAddrsToValAddrs(f.addrDels)
	pks := simtestutil.CreateTestPubKeys(6)
	tstaking := stakingtestutil.NewHelper(t, f.ctx, f.stakingKeeper)

	// create max (5) validators all with the same power
	for i := uint32(0); i < p.MaxValidators; i++ {
		addr, val := valAddrs[i], pks[i]
		tstaking.CreateValidatorWithValPower(addr, val, 100, true)
	}

	f.stakingKeeper.EndBlocker(f.ctx)
	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1)

	// create a 6th validator with less power than the cliff validator (won't be bonded)
	addr, val := valAddrs[5], pks[5]
	amt := f.stakingKeeper.TokensFromConsensusPower(f.ctx, 50)
	msg := tstaking.CreateValidatorMsg(addr, val, amt)
	msg.MinSelfDelegation = amt
	msg.Description = stakingtypes.Description{Moniker: "TestValidator"}
	res, err := tstaking.CreateValidatorWithMsg(f.ctx, msg)
	assert.NilError(t, err)
	assert.Assert(t, res != nil)

	f.stakingKeeper.EndBlocker(f.ctx)
	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1)

	tstaking.CheckValidator(addr, stakingtypes.Unbonded, false)

	// unbond below minimum self-delegation
	assert.Equal(t, p.BondDenom, tstaking.Denom)
	tstaking.Undelegate(sdk.AccAddress(addr), addr, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 1), true)

	f.stakingKeeper.EndBlocker(f.ctx)
	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1)

	// verify that validator is jailed
	tstaking.CheckValidator(addr, -1, true)

	// verify we cannot unjail (yet)
	msgUnjail := slashingtypes.MsgUnjail{
		ValidatorAddr: addr.String(),
	}
	_, err = f.app.RunMsg(
		&msgUnjail,
		integration.WithAutomaticBeginEndBlock(),
		integration.WithAutomaticCommit(),
	)
	assert.ErrorContains(t, err, "cannot be unjailed")

	f.stakingKeeper.EndBlocker(f.ctx)
	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1)
	// bond to meet minimum self-delegation
	tstaking.DelegateWithPower(sdk.AccAddress(addr), addr, 1)

	f.stakingKeeper.EndBlocker(f.ctx)
	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1)

	// verify we can immediately unjail
	_, err = f.app.RunMsg(
		&msgUnjail,
		integration.WithAutomaticBeginEndBlock(),
		integration.WithAutomaticCommit(),
	)
	assert.NilError(t, err)

	tstaking.CheckValidator(addr, -1, false)
}

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately jailed
func TestHandleNewValidator(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	valAddrs := simtestutil.ConvertAddrsToValAddrs(f.addrDels)
	pks := simtestutil.CreateTestPubKeys(1)
	addr, val := valAddrs[0], pks[0]
	tstaking := stakingtestutil.NewHelper(t, f.ctx, f.stakingKeeper)
	f.ctx = f.ctx.WithBlockHeight(f.slashingKeeper.SignedBlocksWindow(f.ctx) + 1)

	f.slashingKeeper.AddPubkey(f.ctx, pks[0])

	info := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(val.Address()), f.ctx.BlockHeight(), int64(0), time.Unix(0, 0), false, int64(0))
	f.slashingKeeper.SetValidatorSigningInfo(f.ctx, sdk.ConsAddress(val.Address()), info)

	// Validator created
	amt := tstaking.CreateValidatorWithValPower(addr, val, 100, true)

	f.stakingKeeper.EndBlocker(f.ctx)
	assert.DeepEqual(
		t, f.bankKeeper.GetAllBalances(f.ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(f.stakingKeeper.GetParams(f.ctx).BondDenom, testutil.InitTokens.Sub(amt))),
	)
	assert.DeepEqual(t, amt, f.stakingKeeper.Validator(f.ctx, addr).GetBondedTokens())

	// Now a validator, for two blocks
	f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), 100, true)
	f.ctx = f.ctx.WithBlockHeight(f.slashingKeeper.SignedBlocksWindow(f.ctx) + 2)
	f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), 100, false)

	info, found := f.slashingKeeper.GetValidatorSigningInfo(f.ctx, sdk.ConsAddress(val.Address()))
	assert.Assert(t, found)
	assert.Equal(t, f.slashingKeeper.SignedBlocksWindow(f.ctx)+1, info.StartHeight)
	assert.Equal(t, int64(2), info.IndexOffset)
	assert.Equal(t, int64(1), info.MissedBlocksCounter)
	assert.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)

	// validator should be bonded still, should not have been jailed or slashed
	validator, _ := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, sdk.GetConsAddress(val))
	assert.Equal(t, stakingtypes.Bonded, validator.GetStatus())
	bondPool := f.stakingKeeper.GetBondedPool(f.ctx)
	expTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, 100)
	assert.Assert(t, expTokens.Equal(f.bankKeeper.GetBalance(f.ctx, bondPool.GetAddress(), f.stakingKeeper.BondDenom(f.ctx)).Amount))
}

// Test a jailed validator being "down" twice
// Ensure that they're only slashed once
func TestHandleAlreadyJailed(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	valAddrs := simtestutil.ConvertAddrsToValAddrs(f.addrDels)
	pks := simtestutil.CreateTestPubKeys(1)
	addr, val := valAddrs[0], pks[0]
	power := int64(100)
	tstaking := stakingtestutil.NewHelper(t, f.ctx, f.stakingKeeper)

	f.slashingKeeper.AddPubkey(f.ctx, pks[0])

	info := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(val.Address()), f.ctx.BlockHeight(), int64(0), time.Unix(0, 0), false, int64(0))
	f.slashingKeeper.SetValidatorSigningInfo(f.ctx, sdk.ConsAddress(val.Address()), info)

	amt := tstaking.CreateValidatorWithValPower(addr, val, power, true)

	f.stakingKeeper.EndBlocker(f.ctx)

	// 1000 first blocks OK
	height := int64(0)
	for ; height < f.slashingKeeper.SignedBlocksWindow(f.ctx); height++ {
		f.ctx = f.ctx.WithBlockHeight(height)
		f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), power, true)
	}

	// 501 blocks missed
	for ; height < f.slashingKeeper.SignedBlocksWindow(f.ctx)+(f.slashingKeeper.SignedBlocksWindow(f.ctx)-f.slashingKeeper.MinSignedPerWindow(f.ctx))+1; height++ {
		f.ctx = f.ctx.WithBlockHeight(height)
		f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), power, false)
	}

	// end block
	f.stakingKeeper.EndBlocker(f.ctx)

	// validator should have been jailed and slashed
	validator, _ := f.stakingKeeper.GetValidatorByConsAddr(f.ctx, sdk.GetConsAddress(val))
	assert.Equal(t, stakingtypes.Unbonding, validator.GetStatus())

	// validator should have been slashed
	resultingTokens := amt.Sub(f.stakingKeeper.TokensFromConsensusPower(f.ctx, 1))
	assert.DeepEqual(t, resultingTokens, validator.GetTokens())

	// another block missed
	f.ctx = f.ctx.WithBlockHeight(height)
	f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), power, false)

	// validator should not have been slashed twice
	validator, _ = f.stakingKeeper.GetValidatorByConsAddr(f.ctx, sdk.GetConsAddress(val))
	assert.DeepEqual(t, resultingTokens, validator.GetTokens())
}

// Test a validator dipping in and out of the validator set
// Ensure that missed blocks are tracked correctly and that
// the start height of the signing info is reset correctly
func TestValidatorDippingInAndOut(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	params := f.stakingKeeper.GetParams(f.ctx)
	params.MaxValidators = 1
	f.stakingKeeper.SetParams(f.ctx, params)
	power := int64(100)

	pks := simtestutil.CreateTestPubKeys(3)
	simtestutil.AddTestAddrsFromPubKeys(f.bankKeeper, f.stakingKeeper, f.ctx, pks, f.stakingKeeper.TokensFromConsensusPower(f.ctx, 200))

	addr, val := pks[0].Address(), pks[0]
	consAddr := sdk.ConsAddress(addr)
	tstaking := stakingtestutil.NewHelper(t, f.ctx, f.stakingKeeper)
	valAddr := sdk.ValAddress(addr)

	f.slashingKeeper.AddPubkey(f.ctx, pks[0])

	info := slashingtypes.NewValidatorSigningInfo(consAddr, f.ctx.BlockHeight(), int64(0), time.Unix(0, 0), false, int64(0))
	f.slashingKeeper.SetValidatorSigningInfo(f.ctx, consAddr, info)

	tstaking.CreateValidatorWithValPower(valAddr, val, power, true)
	validatorUpdates, err := f.stakingKeeper.EndBlocker(f.ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(validatorUpdates))
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	// 100 first blocks OK
	height := int64(0)
	for ; height < int64(100); height++ {
		f.ctx = f.ctx.WithBlockHeight(height)
		f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), power, true)
	}

	// kick first validator out of validator set
	tstaking.CreateValidatorWithValPower(sdk.ValAddress(pks[1].Address()), pks[1], power+1, true)
	validatorUpdates, err = f.stakingKeeper.EndBlocker(f.ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(validatorUpdates))
	tstaking.CheckValidator(sdk.ValAddress(pks[1].Address()), stakingtypes.Bonded, false)
	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, false)

	// 600 more blocks happened
	height += 600
	f.ctx = f.ctx.WithBlockHeight(height)

	// validator added back in
	tstaking.DelegateWithPower(sdk.AccAddress(pks[2].Address()), valAddr, 50)

	validatorUpdates, err = f.stakingKeeper.EndBlocker(f.ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(validatorUpdates))
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)
	newPower := power + 50

	// validator misses a block
	f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), newPower, false)
	height++

	// shouldn't be jailed/kicked yet
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	// validator misses an additional 500 more blocks within the SignedBlockWindow (here 1000 blocks).
	latest := f.slashingKeeper.SignedBlocksWindow(f.ctx) + height
	// misses 500 blocks + within the signing windows i.e. 700-1700
	// validators misses all 1000 blocks of a SignedBlockWindows
	for ; height < latest+1; height++ {
		f.slashingKeeper.HandleValidatorSignature(f.ctx.WithBlockHeight(height), val.Address(), newPower, false)
	}

	// should now be jailed & kicked
	f.stakingKeeper.EndBlocker(f.ctx)
	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, true)

	info = slashingtypes.NewValidatorSigningInfo(consAddr, f.ctx.BlockHeight(), int64(0), time.Unix(0, 0), false, int64(0))
	f.slashingKeeper.SetValidatorSigningInfo(f.ctx, consAddr, info)

	// check all the signing information
	signInfo, found := f.slashingKeeper.GetValidatorSigningInfo(f.ctx, consAddr)
	assert.Assert(t, found)
	assert.Equal(t, int64(700), signInfo.StartHeight)
	assert.Equal(t, int64(0), signInfo.MissedBlocksCounter)
	assert.Equal(t, int64(0), signInfo.IndexOffset)

	// some blocks pass
	height = int64(5000)
	f.ctx = f.ctx.WithBlockHeight(height)

	info = slashingtypes.NewValidatorSigningInfo(consAddr, f.ctx.BlockHeight(), int64(0), time.Unix(0, 0), false, int64(0))
	f.slashingKeeper.SetValidatorSigningInfo(f.ctx, consAddr, info)

	// validator rejoins and starts signing again
	f.stakingKeeper.Unjail(f.ctx, consAddr)

	f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), newPower, true)

	// validator should not be kicked since we reset counter/array when it was jailed
	f.stakingKeeper.EndBlocker(f.ctx)
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	// check start height is correctly set
	signInfo, found = f.slashingKeeper.GetValidatorSigningInfo(f.ctx, consAddr)
	assert.Assert(t, found)
	assert.Equal(t, height, signInfo.StartHeight)

	// validator misses 501 blocks after SignedBlockWindow period (1000 blocks)
	latest = f.slashingKeeper.SignedBlocksWindow(f.ctx) + height
	for ; height < latest+f.slashingKeeper.MinSignedPerWindow(f.ctx); height++ {
		f.ctx = f.ctx.WithBlockHeight(height)
		f.slashingKeeper.HandleValidatorSignature(f.ctx, val.Address(), newPower, false)
	}

	// validator should now be jailed & kicked
	f.stakingKeeper.EndBlocker(f.ctx)
	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, true)
}
