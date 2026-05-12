package keeper_test

import (
	"context"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestBeforeDelegationSharesModified exercises the three reward-withdrawal
// paths the hook can take: fallback to the delegator account, fallback to the
// community pool when the delegator is also blocked, and the strict path
// where a blocked withdraw addr surfaces ErrUnauthorized.
func TestBeforeDelegationSharesModified(t *testing.T) {
	valAddr := sdk.ValAddress(valConsAddr0)
	withdrawAddr := sdk.AccAddress(valConsAddr1)
	owner := sdk.AccAddress(valAddr)

	tests := []struct {
		name string

		strict              bool
		withdrawAddrBlocked bool
		ownerBlocked        bool

		expectedError           error
		expectedWithdrawAddress string
	}{
		{
			name:                    "fallback path, withdraw addr not blocked",
			strict:                  false,
			withdrawAddrBlocked:     false,
			ownerBlocked:            false,
			expectedError:           nil,
			expectedWithdrawAddress: withdrawAddr.String(),
		},
		{
			name:                    "fallback path, withdraw addr blocked, delegator not blocked, redirects to delegator",
			strict:                  false,
			withdrawAddrBlocked:     true,
			ownerBlocked:            false,
			expectedError:           nil,
			expectedWithdrawAddress: owner.String(),
		},
		{
			name:                    "fallback path, withdraw addr blocked, delegator also blocked, redirects to community pool",
			strict:                  false,
			withdrawAddrBlocked:     true,
			ownerBlocked:            true,
			expectedError:           nil,
			expectedWithdrawAddress: disttypes.AttributeValueCommunityPool,
		},
		{
			name:                    "strict path, withdraw addr blocked, ErrUnauthorized",
			strict:                  true,
			withdrawAddrBlocked:     true,
			ownerBlocked:            false, // unread on strict path: resolver errors out before checking fallback
			expectedError:           sdkerrors.ErrUnauthorized,
			expectedWithdrawAddress: "",
		},
		{
			name:                    "strict path, withdraw addr not blocked",
			strict:                  true,
			withdrawAddrBlocked:     false,
			ownerBlocked:            false,
			expectedError:           nil,
			expectedWithdrawAddress: withdrawAddr.String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			key := storetypes.NewKVStoreKey(disttypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
			encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

			bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
			stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
			accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

			accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
			stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()

			distrKeeper := keeper.NewKeeper(
				encCfg.Codec, storeService, accountKeeper, bankKeeper, stakingKeeper,
				"fee_collector", authtypes.NewModuleAddress("gov").String(),
			)

			require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
			require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

			val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
			require.NoError(t, err)
			val.Commission = stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

			del := stakingtypes.NewDelegation(owner.String(), valAddr.String(), val.DelegatorShares)
			stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).AnyTimes()
			stakingKeeper.EXPECT().Delegation(gomock.Any(), owner, valAddr).Return(del, nil).AnyTimes()

			require.NoError(t, distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, owner, valAddr))
			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

			require.NoError(t, distrKeeper.SetDelegatorWithdrawAddr(ctx, owner, withdrawAddr))

			initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
			require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}))

			expRewards := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, initial)}

			bankKeeper.EXPECT().BlockedAddr(withdrawAddr).Return(tc.withdrawAddrBlocked).Times(1)
			switch {
			case !tc.withdrawAddrBlocked:
				// happy path: hook sends rewards directly to configured withdraw addr
				bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), disttypes.ModuleName, withdrawAddr, expRewards).Return(nil).Times(1)
			case !tc.strict:
				// fallback path: check the owner addr next
				bankKeeper.EXPECT().BlockedAddr(owner).Return(tc.ownerBlocked).Times(1)
				if !tc.ownerBlocked {
					bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), disttypes.ModuleName, owner, expRewards).Return(nil).Times(1)
				}
				// else: rewards go to community pool, no bank Send
			}
			// strict + withdrawAddrBlocked: resolver errors immediately, no further bank calls

			var cpBefore disttypes.FeePool
			if tc.expectedWithdrawAddress == disttypes.AttributeValueCommunityPool {
				cpBefore, err = distrKeeper.FeePool.Get(ctx)
				require.NoError(t, err)
			}

			eventCtx := ctx.WithEventManager(sdk.NewEventManager())
			var hookCtx context.Context = eventCtx
			if tc.strict {
				hookCtx = stakingtypes.WithStrictWithdraw(eventCtx)
			}

			err = distrKeeper.Hooks().BeforeDelegationSharesModified(hookCtx, owner, valAddr)
			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)

			// assert there was a redirect event if we did not send funds to the set withdraw address
			if tc.expectedWithdrawAddress != withdrawAddr.String() {
				requireRedirectEvent(t, eventCtx.EventManager(), withdrawAddr.String(), tc.expectedWithdrawAddress)
			}

			// pool-fallback path: pool absorbs the full truncated reward, and
			// outstanding rewards drain to zero.
			if tc.expectedWithdrawAddress == disttypes.AttributeValueCommunityPool {
				cpAfter, err := distrKeeper.FeePool.Get(ctx)
				require.NoError(t, err)
				diff := cpAfter.CommunityPool.Sub(cpBefore.CommunityPool)
				require.True(t,
					diff.AmountOf(sdk.DefaultBondDenom).Equal(math.LegacyNewDecFromInt(expRewards[0].Amount)),
					"community pool delta should equal the truncated reward amount; got %s want %s",
					diff, expRewards[0].Amount,
				)
				outstanding, err := distrKeeper.GetValidatorOutstandingRewardsCoins(ctx, valAddr)
				require.NoError(t, err)
				require.True(t, outstanding.IsZero(), "expected outstanding rewards to be fully drained, got %s", outstanding)
			}
		})
	}
}

