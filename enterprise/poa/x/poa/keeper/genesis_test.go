// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See ./enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package keeper

import (
	"fmt"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
)

func TestExportGenesis(t *testing.T) {
	f := setupTest(t)

	t.Run("when no validators exist", func(t *testing.T) {
		// Set params first
		adminAddr := sdk.AccAddress("admin")
		params := poatypes.Params{
			Admin: adminAddr.String(),
		}
		err := f.poaKeeper.UpdateParams(f.ctx, params)
		require.NoError(t, err)

		// Export genesis
		genesis, err := f.poaKeeper.ExportGenesis(f.ctx)
		require.NoError(t, err)
		require.NotNil(t, genesis)
		require.Equal(t, params.Admin, genesis.Params.Admin)
		require.Empty(t, genesis.Validators)
	})

	t.Run("with validators", func(t *testing.T) {
		// Set params
		adminAddr := sdk.AccAddress("admin")
		params := poatypes.Params{
			Admin: adminAddr.String(),
		}
		err := f.poaKeeper.UpdateParams(f.ctx, params)
		require.NoError(t, err)

		// Create multiple validators
		numValidators := 3
		createdValidators := make([]poatypes.Validator, numValidators)
		consAddrs := make([]sdk.ConsAddress, numValidators)

		for i := 0; i < numValidators; i++ {
			pubKey := ed25519.GenPrivKey().PubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)

			consAddr := sdk.GetConsAddress(pubKey)
			consAddrs[i] = consAddr
			operatorAddr := sdk.AccAddress(fmt.Sprintf("operator%d", i))

			validator := poatypes.Validator{
				PubKey: pubKeyAny,
				Power:  int64((i + 1) * 10), // Powers: 10, 20, 30
				Metadata: &poatypes.ValidatorMetadata{
					Moniker:         fmt.Sprintf("validator-%d", i),
					OperatorAddress: operatorAddr.String(),
				},
			}

			err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
			require.NoError(t, err)
			createdValidators[i] = validator
		}

		// Export genesis
		genesis, err := f.poaKeeper.ExportGenesis(f.ctx)
		require.NoError(t, err)
		require.NotNil(t, genesis)
		require.Equal(t, params.Admin, genesis.Params.Admin)
		require.Len(t, genesis.Validators, numValidators)

		// Verify all validators are exported
		validatorMap := make(map[string]poatypes.Validator)
		for _, v := range genesis.Validators {
			validatorMap[v.Metadata.OperatorAddress] = v
		}

		for i, created := range createdValidators {
			found, exists := validatorMap[created.Metadata.OperatorAddress]
			require.True(t, exists, "validator %d should be exported", i)
			require.Equal(t, created.Power, found.Power)
			require.Equal(t, created.Metadata.Moniker, found.Metadata.Moniker)
			require.Equal(t, created.Metadata.OperatorAddress, found.Metadata.OperatorAddress)
		}
	})

	t.Run("with zero power validator", func(t *testing.T) {
		// Create a validator with power 0
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)
		consAddr := sdk.GetConsAddress(pubKey)
		operatorAddr := sdk.AccAddress("operator-zero")

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  0,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "validator-zero",
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Export genesis
		genesis, err := f.poaKeeper.ExportGenesis(f.ctx)
		require.NoError(t, err)
		require.NotNil(t, genesis)

		// Verify zero power validator is included
		foundZero := false
		for _, v := range genesis.Validators {
			if v.Metadata.OperatorAddress == validator.Metadata.OperatorAddress {
				foundZero = true
				require.Equal(t, int64(0), v.Power)
				break
			}
		}
		require.True(t, foundZero, "validator with power 0 should be exported")
	})
}

