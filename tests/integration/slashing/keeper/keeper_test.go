package keeper_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

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

type KeeperTestSuite struct {
	suite.Suite

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

func (s *KeeperTestSuite) SetupTest() {
	app, err := simtestutil.Setup(
		testutil.AppConfig,
		&s.bankKeeper,
		&s.accountKeeper,
		&s.slashingKeeper,
		&s.stakingKeeper,
		&s.interfaceRegistry,
	)
	s.Require().NoError(err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// TestParams set the SignedBlocksWindow to 1000 and MaxMissedBlocksPerWindow to 500
	s.slashingKeeper.SetParams(ctx, testutil.TestParams())
	addrDels := simtestutil.AddTestAddrsIncremental(s.bankKeeper, s.stakingKeeper, ctx, 5, s.stakingKeeper.TokensFromConsensusPower(ctx, 200))

	info1 := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[0]), int64(4), int64(3),
		time.Unix(2, 0), false, int64(10))
	info2 := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[1]), int64(5), int64(4),
		time.Unix(2, 0), false, int64(10))

	s.slashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]), info1)
	s.slashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[1]), info2)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, s.interfaceRegistry)
	slashingtypes.RegisterQueryServer(queryHelper, s.slashingKeeper)
	queryClient := slashingtypes.NewQueryClient(queryHelper)
	s.queryClient = queryClient

	s.addrDels = addrDels
	s.ctx = ctx
	s.msgServer = slashingkeeper.NewMsgServerImpl(s.slashingKeeper)
}

func (s *KeeperTestSuite) TestUnJailNotBonded() {
	ctx := s.ctx

	p := s.stakingKeeper.GetParams(ctx)
	p.MaxValidators = 5
	s.stakingKeeper.SetParams(ctx, p)

	addrDels := simtestutil.AddTestAddrsIncremental(s.bankKeeper, s.stakingKeeper, ctx, 6, s.stakingKeeper.TokensFromConsensusPower(ctx, 200))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrDels)
	pks := simtestutil.CreateTestPubKeys(6)
	tstaking := stakingtestutil.NewHelper(s.T(), ctx, s.stakingKeeper)

	// create max (5) validators all with the same power
	for i := uint32(0); i < p.MaxValidators; i++ {
		addr, val := valAddrs[i], pks[i]
		tstaking.CreateValidatorWithValPower(addr, val, 100, true)
	}

	staking.EndBlocker(ctx, s.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// create a 6th validator with less power than the cliff validator (won't be bonded)
	addr, val := valAddrs[5], pks[5]
	amt := s.stakingKeeper.TokensFromConsensusPower(ctx, 50)
	msg := tstaking.CreateValidatorMsg(addr, val, amt)
	msg.MinSelfDelegation = amt
	res, err := tstaking.CreateValidatorWithMsg(sdk.WrapSDKContext(ctx), msg)
	s.Require().NoError(err)
	s.Require().NotNil(res)

	staking.EndBlocker(ctx, s.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	tstaking.CheckValidator(addr, stakingtypes.Unbonded, false)

	// unbond below minimum self-delegation
	s.Require().Equal(p.BondDenom, tstaking.Denom)
	tstaking.Undelegate(sdk.AccAddress(addr), addr, s.stakingKeeper.TokensFromConsensusPower(ctx, 1), true)

	staking.EndBlocker(ctx, s.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// verify that validator is jailed
	tstaking.CheckValidator(addr, -1, true)

	// verify we cannot unjail (yet)
	s.Require().Error(s.slashingKeeper.Unjail(ctx, addr))

	staking.EndBlocker(ctx, s.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// bond to meet minimum self-delegation
	tstaking.DelegateWithPower(sdk.AccAddress(addr), addr, 1)

	staking.EndBlocker(ctx, s.stakingKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// verify we can immediately unjail
	s.Require().NoError(s.slashingKeeper.Unjail(ctx, addr))

	tstaking.CheckValidator(addr, -1, false)
}

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately jailed
func (s *KeeperTestSuite) TestHandleNewValidator() {
	ctx := s.ctx

	addrDels := simtestutil.AddTestAddrsIncremental(s.bankKeeper, s.stakingKeeper, ctx, 1, s.stakingKeeper.TokensFromConsensusPower(ctx, 0))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrDels)
	pks := simtestutil.CreateTestPubKeys(1)
	addr, val := valAddrs[0], pks[0]
	tstaking := stakingtestutil.NewHelper(s.T(), ctx, s.stakingKeeper)
	ctx = ctx.WithBlockHeight(s.slashingKeeper.SignedBlocksWindow(ctx) + 1)

	// Validator created
	amt := tstaking.CreateValidatorWithValPower(addr, val, 100, true)

	staking.EndBlocker(ctx, s.stakingKeeper)
	s.Require().Equal(
		s.bankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(s.stakingKeeper.GetParams(ctx).BondDenom, InitTokens.Sub(amt))),
	)
	s.Require().Equal(amt, s.stakingKeeper.Validator(ctx, addr).GetBondedTokens())

	// Now a validator, for two blocks
	s.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), 100, true)
	ctx = ctx.WithBlockHeight(s.slashingKeeper.SignedBlocksWindow(ctx) + 2)
	s.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), 100, false)

	info, found := s.slashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	s.Require().True(found)
	s.Require().Equal(s.slashingKeeper.SignedBlocksWindow(ctx)+1, info.StartHeight)
	s.Require().Equal(int64(2), info.IndexOffset)
	s.Require().Equal(int64(1), info.MissedBlocksCounter)
	s.Require().Equal(time.Unix(0, 0).UTC(), info.JailedUntil)

	// validator should be bonded still, should not have been jailed or slashed
	validator, _ := s.stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	s.Require().Equal(stakingtypes.Bonded, validator.GetStatus())
	bondPool := s.stakingKeeper.GetBondedPool(ctx)
	expTokens := s.stakingKeeper.TokensFromConsensusPower(ctx, 100)
	// adding genesis validator tokens
	expTokens = expTokens.Add(s.stakingKeeper.TokensFromConsensusPower(ctx, 1))
	s.Require().True(expTokens.Equal(s.bankKeeper.GetBalance(ctx, bondPool.GetAddress(), s.stakingKeeper.BondDenom(ctx)).Amount))
}

