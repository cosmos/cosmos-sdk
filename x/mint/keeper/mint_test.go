package keeper_test

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttestutil "github.com/cosmos/cosmos-sdk/x/mint/testutil"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// MintFnTestSuite defines the integration test suite for minting.
type MintFnTestSuite struct {
	suite.Suite

	mintKeeper    keeper.Keeper
	ctx           sdk.Context
	stakingKeeper *minttestutil.MockStakingKeeper
	bankKeeper    *minttestutil.MockBankKeeper
}

// TestMintFnTestSuite runs the mint test suite.
func TestMintFnTestSuite(t *testing.T) {
	suite.Run(t, new(MintFnTestSuite))
}

// SetupTest sets up the context, KV store, and mocks.
func (s *MintFnTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	ctrl := gomock.NewController(s.T())
	accountKeeper := minttestutil.NewMockAccountKeeper(ctrl)
	s.bankKeeper = minttestutil.NewMockBankKeeper(ctrl)
	s.stakingKeeper = minttestutil.NewMockStakingKeeper(ctrl)

	// Return a dummy module address for the mint module.
	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(sdk.AccAddress{}).AnyTimes()

	// Override the default mint function with our dummy inflation calculator.
	s.mintKeeper = keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		s.stakingKeeper,
		accountKeeper,
		s.bankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Set default parameters.
	err := s.mintKeeper.Params.Set(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	// Set a known dummy minter in the store for deterministic behavior.
	s.Require().NoError(s.mintKeeper.Minter.Set(s.ctx, types.DefaultInitialMinter()))
}

// TestDefaultMintFn_Success tests the successful execution of the default mint function.
func (s *MintFnTestSuite) TestDefaultMintFn_Success() {
	// Set the staking keeper expectations.
	stakingSupply := math.NewInt(1_000_000_000)
	bondedRatio := math.LegacyNewDecWithPrec(50, 2) // 0.50
	s.stakingKeeper.EXPECT().StakingTokenSupply(s.ctx).Return(stakingSupply, nil).Times(1)
	s.stakingKeeper.EXPECT().BondedRatio(s.ctx).Return(bondedRatio, nil).Times(1)

	expectedCoins := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(20)))

	minter, err := s.mintKeeper.Minter.Get(s.ctx)
	s.Require().NoError(err)
	expectedInflation := types.DefaultInflationCalculationFn(context.TODO(), minter, types.DefaultParams(), bondedRatio)

	// Set bank keeper expectations for minting and fee collection.
	s.bankKeeper.EXPECT().MintCoins(s.ctx, types.ModuleName, expectedCoins).Return(nil).Times(1)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(s.ctx, types.ModuleName, authtypes.FeeCollectorName, expectedCoins).Return(nil).Times(1)

	// Call the mint function.
	err = s.mintKeeper.MintFn(s.ctx)
	s.Require().NoError(err)

	// Retrieve the updated minter from storage.
	updatedMinter, err := s.mintKeeper.Minter.Get(s.ctx)
	s.Require().NoError(err)

	// check that minter values are updated as expected
	s.Require().Equal(expectedInflation, updatedMinter.Inflation)
	s.Require().Equal(expectedInflation.MulInt(stakingSupply), updatedMinter.AnnualProvisions)

	// Optionally, verify that a mint event has been emitted.
	events := s.ctx.EventManager().Events()
	s.Require().True(slices.ContainsFunc(events, func(event sdk.Event) bool {
		return event.Type == types.EventTypeMint
	}), "expected a mint event to be emitted")
}

// customMintFn defines a custom minting function that overrides minter behavior.
func customMintFn(ctx sdk.Context, k *keeper.Keeper) error {
	// Retrieve the current minter and parameters.
	minter, err := k.Minter.Get(ctx)
	if err != nil {
		return err
	}
	_, err = k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Custom logic: override minter values.
	minter.Inflation = math.LegacyMustNewDecFromStr("0.1")
	minter.AnnualProvisions = math.LegacyMustNewDecFromStr("200")
	if err := k.Minter.Set(ctx, minter); err != nil {
		return err
	}

	// Instead of the default block provision, mint a custom coin.
	mintedCoin := sdk.NewCoin("custom", math.NewInt(50))
	mintedCoins := sdk.NewCoins(mintedCoin)

	// Execute bank keeper methods.
	if err := k.MintCoins(ctx, mintedCoins); err != nil {
		return err
	}
	if err := k.AddCollectedFees(ctx, mintedCoins); err != nil {
		return err
	}

	// Emit a custom event.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"custom_mint",
		sdk.NewAttribute("custom_attribute", "true"),
	))

	return nil
}

// TestCustomMintFn tests the custom mint function.
func (s *MintFnTestSuite) TestCustomMintFn() {
	// Reinitialize the keeper with the custom mint function.
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	s.ctx = testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test")).Ctx

	ctrl := gomock.NewController(s.T())
	accountKeeper := minttestutil.NewMockAccountKeeper(ctrl)
	// Use fresh mocks for account keeper if needed.
	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(sdk.AccAddress{}).AnyTimes()

	// Reuse the existing stakingKeeper and bankKeeper from the suite.
	s.mintKeeper = keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		s.stakingKeeper,
		accountKeeper,
		s.bankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		keeper.WithMintFn(customMintFn),
	)

	// Set default parameters and initial minter.
	err := s.mintKeeper.Params.Set(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
	s.Require().NoError(s.mintKeeper.Minter.Set(s.ctx, types.DefaultInitialMinter()))

	// Expect bank keeper calls to be made for the custom minted coin.
	expectedCoins := sdk.NewCoins(sdk.NewCoin("custom", math.NewInt(50)))
	s.bankKeeper.EXPECT().MintCoins(s.ctx, types.ModuleName, expectedCoins).Return(nil).Times(1)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(s.ctx, types.ModuleName, authtypes.FeeCollectorName, expectedCoins).Return(nil).Times(1)

	// Call the custom mint function.
	err = s.mintKeeper.MintFn(s.ctx)
	s.Require().NoError(err)

	// Retrieve and verify the updated minter values.
	storedMinter, err := s.mintKeeper.Minter.Get(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(math.LegacyMustNewDecFromStr("0.1"), storedMinter.Inflation)
	s.Require().Equal(math.LegacyMustNewDecFromStr("200"), storedMinter.AnnualProvisions)

	// Check that the custom mint event was emitted.
	events := s.ctx.EventManager().Events()
	s.Require().True(slices.ContainsFunc(events, func(event sdk.Event) bool {
		return event.Type == "custom_mint"
	}), "expected custom_mint event to be emitted")
}
