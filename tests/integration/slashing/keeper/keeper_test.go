package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// The default power validators are initialized to have within tests
var InitTokens = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)

type fixture struct {
	ctx               sdk.Context
	slashingKeeper    slashingkeeper.Keeper
	stakingKeeper     *stakingkeeper.Keeper
	bankKeeper        bankkeeper.Keeper
	accountKeeper     authkeeper.AccountKeeper
	interfaceRegistry codectypes.InterfaceRegistry
	addrDels          []sdk.AccAddress
	queryClient       slashingtypes.QueryClient
	msgServer         slashingtypes.MsgServer
}

func initFixture(t assert.TestingT) *fixture {
	f := &fixture{}
	app, err := simtestutil.Setup(
		testutil.AppConfig,
		&f.bankKeeper,
		&f.accountKeeper,
		&f.slashingKeeper,
		&f.stakingKeeper,
		&f.interfaceRegistry,
	)
	assert.NilError(t, err)

	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})

	// TestParams set the SignedBlocksWindow to 1000 and MaxMissedBlocksPerWindow to 500
	f.slashingKeeper.SetParams(ctx, testutil.TestParams())
	addrDels := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, ctx, 5, f.stakingKeeper.TokensFromConsensusPower(ctx, 200))

	info1 := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[0]), int64(4), int64(3),
		time.Unix(2, 0), false, int64(10))
	info2 := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[1]), int64(5), int64(4),
		time.Unix(2, 0), false, int64(10))

	f.slashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]), info1)
	f.slashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[1]), info2)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, f.interfaceRegistry)
	slashingtypes.RegisterQueryServer(queryHelper, f.slashingKeeper)
	queryClient := slashingtypes.NewQueryClient(queryHelper)
	f.queryClient = queryClient

	f.addrDels = addrDels
	f.ctx = ctx
	f.msgServer = slashingkeeper.NewMsgServerImpl(f.slashingKeeper)

	return f
}

