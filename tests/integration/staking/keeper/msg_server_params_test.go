package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestMsgUpdateParams_BondDenomValidation tests that MsgUpdateParams properly
// validates that the bond_denom exists on-chain before allowing the update.
// This is an integration test using real keepers (not mocks).
func TestMsgUpdateParams_BondDenomValidation(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)
	authority := authtypes.NewModuleAddress("gov").String()

	testCases := []struct {
		name      string
		setupFunc func()
		bondDenom string
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid bond denom with existing supply",
			setupFunc: func() {
				// Mint coins to create supply for "validcoin"
				mintCoins := sdk.NewCoins(sdk.NewInt64Coin("validcoin", 1000000))
				require.NoError(t, testutil.FundModuleAccount(ctx, f.bankKeeper, types.ModuleName, mintCoins))
			},
			bondDenom: "validcoin",
			expectErr: false,
		},
		{
			name: "invalid bond denom with zero supply",
			setupFunc: func() {
				// Don't mint anything for "ghosttoken" - it will have zero supply
			},
			bondDenom: "ghosttoken",
			expectErr: true,
			errMsg:    "does not exist or has zero supply",
		},
		{
			name: "valid default bond denom (stake)",
			setupFunc: func() {
				// The default bond denom "stake" should already have supply from test setup
				// Just ensure it has supply
				currentParams, err := f.stakingKeeper.GetParams(ctx)
				require.NoError(t, err)
				if currentParams.BondDenom == "stake" {
					// Already has supply from initialization
					return
				}
				// Otherwise mint some
				mintCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000000))
				require.NoError(t, testutil.FundModuleAccount(ctx, f.bankKeeper, types.ModuleName, mintCoins))
			},
			bondDenom: "stake",
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run setup if provided
			if tc.setupFunc != nil {
				tc.setupFunc()
			}

			// Create params with the test bond denom
			params := types.DefaultParams()
			params.BondDenom = tc.bondDenom

			// Create update params message
			msg := &types.MsgUpdateParams{
				Authority: authority,
				Params:    params,
			}

			// Execute the message
			_, err := msgServer.UpdateParams(ctx, msg)

			// Check expectations
			if tc.expectErr {
				require.Error(t, err, "expected error but got none")
				require.ErrorContains(t, err, tc.errMsg)
				require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
			} else {
				require.NoError(t, err)

				// Verify params were actually updated
				updatedParams, err := f.stakingKeeper.GetParams(ctx)
				require.NoError(t, err)
				require.Equal(t, tc.bondDenom, updatedParams.BondDenom)
			}
		})
	}
}

// TestMsgUpdateParams_OtherValidations ensures other parameter validations still work
func TestMsgUpdateParams_OtherValidations(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)
	authority := authtypes.NewModuleAddress("gov").String()

	// Ensure default bond denom has supply
	currentParams, err := f.stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	mintCoins := sdk.NewCoins(sdk.NewInt64Coin(currentParams.BondDenom, 1000000))
	require.NoError(t, testutil.FundModuleAccount(ctx, f.bankKeeper, types.ModuleName, mintCoins))

	testCases := []struct {
		name      string
		params    types.Params
		expectErr bool
		errMsg    string
	}{
		{
			name: "invalid authority",
			params: types.Params{
				MinCommissionRate: types.DefaultMinCommissionRate,
				UnbondingTime:     types.DefaultUnbondingTime,
				MaxValidators:     types.DefaultMaxValidators,
				MaxEntries:        types.DefaultMaxEntries,
				HistoricalEntries: types.DefaultHistoricalEntries,
				BondDenom:         currentParams.BondDenom,
			},
			expectErr: true,
			errMsg:    "invalid authority",
		},
		{
			name: "blank bond denom",
			params: types.Params{
				MinCommissionRate: types.DefaultMinCommissionRate,
				UnbondingTime:     types.DefaultUnbondingTime,
				MaxValidators:     types.DefaultMaxValidators,
				MaxEntries:        types.DefaultMaxEntries,
				HistoricalEntries: types.DefaultHistoricalEntries,
				BondDenom:         "",
			},
			expectErr: true,
			errMsg:    "bond denom cannot be blank",
		},
		{
			name: "negative commission rate",
			params: types.Params{
				MinCommissionRate: math.LegacyNewDec(-1),
				UnbondingTime:     types.DefaultUnbondingTime,
				MaxValidators:     types.DefaultMaxValidators,
				MaxEntries:        types.DefaultMaxEntries,
				HistoricalEntries: types.DefaultHistoricalEntries,
				BondDenom:         currentParams.BondDenom,
			},
			expectErr: true,
			errMsg:    "minimum commission rate cannot be negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use wrong authority for the first test case
			auth := authority
			if tc.name == "invalid authority" {
				auth = "invalid"
			}

			msg := &types.MsgUpdateParams{
				Authority: auth,
				Params:    tc.params,
			}

			_, err := msgServer.UpdateParams(ctx, msg)

			if tc.expectErr {
				require.Error(t, err, "expected error but got none")
				require.ErrorContains(t, err, tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