func TestInitGenesis(t *testing.T) {
	f := setupTest(t)

	t.Run("with empty genesis state", func(t *testing.T) {
		adminAddr := sdk.AccAddress("admin")
		genesis := &poatypes.GenesisState{
			Params: poatypes.Params{
				Admin: adminAddr.String(),
			},
			Validators: []poatypes.Validator{},
		}

		updates, err := f.poaKeeper.InitGenesis(f.ctx, f.cdc, genesis)
		require.NoError(t, err)
		require.Empty(t, updates)

		// Verify params were set
		params, err := f.poaKeeper.GetParams(f.ctx)
		require.NoError(t, err)
		require.Equal(t, genesis.Params.Admin, params.Admin)

		// Verify no validators exist
		validators, err := f.poaKeeper.GetAllValidators(f.ctx)
		require.NoError(t, err)
		require.Empty(t, validators)

		// Verify total power is 0 (no validators)
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(0), totalPower)
	})

	t.Run("with validators", func(t *testing.T) {
		// Create genesis state with validators
		adminAddr := sdk.AccAddress("admin")
		numValidators := 3
		validators := make([]poatypes.Validator, numValidators)

		for i := 0; i < numValidators; i++ {
			pubKey := ed25519.GenPrivKey().PubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)

			operatorAddr := sdk.AccAddress(fmt.Sprintf("operator%d", i))

			validators[i] = poatypes.Validator{
				PubKey: pubKeyAny,
				Power:  int64((i + 1) * 10), // Powers: 10, 20, 30
				Metadata: &poatypes.ValidatorMetadata{
					Moniker:         fmt.Sprintf("validator-%d", i),
					OperatorAddress: operatorAddr.String(),
				},
			}
		}

		genesis := &poatypes.GenesisState{
			Params: poatypes.Params{
				Admin: adminAddr.String(),
			},
			Validators: validators,
		}

		updates, err := f.poaKeeper.InitGenesis(f.ctx, f.cdc, genesis)
		require.NoError(t, err)
		require.Len(t, updates, numValidators)

		// Verify params were set
		params, err := f.poaKeeper.GetParams(f.ctx)
		require.NoError(t, err)
		require.Equal(t, genesis.Params.Admin, params.Admin)

		// Verify all validators were created
		allValidators, err := f.poaKeeper.GetAllValidators(f.ctx)
		require.NoError(t, err)
		require.Len(t, allValidators, numValidators)

		// Verify validator updates match
		validatorMap := make(map[string]poatypes.Validator)
		for _, v := range allValidators {
			validatorMap[v.Metadata.OperatorAddress] = v
		}

		expectedTotalPower := int64(0)
		for i, expected := range validators {
			found, exists := validatorMap[expected.Metadata.OperatorAddress]
			require.True(t, exists, "validator %d should be created", i)
			require.Equal(t, expected.Power, found.Power)
			require.Equal(t, expected.Metadata.Moniker, found.Metadata.Moniker)

			// Verify validator update
			require.Equal(t, expected.Power, updates[i].Power)

			// Sum up total power (only validators with power > 0)
			if expected.Power > 0 {
				expectedTotalPower += expected.Power
			}
		}

		// Verify total power is set correctly
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, expectedTotalPower, totalPower)
	})

	t.Run("with zero power validator", func(t *testing.T) {
		// Use a fresh fixture to avoid state pollution from previous tests
		f := setupTest(t)

		// Create genesis state with a zero power validator
		adminAddr := sdk.AccAddress("admin-zero")
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)
		operatorAddr := sdk.AccAddress("operator-zero")

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  0,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "validator-zero",
				OperatorAddress: operatorAddr.String(),
				Description:     "Zero power validator description",
			},
		}

		genesis := &poatypes.GenesisState{
			Params: poatypes.Params{
				Admin: adminAddr.String(),
			},
			Validators: []poatypes.Validator{validator},
		}

		updates, err := f.poaKeeper.InitGenesis(f.ctx, f.cdc, genesis)
		require.NoError(t, err)
		// Zero power validators don't generate updates
		require.Empty(t, updates)

		// Verify validator was created
		allValidators, err := f.poaKeeper.GetAllValidators(f.ctx)
		require.NoError(t, err)
		require.Len(t, allValidators, 1)

		foundZero := false
		for _, v := range allValidators {
			if v.Metadata.OperatorAddress == validator.Metadata.OperatorAddress {
				foundZero = true
				require.Equal(t, int64(0), v.Power)
				break
			}
		}
		require.True(t, foundZero, "validator with power 0 should be created")

		// Verify total power is 0 (only zero power validator)
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(0), totalPower)
	})

	t.Run("with validators and allocated fees", func(t *testing.T) {
		// Use a fresh fixture to avoid state pollution from previous tests
		f := setupTest(t)

		// Create genesis state with validators that have allocated fees
		adminAddr := sdk.AccAddress("admin-fees")
		numValidators := 2
		validators := make([]poatypes.Validator, numValidators)

		for i := 0; i < numValidators; i++ {
			pubKey := ed25519.GenPrivKey().PubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)

			operatorAddr := sdk.AccAddress(fmt.Sprintf("operator-fees-%d", i))

			// Create validator with allocated fees
			allocatedFees := sdk.NewDecCoins(
				sdk.NewInt64DecCoin("stake", int64((i+1)*100)),
			)

			validators[i] = poatypes.Validator{
				PubKey:        pubKeyAny,
				Power:         int64((i + 1) * 10), // Powers: 10, 20
				AllocatedFees: allocatedFees,
				Metadata: &poatypes.ValidatorMetadata{
					Moniker:         fmt.Sprintf("validator-fees-%d", i),
					OperatorAddress: operatorAddr.String(),
					Description:     fmt.Sprintf("Validator %d with fees", i),
				},
			}
		}

		genesis := &poatypes.GenesisState{
			Params: poatypes.Params{
				Admin: adminAddr.String(),
			},
			Validators: validators,
		}

		updates, err := f.poaKeeper.InitGenesis(f.ctx, f.cdc, genesis)
		require.NoError(t, err)
		require.Len(t, updates, numValidators)

		// Verify params were set
		params, err := f.poaKeeper.GetParams(f.ctx)
		require.NoError(t, err)
		require.Equal(t, genesis.Params.Admin, params.Admin)

		// Verify all validators were created with allocated fees
		allValidators, err := f.poaKeeper.GetAllValidators(f.ctx)
		require.NoError(t, err)
		require.Len(t, allValidators, numValidators)

		validatorMap := make(map[string]poatypes.Validator)
		for _, v := range allValidators {
			validatorMap[v.Metadata.OperatorAddress] = v
		}

		expectedTotalPower := int64(0)
		for i, expected := range validators {
			found, exists := validatorMap[expected.Metadata.OperatorAddress]
			require.True(t, exists, "validator %d should be created", i)
			require.Equal(t, expected.Power, found.Power)
			require.Equal(t, expected.Metadata.Moniker, found.Metadata.Moniker)

			// Verify allocated fees are preserved
			require.Equal(t, expected.AllocatedFees, found.AllocatedFees, "allocated fees should be preserved for validator %d", i)

			// Verify validator update
			require.Equal(t, expected.Power, updates[i].Power)

			// Sum up total power (only validators with power > 0)
			if expected.Power > 0 {
				expectedTotalPower += expected.Power
			}
		}

		// Verify total power is set correctly
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, expectedTotalPower, totalPower)
	})

	t.Run("export and import round trip", func(t *testing.T) {
		// Create a fresh fixture for export
		f := setupTest(t)

		// Create initial state in export fixture
		adminAddr := sdk.AccAddress("admin-roundtrip")
		params := poatypes.Params{
			Admin: adminAddr.String(),
		}
		err := f.poaKeeper.UpdateParams(f.ctx, params)
		require.NoError(t, err)

		// Create validators with unique addresses
		numValidators := 2
		genesisValidators := make([]poatypes.Validator, numValidators)
		for i := 0; i < numValidators; i++ {
			pubKey := ed25519.GenPrivKey().PubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)
			consAddr := sdk.GetConsAddress(pubKey)
			operatorAddr := sdk.AccAddress(fmt.Sprintf("roundtrip-operator%d", i))

			validator := poatypes.Validator{
				PubKey: pubKeyAny,
				Power:  int64((i + 1) * 10),
				Metadata: &poatypes.ValidatorMetadata{
					Moniker:         fmt.Sprintf("roundtrip-validator-%d", i),
					OperatorAddress: operatorAddr.String(),
				},
			}

			err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
			require.NoError(t, err)
			genesisValidators[i] = validator
		}

		// Export genesis
		exportedGenesis, err := f.poaKeeper.ExportGenesis(f.ctx)
		require.NoError(t, err)
		require.Len(t, exportedGenesis.Validators, numValidators)

		// Create a new keeper instance (simulating a new chain state)
		f = setupTest(t)

		// Import the exported genesis
		updates, err := f.poaKeeper.InitGenesis(f.ctx, f.cdc, exportedGenesis)
		require.NoError(t, err)
		require.Len(t, updates, numValidators)

		// Verify the state matches
		importedParams, err := f.poaKeeper.GetParams(f.ctx)
		require.NoError(t, err)
		require.Equal(t, exportedGenesis.Params.Admin, importedParams.Admin)

		importedValidators, err := f.poaKeeper.GetAllValidators(f.ctx)
		require.NoError(t, err)
		require.Len(t, importedValidators, numValidators)

		// Verify all validators match
		validatorMap := make(map[string]poatypes.Validator)
		for _, v := range importedValidators {
			validatorMap[v.Metadata.OperatorAddress] = v
		}

		expectedTotalPower := int64(0)
		for _, expected := range exportedGenesis.Validators {
			found, exists := validatorMap[expected.Metadata.OperatorAddress]
			require.True(t, exists, "validator should be imported")
			require.Equal(t, expected.Power, found.Power)
			require.Equal(t, expected.Metadata.Moniker, found.Metadata.Moniker)

			// Sum up total power (only validators with power > 0)
			if expected.Power > 0 {
				expectedTotalPower += expected.Power
			}
		}

		// Verify total power is set correctly after import
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, expectedTotalPower, totalPower)
	})

	t.Run("export and import preserves fee collector balances and allocated fees", func(t *testing.T) {
		// Create a fresh fixture for export
		f := setupTest(t)

		// Set up params
		adminAddr := sdk.AccAddress("admin-fees")
		params := poatypes.Params{
			Admin: adminAddr.String(),
		}
		err := f.poaKeeper.UpdateParams(f.ctx, params)
		require.NoError(t, err)

		// Create validators with allocated fees
		numValidators := 5
		validators := make([]poatypes.Validator, numValidators)
		consAddrs := make([]sdk.ConsAddress, numValidators)

		for i := 0; i < numValidators; i++ {
			pubKey := ed25519.GenPrivKey().PubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)
			consAddr := sdk.GetConsAddress(pubKey)
			consAddrs[i] = consAddr
			operatorAddr := sdk.AccAddress(fmt.Sprintf("operator-fees-%d", i))

			// Create validator with allocated fees
			allocatedFees := sdk.NewDecCoins(
				sdk.NewInt64DecCoin("stake", int64((i+1)*100)),
				sdk.NewInt64DecCoin("atom", int64((i+1)*50)),
			)

			validators[i] = poatypes.Validator{
				PubKey:        pubKeyAny,
				Power:         int64((i + 1) * 10), // Powers: 10, 20
				AllocatedFees: allocatedFees,
				Metadata: &poatypes.ValidatorMetadata{
					Moniker:         fmt.Sprintf("validator-fees-%d", i),
					OperatorAddress: operatorAddr.String(),
					Description:     fmt.Sprintf("Validator %d with fees", i),
				},
			}

			err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validators[i], true)
			require.NoError(t, err)
		}

		// Note: In a real scenario, totalAllocatedFees would be set by the keeper
		// when validators are created or fees are allocated. For this test, we'll
		// verify that the sum of validator allocated fees matches what we expect.

		// Export genesis
		exportedGenesis, err := f.poaKeeper.ExportGenesis(f.ctx)
		require.NoError(t, err)
		require.Len(t, exportedGenesis.Validators, numValidators)

		// Verify exported validators have allocated fees preserved
		for i, exportedVal := range exportedGenesis.Validators {
			require.NotNil(t, exportedVal.AllocatedFees, "validator %d should have allocated fees exported", i)
		}

		// Create a new keeper instance (simulating a new chain state)
		f = setupTest(t)

		// Set up fee collector with balances (simulating bank module state restoration)
		// These balances represent fees that were collected but not yet allocated
		feeCollectorFees := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 1000),
			sdk.NewInt64Coin("atom", 500),
		)
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, feeCollectorFees)
		require.NoError(t, err)

		// Import the exported genesis
		updates, err := f.poaKeeper.InitGenesis(f.ctx, f.cdc, exportedGenesis)
		require.NoError(t, err)
		require.Len(t, updates, numValidators)

		// Verify fee collector has the balances
		feeCollectorImport := f.authKeeper.GetModuleAccount(f.ctx, authtypes.FeeCollectorName)
		feeCollectorBalanceImport := f.bankKeeper.GetAllBalances(f.ctx, feeCollectorImport.GetAddress())
		require.Equal(t, feeCollectorFees, feeCollectorBalanceImport, "fee collector balances should be set")

		// Verify all validators were imported with allocated fees
		importedValidators, err := f.poaKeeper.GetAllValidators(f.ctx)
		require.NoError(t, err)
		require.Len(t, importedValidators, numValidators)

		validatorMap := make(map[string]poatypes.Validator)
		for _, v := range importedValidators {
			validatorMap[v.Metadata.OperatorAddress] = v
		}

		// Verify allocated fees are preserved for each validator
		// Since genesis doesn't checkpoint, each validator should have the same
		// allocated fees before and after the export/import cycle
		for i, expected := range exportedGenesis.Validators {
			found, exists := validatorMap[expected.Metadata.OperatorAddress]
			require.True(t, exists, "validator %d should be imported", i)
			require.Equal(t, expected.Power, found.Power)
			require.Equal(t, expected.Metadata.Moniker, found.Metadata.Moniker)

			// Since genesis doesn't checkpoint, allocated fees should be exactly the same
			require.Equal(t, expected.AllocatedFees, found.AllocatedFees,
				"validator %d allocated fees should be preserved during export/import", i)
		}

		// Verify fee collector balance remains unchanged (genesis doesn't distribute fees)
		feeCollectorAfterImport := f.authKeeper.GetModuleAccount(f.ctx, authtypes.FeeCollectorName)
		feeCollectorBalanceAfterImport := f.bankKeeper.GetAllBalances(f.ctx, feeCollectorAfterImport.GetAddress())
		require.Equal(t, feeCollectorFees, feeCollectorBalanceAfterImport,
			"fee collector balance should remain unchanged during genesis import")
	})
}