// Test a jailed validator being "down" twice
// Ensure that they're only slashed once
func (s *KeeperTestSuite) TestHandleAlreadyJailed() {
	// initial setup
	ctx := s.ctx

	addrDels := simtestutil.AddTestAddrsIncremental(s.bankKeeper, s.stakingKeeper, ctx, 1, s.stakingKeeper.TokensFromConsensusPower(ctx, 200))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrDels)
	pks := simtestutil.CreateTestPubKeys(1)
	addr, val := valAddrs[0], pks[0]
	power := int64(100)
	tstaking := stakingtestutil.NewHelper(s.T(), ctx, s.stakingKeeper)

	amt := tstaking.CreateValidatorWithValPower(addr, val, power, true)

	staking.EndBlocker(ctx, s.stakingKeeper)

	// 1000 first blocks OK
	height := int64(0)
	for ; height < s.slashingKeeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		s.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, true)
	}

	// 501 blocks missed
	for ; height < s.slashingKeeper.SignedBlocksWindow(ctx)+(s.slashingKeeper.SignedBlocksWindow(ctx)-s.slashingKeeper.MinSignedPerWindow(ctx))+1; height++ {
		ctx = ctx.WithBlockHeight(height)
		s.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, s.stakingKeeper)

	// validator should have been jailed and slashed
	validator, _ := s.stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	s.Require().Equal(stakingtypes.Unbonding, validator.GetStatus())

	// validator should have been slashed
	resultingTokens := amt.Sub(s.stakingKeeper.TokensFromConsensusPower(ctx, 1))
	s.Require().Equal(resultingTokens, validator.GetTokens())

	// another block missed
	ctx = ctx.WithBlockHeight(height)
	s.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)

	// validator should not have been slashed twice
	validator, _ = s.stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	s.Require().Equal(resultingTokens, validator.GetTokens())
}