func TestUnJailNotBonded(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.ctx

	p := f.stakingKeeper.GetParams(ctx)
	p.MaxValidators = 5
	f.stakingKeeper.SetParams(ctx, p)

	addrDels := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, ctx, 6, f.stakingKeeper.TokensFromConsensusPower(ctx, 200))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrDels)
	pks := simtestutil.CreateTestPubKeys(6)
	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)

	// create max (5) validators all with the same power
	for i := uint32(0); i < p.MaxValidators; i++ {
		addr, val := valAddrs[i], pks[i]
		tstaking.CreateValidatorWithValPower(addr, val, 100, true)
	}

	staking.EndBlocker(ctx, f.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// create a 6th validator with less power than the cliff validator (won't be bonded)
	addr, val := valAddrs[5], pks[5]
	amt := f.stakingKeeper.TokensFromConsensusPower(ctx, 50)
	msg := tstaking.CreateValidatorMsg(addr, val, amt)
	msg.MinSelfDelegation = amt
	res, err := tstaking.CreateValidatorWithMsg(ctx, msg)
	assert.NilError(t, err)
	assert.Assert(t, res != nil)

	staking.EndBlocker(ctx, f.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	tstaking.CheckValidator(addr, stakingtypes.Unbonded, false)

	// unbond below minimum self-delegation
	assert.Equal(t, p.BondDenom, tstaking.Denom)
	tstaking.Undelegate(sdk.AccAddress(addr), addr, f.stakingKeeper.TokensFromConsensusPower(ctx, 1), true)

	staking.EndBlocker(ctx, f.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// verify that validator is jailed
	tstaking.CheckValidator(addr, -1, true)

	// verify we cannot unjail (yet)
	assert.ErrorContains(t, f.slashingKeeper.Unjail(ctx, addr), "cannot be unjailed")

	staking.EndBlocker(ctx, f.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// bond to meet minimum self-delegation
	tstaking.DelegateWithPower(sdk.AccAddress(addr), addr, 1)

	staking.EndBlocker(ctx, f.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// verify we can immediately unjail
	assert.NilError(t, f.slashingKeeper.Unjail(ctx, addr))

	tstaking.CheckValidator(addr, -1, false)
}

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately jailed
func TestHandleNewValidator(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.ctx

	addrDels := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, ctx, 1, f.stakingKeeper.TokensFromConsensusPower(ctx, 0))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrDels)
	pks := simtestutil.CreateTestPubKeys(1)
	addr, val := valAddrs[0], pks[0]
	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)
	ctx = ctx.WithBlockHeight(f.slashingKeeper.SignedBlocksWindow(ctx) + 1)

	// Validator created
	amt := tstaking.CreateValidatorWithValPower(addr, val, 100, true)

	staking.EndBlocker(ctx, f.stakingKeeper)
	assert.DeepEqual(
		t, f.bankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(f.stakingKeeper.GetParams(ctx).BondDenom, InitTokens.Sub(amt))),
	)
	assert.DeepEqual(t, amt, f.stakingKeeper.Validator(ctx, addr).GetBondedTokens())

	// Now a validator, for two blocks
	f.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), 100, true)
	ctx = ctx.WithBlockHeight(f.slashingKeeper.SignedBlocksWindow(ctx) + 2)
	f.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), 100, false)

	info, found := f.slashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	assert.Assert(t, found)
	assert.Equal(t, f.slashingKeeper.SignedBlocksWindow(ctx)+1, info.StartHeight)
	assert.Equal(t, int64(2), info.IndexOffset)
	assert.Equal(t, int64(1), info.MissedBlocksCounter)
	assert.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)

	// validator should be bonded still, should not have been jailed or slashed
	validator, _ := f.stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	assert.Equal(t, stakingtypes.Bonded, validator.GetStatus())
	bondPool := f.stakingKeeper.GetBondedPool(ctx)
	expTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 100)
	// adding genesis validator tokens
	expTokens = expTokens.Add(f.stakingKeeper.TokensFromConsensusPower(ctx, 1))
	assert.Assert(t, expTokens.Equal(f.bankKeeper.GetBalance(ctx, bondPool.GetAddress(), f.stakingKeeper.BondDenom(ctx)).Amount))
}

// Test a jailed validator being "down" twice
// Ensure that they're only slashed once
func TestHandleAlreadyJailed(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// initial setup
	ctx := f.ctx

	addrDels := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, ctx, 1, f.stakingKeeper.TokensFromConsensusPower(ctx, 200))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrDels)
	pks := simtestutil.CreateTestPubKeys(1)
	addr, val := valAddrs[0], pks[0]
	power := int64(100)
	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)

	amt := tstaking.CreateValidatorWithValPower(addr, val, power, true)

	staking.EndBlocker(ctx, f.stakingKeeper)

	// 1000 first blocks OK
	height := int64(0)
	for ; height < f.slashingKeeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		f.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, true)
	}

	// 501 blocks missed
	for ; height < f.slashingKeeper.SignedBlocksWindow(ctx)+(f.slashingKeeper.SignedBlocksWindow(ctx)-f.slashingKeeper.MinSignedPerWindow(ctx))+1; height++ {
		ctx = ctx.WithBlockHeight(height)
		f.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, f.stakingKeeper)

	// validator should have been jailed and slashed
	validator, _ := f.stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	assert.Equal(t, stakingtypes.Unbonding, validator.GetStatus())

	// validator should have been slashed
	resultingTokens := amt.Sub(f.stakingKeeper.TokensFromConsensusPower(ctx, 1))
	assert.DeepEqual(t, resultingTokens, validator.GetTokens())

	// another block missed
	ctx = ctx.WithBlockHeight(height)
	f.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)

	// validator should not have been slashed twice
	validator, _ = f.stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	assert.DeepEqual(t, resultingTokens, validator.GetTokens())
}

