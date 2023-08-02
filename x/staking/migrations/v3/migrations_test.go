package v3_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	v3 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v3"
	legacytypes "github.com/cosmos/cosmos-sdk/x/staking/migrations/v3/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// Helper function to write a validator using the old schema
func setLegacyValidator(store sdk.KVStore, cdc codec.BinaryCodec, validator legacytypes.Validator) {
	bz := cdc.MustMarshal(&validator)
	store.Set(types.GetValidatorKey(validator.GetOperator()), bz)
}

// Helper function to write a delegation using the old schema
func setLegacyDelegation(store sdk.KVStore, cdc codec.BinaryCodec, delegation legacytypes.Delegation) {
	delegatorAddress := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)

	bz := cdc.MustMarshal(&delegation)
	store.Set(types.GetDelegationKey(delegatorAddress, delegation.GetValidatorAddr()), bz)
}

// Test setting params in the staking module
func TestMigrateParamsStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig()
	stakingKey := storetypes.NewKVStoreKey(types.ModuleName)
	tStakingKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(stakingKey, tStakingKey)
	paramstore := paramtypes.NewSubspace(cdc.Marshaler, cdc.Amino, stakingKey, tStakingKey, types.ModuleName)

	// Check there are no LSM params
	require.False(t, paramstore.Has(ctx, types.KeyValidatorBondFactor))
	require.False(t, paramstore.Has(ctx, types.KeyGlobalLiquidStakingCap))
	require.False(t, paramstore.Has(ctx, types.KeyValidatorLiquidStakingCap))

	// Run migrations
	v3.MigrateParamsStore(ctx, paramstore)

	// Make sure the new params are set
	require.True(t, paramstore.Has(ctx, types.KeyValidatorBondFactor))
	require.True(t, paramstore.Has(ctx, types.KeyGlobalLiquidStakingCap))
	require.True(t, paramstore.Has(ctx, types.KeyValidatorLiquidStakingCap))

	// Confirm default values are set
	var validatorBondFactor sdk.Dec
	paramstore.Get(ctx, types.KeyValidatorBondFactor, &validatorBondFactor)
	require.Equal(t, types.DefaultValidatorBondFactor, validatorBondFactor)

	var globalLiquidStakingCap sdk.Dec
	paramstore.Get(ctx, types.KeyGlobalLiquidStakingCap, &globalLiquidStakingCap)
	require.Equal(t, types.DefaultGlobalLiquidStakingCap, globalLiquidStakingCap)

	var validatorLiquidStakingCap sdk.Dec
	paramstore.Get(ctx, types.KeyValidatorLiquidStakingCap, &validatorLiquidStakingCap)
	require.Equal(t, types.DefaultValidatorLiquidStakingCap, validatorLiquidStakingCap)
}