func TestAfterValidatorRemoved(t *testing.T) {
	valAddr := sdk.ValAddress(valConsAddr0)
	withdrawAddr := sdk.AccAddress(valConsAddr1)
	owner := sdk.AccAddress(valAddr)

	tests := []struct {
		name string

		strict              bool
		withdrawAddrBlocked bool
		ownerBlocked        bool // only consulted on the fallback path when withdrawAddrBlocked

		expectedError           error
		expectedWithdrawAddress string
	}{
		{
			name:                    "fallback path, withdraw addr not blocked",
			strict:                  false,
			withdrawAddrBlocked:     false,
			ownerBlocked:            false,
			expectedError:           nil,
			expectedWithdrawAddress: withdrawAddr.String(),
		},
		{
			name:                    "fallback path, withdraw addr blocked, owner not blocked, redirects to validator owner",
			strict:                  false,
			withdrawAddrBlocked:     true,
			ownerBlocked:            false,
			expectedError:           nil,
			expectedWithdrawAddress: owner.String(),
		},
		{
			name:                    "fallback path, withdraw addr blocked, owner also blocked, redirects to community pool",
			strict:                  false,
			withdrawAddrBlocked:     true,
			ownerBlocked:            true,
			expectedError:           nil,
			expectedWithdrawAddress: disttypes.AttributeValueCommunityPool,
		},
		{
			name:                    "strict path, withdraw addr not blocked",
			strict:                  true,
			withdrawAddrBlocked:     false,
			ownerBlocked:            false,
			expectedError:           nil,
			expectedWithdrawAddress: withdrawAddr.String(),
		},
		{
			name:                    "strict path, withdraw addr blocked, ErrUnauthorized",
			strict:                  true,
			withdrawAddrBlocked:     true,
			ownerBlocked:            false, // unread on strict path: resolver errors out before checking fallback
			expectedError:           sdkerrors.ErrUnauthorized,
			expectedWithdrawAddress: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			key := storetypes.NewKVStoreKey(disttypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
			encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

			bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
			stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
			accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

			accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
			stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()

			distrKeeper := keeper.NewKeeper(
				encCfg.Codec, storeService, accountKeeper, bankKeeper, stakingKeeper,
				"fee_collector", authtypes.NewModuleAddress("gov").String(),
			)

			require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
			require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

			val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
			require.NoError(t, err)
			// 50% commission rate so commission accrues
			val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

			del := stakingtypes.NewDelegation(owner.String(), valAddr.String(), val.DelegatorShares)
			stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).AnyTimes()
			stakingKeeper.EXPECT().Delegation(gomock.Any(), owner, valAddr).Return(del, nil).AnyTimes()

			require.NoError(t, distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, owner, valAddr))
			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

			require.NoError(t, distrKeeper.SetDelegatorWithdrawAddr(ctx, owner, withdrawAddr))

			initial := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
			require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, initial)}))

			expCommission := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, initial.Quo(math.NewInt(2)))}

			bankKeeper.EXPECT().BlockedAddr(withdrawAddr).Return(tc.withdrawAddrBlocked).Times(1)
			switch {
			case !tc.withdrawAddrBlocked:
				// happy path: hook sends commission directly to configured withdraw addr
				bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), disttypes.ModuleName, withdrawAddr, expCommission).Return(nil).Times(1)
			case !tc.strict:
				// fallback path: check the validator owner addr next
				bankKeeper.EXPECT().BlockedAddr(owner).Return(tc.ownerBlocked).Times(1)
				if !tc.ownerBlocked {
					bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), disttypes.ModuleName, owner, expCommission).Return(nil).Times(1)
				}
				// else: commission goes to community pool, no bank Send
			}
			// strict + withdrawAddrBlocked: resolver errors immediately, no further bank calls

			var cpBefore disttypes.FeePool
			if tc.expectedWithdrawAddress == disttypes.AttributeValueCommunityPool {
				cpBefore, err = distrKeeper.FeePool.Get(ctx)
				require.NoError(t, err)
			}

			eventCtx := ctx.WithEventManager(sdk.NewEventManager())
			var hookCtx context.Context = eventCtx
			if tc.strict {
				hookCtx = stakingtypes.WithStrictWithdraw(eventCtx)
			}

			err = distrKeeper.Hooks().AfterValidatorRemoved(hookCtx, valConsAddr0, valAddr)
			if tc.expectedError != nil {
				// partial keeper-level state mutation isn't asserted; staking
				// already swallows hook errors, and any tx-level rollback
				// covers the user-msg path.
				require.ErrorIs(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)

			// assert there was a redirect event if we did not send funds to the set withdraw address
			if tc.expectedWithdrawAddress != withdrawAddr.String() {
				requireRedirectEvent(t, eventCtx.EventManager(), withdrawAddr.String(), tc.expectedWithdrawAddress)
			}

			// pool-fallback path: pool absorbs the full allocation (commission
			// redirected to pool plus the leftover dust outstanding).
			if tc.expectedWithdrawAddress == disttypes.AttributeValueCommunityPool {
				cpAfter, err := distrKeeper.FeePool.Get(ctx)
				require.NoError(t, err)
				diff := cpAfter.CommunityPool.Sub(cpBefore.CommunityPool)
				require.True(t,
					diff.AmountOf(sdk.DefaultBondDenom).Equal(math.LegacyNewDecFromInt(initial)),
					"community pool delta should equal the full allocation; got %s want %s",
					diff, initial,
				)
			}
		})
	}
}

func requireRedirectEvent(t *testing.T, em *sdk.EventManager, original, final string) {
	t.Helper()
	for _, ev := range em.Events() {
		if ev.Type != disttypes.EventTypeWithdrawAddrRedirected {
			continue
		}
		var foundOriginal, foundFinal bool
		for _, attr := range ev.Attributes {
			if attr.Key == disttypes.AttributeKeyOriginalWithdrawAddress && attr.Value == original {
				foundOriginal = true
			}
			if attr.Key == disttypes.AttributeKeyWithdrawAddress && attr.Value == final {
				foundFinal = true
			}
		}
		if foundOriginal && foundFinal {
			return
		}
	}
	t.Fatalf("expected withdraw_addr_redirected event with original %q and final %q", original, final)
}