// Test a validator dipping in and out of the validator set
// Ensure that missed blocks are tracked correctly and that
// the start height of the signing info is reset correctly
func TestValidatorDippingInAndOut(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// initial setup
	ctx := f.ctx
	params := f.stakingKeeper.GetParams(ctx)
	params.MaxValidators = 1
	f.stakingKeeper.SetParams(ctx, params)
	power := int64(100)

	pks := simtestutil.CreateTestPubKeys(3)
	simtestutil.AddTestAddrsFromPubKeys(f.bankKeeper, f.stakingKeeper, ctx, pks, f.stakingKeeper.TokensFromConsensusPower(ctx, 200))

	addr, val := pks[0].Address(), pks[0]
	consAddr := sdk.ConsAddress(addr)
	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)
	valAddr := sdk.ValAddress(addr)

	tstaking.CreateValidatorWithValPower(valAddr, val, power, true)
	validatorUpdates := staking.EndBlocker(ctx, f.stakingKeeper)
	assert.Equal(t, 2, len(validatorUpdates))
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	// 100 first blocks OK
	height := int64(0)
	for ; height < int64(100); height++ {
		ctx = ctx.WithBlockHeight(height)
		f.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, true)
	}

	// kick first validator out of validator set
	tstaking.CreateValidatorWithValPower(sdk.ValAddress(pks[1].Address()), pks[1], power+1, true)
	validatorUpdates = staking.EndBlocker(ctx, f.stakingKeeper)
	assert.Equal(t, 2, len(validatorUpdates))
	tstaking.CheckValidator(sdk.ValAddress(pks[1].Address()), stakingtypes.Bonded, false)
	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, false)

	// 600 more blocks happened
	height = height + 600
	ctx = ctx.WithBlockHeight(height)

	// validator added back in
	tstaking.DelegateWithPower(sdk.AccAddress(pks[2].Address()), valAddr, 50)

	validatorUpdates = staking.EndBlocker(ctx, f.stakingKeeper)
	assert.Equal(t, 2, len(validatorUpdates))
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)
	newPower := power + 50

	// validator misses a block
	f.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), newPower, false)
	height++

	// shouldn't be jailed/kicked yet
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	// validator misses an additional 500 more blocks within the SignedBlockWindow (here 1000 blocks).
	latest := f.slashingKeeper.SignedBlocksWindow(ctx) + height
	// misses 500 blocks + within the signing windows i.e. 700-1700
	// validators misses all 1000 block of a SignedBlockWindows
	for ; height < latest+1; height++ {
		ctx = ctx.WithBlockHeight(height)
		f.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), newPower, false)
	}

	// should now be jailed & kicked
	staking.EndBlocker(ctx, f.stakingKeeper)
	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, true)

	// check all the signing information
	signInfo, found := f.slashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
	assert.Assert(t, found)
	assert.Equal(t, int64(700), signInfo.StartHeight)
	assert.Equal(t, int64(0), signInfo.MissedBlocksCounter)
	assert.Equal(t, int64(0), signInfo.IndexOffset)

	// some blocks pass
	height = int64(5000)
	ctx = ctx.WithBlockHeight(height)

	// validator rejoins and starts signing again
	f.stakingKeeper.Unjail(ctx, consAddr)

	f.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), newPower, true)

	// validator should not be kicked since we reset counter/array when it was jailed
	staking.EndBlocker(ctx, f.stakingKeeper)
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	// check start height is correctly set
	signInfo, found = f.slashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
	assert.Assert(t, found)
	assert.Equal(t, height, signInfo.StartHeight)

	// validator misses 501 blocks after SignedBlockWindow period (1000 blocks)
	latest = f.slashingKeeper.SignedBlocksWindow(ctx) + height
	for ; height < latest+f.slashingKeeper.MinSignedPerWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		f.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), newPower, false)
	}

	// validator should now be jailed & kicked
	staking.EndBlocker(ctx, f.stakingKeeper)
	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, true)
}