// Test a validator dipping in and out of the validator set
// Ensure that missed blocks are tracked correctly and that
// the start height of the signing info is reset correctly
func (s *KeeperTestSuite) TestValidatorDippingInAndOut() {
	// initial setup
	ctx := s.ctx
	params := s.stakingKeeper.GetParams(ctx)
	params.MaxValidators = 1
	s.stakingKeeper.SetParams(ctx, params)
	power := int64(100)

	pks := simtestutil.CreateTestPubKeys(3)
	simtestutil.AddTestAddrsFromPubKeys(s.bankKeeper, s.stakingKeeper, ctx, pks, s.stakingKeeper.TokensFromConsensusPower(ctx, 200))

	addr, val := pks[0].Address(), pks[0]
	consAddr := sdk.ConsAddress(addr)
	tstaking := stakingtestutil.NewHelper(s.T(), ctx, s.stakingKeeper)
	valAddr := sdk.ValAddress(addr)

	tstaking.CreateValidatorWithValPower(valAddr, val, power, true)
	validatorUpdates := staking.EndBlocker(ctx, s.stakingKeeper)
	s.Require().Equal(2, len(validatorUpdates))
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	// 100 first blocks OK
	height := int64(0)
	for ; height < int64(100); height++ {
		ctx = ctx.WithBlockHeight(height)
		s.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, true)
	}

	// kick first validator out of validator set
	tstaking.CreateValidatorWithValPower(sdk.ValAddress(pks[1].Address()), pks[1], power+1, true)
	validatorUpdates = staking.EndBlocker(ctx, s.stakingKeeper)
	s.Require().Equal(2, len(validatorUpdates))
	tstaking.CheckValidator(sdk.ValAddress(pks[1].Address()), stakingtypes.Bonded, false)
	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, false)

	// 600 more blocks happened
	height = height + 600
	ctx = ctx.WithBlockHeight(height)

	// validator added back in
	tstaking.DelegateWithPower(sdk.AccAddress(pks[2].Address()), valAddr, 50)

	validatorUpdates = staking.EndBlocker(ctx, s.stakingKeeper)
	s.Require().Equal(2, len(validatorUpdates))
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)
	newPower := power + 50

	// validator misses a block
	s.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), newPower, false)
	height++

	// shouldn't be jailed/kicked yet
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	// validator misses an additional 500 more blocks within the SignedBlockWindow (here 1000 blocks).
	latest := s.slashingKeeper.SignedBlocksWindow(ctx) + height
	// misses 500 blocks + within the signing windows i.e. 700-1700
	// validators misses all 1000 block of a SignedBlockWindows
	for ; height < latest+1; height++ {
		ctx = ctx.WithBlockHeight(height)
		s.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), newPower, false)
	}

	// should now be jailed & kicked
	staking.EndBlocker(ctx, s.stakingKeeper)
	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, true)

	// check all the signing information
	signInfo, found := s.slashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
	s.Require().True(found)
	s.Require().Equal(int64(700), signInfo.StartHeight)
	s.Require().Equal(int64(0), signInfo.MissedBlocksCounter)
	s.Require().Equal(int64(0), signInfo.IndexOffset)

	// some blocks pass
	height = int64(5000)
	ctx = ctx.WithBlockHeight(height)

	// validator rejoins and starts signing again
	s.stakingKeeper.Unjail(ctx, consAddr)

	s.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), newPower, true)

	// validator should not be kicked since we reset counter/array when it was jailed
	staking.EndBlocker(ctx, s.stakingKeeper)
	tstaking.CheckValidator(valAddr, stakingtypes.Bonded, false)

	// check start height is correctly set
	signInfo, found = s.slashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
	s.Require().True(found)
	s.Require().Equal(height, signInfo.StartHeight)

	// validator misses 501 blocks after SignedBlockWindow period (1000 blocks)
	latest = s.slashingKeeper.SignedBlocksWindow(ctx) + height
	for ; height < latest+s.slashingKeeper.MinSignedPerWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		s.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), newPower, false)
	}

	// validator should now be jailed & kicked
	staking.EndBlocker(ctx, s.stakingKeeper)
	tstaking.CheckValidator(valAddr, stakingtypes.Unbonding, true)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