// Test setting each validator's TotalValidatorBondShares and TotalLiquidShares to 0
func TestMigrateValidators(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	store := ctx.KVStore(app.GetKey(legacytypes.StoreKey))

	addresses := simapp.AddTestAddrs(app, ctx, 10, sdk.NewInt(1_000_000))
	pubKeys := simapp.CreateTestPubKeys(10)

	// Write each validator with the old type
	oldValidators := []legacytypes.Validator{}
	for i := int64(0); i < 10; i++ {
		valAddress := sdk.ValAddress(addresses[i]).String()
		pkAny, err := codectypes.NewAnyWithValue(pubKeys[i])
		require.NoError(t, err)

		dummyTime := time.Date(2023, 1, 1, 0, 0, int(i), 0, time.UTC)

		description := legacytypes.Description{
			Moniker:         fmt.Sprintf("moniker-%d", i),
			Identity:        fmt.Sprintf("identity-%d", i),
			Website:         fmt.Sprintf("website-%d", i),
			SecurityContact: fmt.Sprintf("security-contact-%d", i),
			Details:         fmt.Sprintf("details-%d", i),
		}

		commission := legacytypes.Commission{
			UpdateTime: dummyTime,
			CommissionRates: legacytypes.CommissionRates{
				Rate:          sdk.NewDec(i),
				MaxRate:       sdk.NewDec(i),
				MaxChangeRate: sdk.NewDec(i),
			},
		}

		validator := legacytypes.Validator{
			OperatorAddress:         valAddress,
			ConsensusPubkey:         pkAny,
			Jailed:                  true,
			Status:                  legacytypes.Bonded,
			Tokens:                  sdk.NewInt(1_000_000),
			DelegatorShares:         sdk.NewDec(1_000_000),
			UnbondingHeight:         i,
			UnbondingTime:           dummyTime,
			MinSelfDelegation:       sdk.NewInt(1_000),
			UnbondingOnHoldRefCount: 1,
			UnbondingIds:            []uint64{uint64(i)},
			Description:             description,
			Commission:              commission,
		}

		oldValidators = append(oldValidators, validator)
		setLegacyValidator(store, app.AppCodec(), validator)
	}

	// Migrate to the new types which adds TotalValidatorBondShares and TotalLiquidShares
	v3.MigrateValidators(ctx, app.StakingKeeper)

	// check that the validator TotalValidatorBondShares and TotalLiquidShares were correctly set to 0
	for _, val := range app.StakingKeeper.GetAllValidators(ctx) {
		require.Equal(t, sdk.ZeroDec(), val.TotalValidatorBondShares)
		require.Equal(t, sdk.ZeroDec(), val.TotalLiquidShares)
	}

	// check that the other validator attributes were unchanged
	for _, oldValidator := range oldValidators {
		newValidator, found := app.StakingKeeper.GetValidator(ctx, oldValidator.GetOperator())
		require.True(t, found)

		require.Equal(t, oldValidator.ConsensusPubkey, newValidator.ConsensusPubkey, "pub key")
		require.Equal(t, oldValidator.Jailed, newValidator.Jailed, "jailed")
		require.Equal(t, oldValidator.Status.String(), newValidator.Status.String(), "status")
		require.Equal(t, oldValidator.Tokens.Int64(), newValidator.Tokens.Int64(), "tokens")
		require.Equal(t, oldValidator.DelegatorShares.TruncateInt64(), newValidator.DelegatorShares.TruncateInt64(), "shares")

		require.Equal(t, oldValidator.UnbondingHeight, newValidator.UnbondingHeight, "unbonding height")
		require.Equal(t, oldValidator.UnbondingTime, newValidator.UnbondingTime, "unbonding time")
		require.Equal(t, oldValidator.UnbondingOnHoldRefCount, newValidator.UnbondingOnHoldRefCount, "unbonding on hold")
		require.Equal(t, oldValidator.UnbondingIds, newValidator.UnbondingIds, "unbonding ids")
		require.Equal(t, oldValidator.MinSelfDelegation.String(), newValidator.MinSelfDelegation, "min self delegation")

		oldDescription := oldValidator.Description
		newDescription := newValidator.Description
		require.Equal(t, oldDescription.Moniker, newDescription.Moniker, "moniker")
		require.Equal(t, oldDescription.Identity, newDescription.Identity, "identity")
		require.Equal(t, oldDescription.Website, newDescription.Website, "website")
		require.Equal(t, oldDescription.SecurityContact, newDescription.SecurityContact, "security contact")
		require.Equal(t, oldDescription.Details, newDescription.Details, "details")

		oldCommissionRate := oldValidator.Commission.CommissionRates
		newCommissionRate := newValidator.Commission.CommissionRates
		require.Equal(t, oldValidator.Commission.UpdateTime, newValidator.Commission.UpdateTime, "commission update time")
		require.Equal(t, oldCommissionRate.Rate, newCommissionRate.Rate, "commission rate")
		require.Equal(t, oldCommissionRate.MaxRate, newCommissionRate.MaxRate, "commission max rate")
		require.Equal(t, oldCommissionRate.MaxChangeRate, newCommissionRate.MaxChangeRate, "commission max rate change")
	}
}

// Test setting each delegation's validator bond to false
func TestMigrateDelegations(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	store := ctx.KVStore(app.GetKey(legacytypes.StoreKey))

	validatorAddresses := simapp.AddTestAddrs(app, ctx, 10, sdk.NewInt(1_000_000))
	delegatorAddresses := simapp.AddTestAddrs(app, ctx, 10, sdk.NewInt(1_000_000))

	// Write each delegation with the old type
	oldDelegations := []legacytypes.Delegation{}
	for i := int64(0); i < 10; i++ {
		delegation := legacytypes.Delegation{
			DelegatorAddress: delegatorAddresses[i].String(),
			ValidatorAddress: sdk.ValAddress(validatorAddresses[i]).String(),
			Shares:           sdk.NewDec(i * 1000),
		}

		oldDelegations = append(oldDelegations, delegation)
		setLegacyDelegation(store, app.AppCodec(), delegation)
	}

	// Migrate the new delegations which should add the ValidatorBond field
	v3.MigrateDelegations(ctx, app.StakingKeeper)

	// check that the delegation is not a validator bond
	for _, delegation := range app.StakingKeeper.GetAllDelegations(ctx) {
		require.Equal(t, false, delegation.ValidatorBond)
	}

	// check that the other delegation attributes were unchanged
	for _, oldDelegation := range oldDelegations {
		newDelegation, found := app.StakingKeeper.GetDelegation(ctx, oldDelegation.GetDelegatorAddr(), oldDelegation.GetValidatorAddr())
		require.True(t, found)

		require.Equal(t, oldDelegation.DelegatorAddress, newDelegation.DelegatorAddress, "delegator address")
		require.Equal(t, oldDelegation.ValidatorAddress, newDelegation.ValidatorAddress, "validator address")
		require.Equal(t, oldDelegation.Shares.TruncateInt64(), newDelegation.Shares.TruncateInt64(), "shares")
	}
}
