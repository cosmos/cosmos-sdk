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
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package keeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestCreateValidator(t *testing.T) {
	f := setupTest(t)

	// Create validator with power 0
	pubKey := ed25519.GenPrivKey().PubKey()
	pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	consAddr := sdk.GetConsAddress(pubKey)
	operatorAddr := sdk.AccAddress("operator1")

	validator := types.Validator{
		PubKey: pubKeyAny,
		Power:  0,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "test-validator",
			OperatorAddress: operatorAddr.String(),
		},
	}

	err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
	require.NoError(t, err)

	// Verify validator exists and can be retrieved
	exists, err := f.poaKeeper.HasValidator(f.ctx, consAddr)
	require.NoError(t, err)
	require.True(t, exists)

	retrieved, err := f.poaKeeper.GetValidator(f.ctx, consAddr)
	require.NoError(t, err)
	require.Equal(t, validator.Power, retrieved.Power)
	require.Equal(t, validator.Metadata.Moniker, retrieved.Metadata.Moniker)
	require.Equal(t, validator.Metadata.OperatorAddress, retrieved.Metadata.OperatorAddress)

	// Test consensus address uniqueness - cannot create validator with same cons address
	validator2 := types.Validator{
		PubKey: pubKeyAny,
		Power:  100,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "different-validator",
			OperatorAddress: sdk.AccAddress("operator2").String(),
		},
	}

	err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator2, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot create duplicate validator with same consensus address")

	// Test operator address uniqueness - cannot create validator with same operator address
	pubKey3 := ed25519.GenPrivKey().PubKey()
	pubKeyAny3, err := codectypes.NewAnyWithValue(pubKey3)
	require.NoError(t, err)
	consAddr3 := sdk.GetConsAddress(pubKey3)

	validator3 := types.Validator{
		PubKey: pubKeyAny3,
		Power:  200,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "another-validator",
			OperatorAddress: operatorAddr.String(), // Same operator address
		},
	}

	err = f.poaKeeper.CreateValidator(f.ctx, consAddr3, validator3, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "uniqueness")

	// Verify original validator is unchanged
	retrieved, err = f.poaKeeper.GetValidator(f.ctx, consAddr)
	require.NoError(t, err)
	require.Equal(t, int64(0), retrieved.Power)
	require.Equal(t, "test-validator", retrieved.Metadata.Moniker)

	// Create validator with power > 0
	pubKey4 := ed25519.GenPrivKey().PubKey()
	pubKeyAny4, err := codectypes.NewAnyWithValue(pubKey4)
	require.NoError(t, err)
	consAddr4 := sdk.GetConsAddress(pubKey4)

	validator4 := types.Validator{
		PubKey: pubKeyAny4,
		Power:  100,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "test-validator-with-power",
			OperatorAddress: sdk.AccAddress("operator2").String(),
		},
	}

	err = f.poaKeeper.CreateValidator(f.ctx, consAddr4, validator4, true)
	require.NoError(t, err)

	// Verify power is set correctly
	power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr4)
	require.NoError(t, err)
	require.Equal(t, int64(100), power)

	// Test same key for operator and consensus - should fail
	t.Run("rejects same key for operator and consensus", func(t *testing.T) {
		sameKey := ed25519.GenPrivKey()
		operatorAddr := sdk.AccAddress(sameKey.PubKey().Address())

		pubKeyAny, err := codectypes.NewAnyWithValue(sameKey.PubKey())
		require.NoError(t, err)

		validator := types.Validator{
			PubKey: pubKeyAny,
			Power:  0,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		consAddress := sdk.GetConsAddress(sameKey.PubKey())
		err = f.poaKeeper.CreateValidator(f.ctx, consAddress, validator, true)

		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrSameKeyForOperatorAndConsensus)
		require.Contains(t, err.Error(), "operator address and consensus pubkey must use different keys")
	})

	// Test different keys for operator and consensus - should succeed
	t.Run("accepts different keys for operator and consensus", func(t *testing.T) {
		operatorKey := ed25519.GenPrivKey()
		operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address())

		consensusKey := ed25519.GenPrivKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(consensusKey.PubKey())
		require.NoError(t, err)

		validator := types.Validator{
			PubKey: pubKeyAny,
			Power:  0,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "test-validator-different-keys",
				OperatorAddress: operatorAddr.String(),
			},
		}

		consAddress := sdk.GetConsAddress(consensusKey.PubKey())
		err = f.poaKeeper.CreateValidator(f.ctx, consAddress, validator, true)

		require.NoError(t, err)

		// Verify validator was created
		exists, err := f.poaKeeper.HasValidator(f.ctx, consAddress)
		require.NoError(t, err)
		require.True(t, exists)
	})
}

func TestHasValidator(t *testing.T) {
	f := setupTest(t)

	consAddr := sdk.ConsAddress("cons1")
	operatorAddr := sdk.AccAddress("operator1")

	pubKey := ed25519.GenPrivKey().PubKey()
	pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	validator := types.Validator{
		PubKey: pubKeyAny,
		Power:  0,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "test-validator",
			OperatorAddress: operatorAddr.String(),
		},
	}

	t.Run("returns false for non-existing validator", func(t *testing.T) {
		exists, err := f.poaKeeper.HasValidator(f.ctx, consAddr)
		require.NoError(t, err)
		require.False(t, exists)
	})

	t.Run("returns true for existing validator", func(t *testing.T) {
		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		exists, err := f.poaKeeper.HasValidator(f.ctx, consAddr)
		require.NoError(t, err)
		require.True(t, exists)
	})
}

func TestGetValidator(t *testing.T) {
	f := setupTest(t)

	consAddr := sdk.ConsAddress("cons1")
	operatorAddr := sdk.AccAddress("operator1")

	pubKey := ed25519.GenPrivKey().PubKey()
	pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	validator := types.Validator{
		PubKey: pubKeyAny,
		Power:  100,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "test-validator",
			OperatorAddress: operatorAddr.String(),
		},
	}

	t.Run("error for non-existing validator", func(t *testing.T) {
		_, err := f.poaKeeper.GetValidator(f.ctx, consAddr)
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("successfully gets existing validator", func(t *testing.T) {
		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		retrieved, err := f.poaKeeper.GetValidator(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, validator.Power, retrieved.Power)
		require.Equal(t, validator.Metadata.Moniker, retrieved.Metadata.Moniker)
		require.Equal(t, validator.Metadata.OperatorAddress, retrieved.Metadata.OperatorAddress)
	})
}

func TestGetValidatorByConsAddress(t *testing.T) {
	f := setupTest(t)

	consAddr := sdk.ConsAddress("cons1")
	operatorAddr := sdk.AccAddress("operator1")

	pubKey := ed25519.GenPrivKey().PubKey()
	pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	validator := types.Validator{
		PubKey: pubKeyAny,
		Power:  100,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "test-validator",
			OperatorAddress: operatorAddr.String(),
		},
	}

	t.Run("error for non-existing consensus address", func(t *testing.T) {
		_, err := f.poaKeeper.GetValidatorByConsAddress(f.ctx, consAddr)
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("successfully gets validator by consensus address", func(t *testing.T) {
		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		retrieved, err := f.poaKeeper.GetValidatorByConsAddress(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, validator.Power, retrieved.Power)
		require.Equal(t, validator.Metadata.Moniker, retrieved.Metadata.Moniker)
	})
}

func TestGetValidatorByOperatorAddress(t *testing.T) {
	f := setupTest(t)

	consAddr := sdk.ConsAddress("cons1")
	operatorAddr := sdk.AccAddress("operator1")

	pubKey := ed25519.GenPrivKey().PubKey()
	pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	validator := types.Validator{
		PubKey: pubKeyAny,
		Power:  100,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "test-validator",
			OperatorAddress: operatorAddr.String(),
		},
	}

	t.Run("error for non-existing operator address", func(t *testing.T) {
		_, err := f.poaKeeper.GetValidatorByOperatorAddress(f.ctx, operatorAddr)
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("successfully gets validator by operator address", func(t *testing.T) {
		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		retrieved, err := f.poaKeeper.GetValidatorByOperatorAddress(f.ctx, operatorAddr)
		require.NoError(t, err)
		require.Equal(t, validator.Power, retrieved.Power)
		require.Equal(t, validator.Metadata.OperatorAddress, retrieved.Metadata.OperatorAddress)
	})
}

func TestSetValidatorPower(t *testing.T) {
	f := setupTest(t)

	consAddr := sdk.ConsAddress("cons1")
	operatorAddr := sdk.AccAddress("operator1")

	pubKey := ed25519.GenPrivKey().PubKey()
	pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	validator := types.Validator{
		PubKey: pubKeyAny,
		Power:  100,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "test-validator",
			OperatorAddress: operatorAddr.String(),
		},
	}

	t.Run("error for non-existing validator", func(t *testing.T) {
		err := f.poaKeeper.SetValidatorPower(f.ctx, consAddr, 200)
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("successfully updates power", func(t *testing.T) {
		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Update power
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr, 200)
		require.NoError(t, err)

		// Verify power changed
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(200), power)

		// Verify validator is still retrievable
		retrieved, err := f.poaKeeper.GetValidator(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(200), retrieved.Power)
	})

	t.Run("can set power to 0 when other validators exist", func(t *testing.T) {
		f := setupTest(t)

		// Create first validator
		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Create second validator to prevent total power from becoming 0
		consAddr2 := sdk.ConsAddress("cons2")
		operatorAddr2 := sdk.AccAddress("operator2")
		pubKey2 := ed25519.GenPrivKey().PubKey()
		pubKeyAny2, err := codectypes.NewAnyWithValue(pubKey2)
		require.NoError(t, err)

		validator2 := types.Validator{
			PubKey: pubKeyAny2,
			Power:  150,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "test-validator-2",
				OperatorAddress: operatorAddr2.String(),
			},
		}
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr2, validator2, true)
		require.NoError(t, err)

		// Now we can set first validator to 0
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr, 0)
		require.NoError(t, err)

		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(0), power)

		// Total power should be 150 (only validator2)
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(150), totalPower)
	})
}

func TestGetValidatorPower(t *testing.T) {
	f := setupTest(t)

	consAddr := sdk.ConsAddress("cons1")
	operatorAddr := sdk.AccAddress("operator1")

	pubKey := ed25519.GenPrivKey().PubKey()
	pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	validator := types.Validator{
		PubKey: pubKeyAny,
		Power:  150,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "test-validator",
			OperatorAddress: operatorAddr.String(),
		},
	}

	t.Run("error for non-existing validator", func(t *testing.T) {
		_, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("successfully gets power", func(t *testing.T) {
		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(150), power)
	})
}

func TestUpdateValidator(t *testing.T) {
	f := setupTest(t)

	// Create initial validator with power 100
	operatorAddr := sdk.AccAddress("operator1")

	pubKey := ed25519.GenPrivKey().PubKey()
	pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	consAddr := sdk.GetConsAddress(pubKey)

	validator := types.Validator{
		PubKey: pubKeyAny,
		Power:  0,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "test-validator",
			OperatorAddress: operatorAddr.String(),
		},
	}

	err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
	require.NoError(t, err)

	t.Run("error for negative power", func(t *testing.T) {
		updates := types.Validator{
			Power: -1,
		}

		err := f.poaKeeper.UpdateValidator(f.ctx, consAddr, updates)
		require.ErrorIs(t, err, types.ErrNegativeValidatorPower)

		// Verify power was not changed
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(0), power)
	})

	t.Run("error for unknown validator", func(t *testing.T) {
		unknownPubKey := ed25519.GenPrivKey().PubKey()
		unknownConsAddr := sdk.GetConsAddress(unknownPubKey)

		updates := types.Validator{
			Power: 200,
		}

		err := f.poaKeeper.UpdateValidator(f.ctx, unknownConsAddr, updates)
		require.ErrorIs(t, err, types.ErrUnknownValidator)
	})

	t.Run("successfully updates validator power", func(t *testing.T) {
		updates := types.Validator{
			Power: 200,
		}

		err := f.poaKeeper.UpdateValidator(f.ctx, consAddr, updates)
		require.NoError(t, err)

		// Verify power updated
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(200), power)

		// Verify validator update was queued
		validatorUpdates := f.poaKeeper.ReapValidatorUpdates(f.ctx)
		require.Len(t, validatorUpdates, 1)
		require.Equal(t, int64(200), validatorUpdates[0].Power)
	})

	t.Run("no update queued if power doesn't change", func(t *testing.T) {
		f := setupTest(t)

		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr, 200)
		require.NoError(t, err)

		updates := types.Validator{
			Power: 200,
		}
		err = f.poaKeeper.UpdateValidator(f.ctx, consAddr, updates)
		require.NoError(t, err)

		validatorUpdates := f.poaKeeper.ReapValidatorUpdates(f.ctx)
		require.Len(t, validatorUpdates, 0)
	})

	t.Run("rejects update with operator address matching consensus key", func(t *testing.T) {
		f := setupTest(t)

		// Create validator with different keys
		operatorKey := ed25519.GenPrivKey()
		operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address())

		consensusKey := ed25519.GenPrivKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(consensusKey.PubKey())
		require.NoError(t, err)

		consAddr := sdk.GetConsAddress(consensusKey.PubKey())

		validator := types.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Try to update operator address to one derived from consensus key
		badOperatorAddr := sdk.AccAddress(consensusKey.PubKey().Address())

		updates := types.Validator{
			Power: 100,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: badOperatorAddr.String(),
			},
		}

		err = f.poaKeeper.UpdateValidator(f.ctx, consAddr, updates)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrSameKeyForOperatorAndConsensus)
		require.Contains(t, err.Error(), "operator address and consensus pubkey must use different keys")
	})

	t.Run("accepts update with different operator address", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		operatorKey := ed25519.GenPrivKey()
		operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address())

		consensusKey := ed25519.GenPrivKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(consensusKey.PubKey())
		require.NoError(t, err)

		consAddr := sdk.GetConsAddress(consensusKey.PubKey())

		validator := types.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Update with a different operator address (still different from consensus key)
		newOperatorKey := ed25519.GenPrivKey()
		newOperatorAddr := sdk.AccAddress(newOperatorKey.PubKey().Address())

		updates := types.Validator{
			Power: 200,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "updated-validator",
				OperatorAddress: newOperatorAddr.String(),
			},
		}

		err = f.poaKeeper.UpdateValidator(f.ctx, consAddr, updates)
		require.NoError(t, err)

		// Verify updates were applied
		updated, err := f.poaKeeper.GetValidator(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(200), updated.Power)
		require.Equal(t, "updated-validator", updated.Metadata.Moniker)
		require.Equal(t, newOperatorAddr.String(), updated.Metadata.OperatorAddress)
	})
}

func TestUpdateValidators(t *testing.T) {
	f := setupTest(t)

	// Create initial validator with power 100
	operatorAddr := sdk.AccAddress("operator1")

	pubKey := ed25519.GenPrivKey().PubKey()
	pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(t, err)

	// Derive consensus address from pubkey
	consAddr := sdk.GetConsAddress(pubKey)

	validator := types.Validator{
		PubKey: pubKeyAny,
		Power:  0,
		Metadata: &types.ValidatorMetadata{
			Moniker:         "test-validator",
			OperatorAddress: operatorAddr.String(),
		},
	}

	err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
	require.NoError(t, err)

	t.Run("error for unknown validator", func(t *testing.T) {
		unknownPubKey := ed25519.GenPrivKey().PubKey()
		unknownPubKeyAny, err := codectypes.NewAnyWithValue(unknownPubKey)
		require.NoError(t, err)

		unknownValidator := types.Validator{
			PubKey: unknownPubKeyAny,
			Power:  100,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "unknown",
				OperatorAddress: sdk.AccAddress("unknown_op").String(),
			},
		}

		err = f.poaKeeper.UpdateValidators(f.ctx, []types.Validator{unknownValidator})
		require.ErrorIs(t, err, types.ErrUnknownValidator)
	})

	t.Run("error for negative power", func(t *testing.T) {
		negativeValidator := validator
		negativeValidator.Power = -1

		err := f.poaKeeper.UpdateValidators(f.ctx, []types.Validator{negativeValidator})
		require.ErrorIs(t, err, types.ErrNegativeValidatorPower)

		// Verify power was not changed
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(0), power)
	})

	t.Run("error for negative power with large negative value", func(t *testing.T) {
		negativeValidator := validator
		negativeValidator.Power = -1000

		err := f.poaKeeper.UpdateValidators(f.ctx, []types.Validator{negativeValidator})
		require.ErrorIs(t, err, types.ErrNegativeValidatorPower)

		// Verify power was not changed
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(0), power)
	})

	t.Run("successfully updates validator power", func(t *testing.T) {
		updatedValidator := validator
		updatedValidator.Power = 200

		err := f.poaKeeper.UpdateValidators(f.ctx, []types.Validator{updatedValidator})
		require.NoError(t, err)

		// Verify power updated
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(200), power)

		// Verify validator update was queued
		updates := f.poaKeeper.ReapValidatorUpdates(f.ctx)
		require.Len(t, updates, 1)
		require.Equal(t, int64(200), updates[0].Power)
	})

	t.Run("no update queued if power doesn't change", func(t *testing.T) {
		f := setupTest(t)

		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr, 200)
		require.NoError(t, err)

		sameValidator := validator
		sameValidator.Power = 200
		err = f.poaKeeper.UpdateValidators(f.ctx, []types.Validator{sameValidator})
		require.NoError(t, err)

		validatorUpdates := f.poaKeeper.ReapValidatorUpdates(f.ctx)
		require.Len(t, validatorUpdates, 0)
	})
}

func TestCompositeKeyUniqueness(t *testing.T) {
	f := setupTest(t)

	t.Run("cannot create duplicate validators with different powers", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		consAddr := sdk.GetConsAddress(pubKey)
		operatorAddr := sdk.AccAddress("operator1")

		validator := types.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		// Create validator with power 100
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Try to create same validator with different power
		validator2 := validator
		validator2.Power = 200
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator2, true)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot create duplicate validator with same consensus address")

		// Verify only one validator exists with original power
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(100), power)
	})

	t.Run("multiple validators can have the same power", func(t *testing.T) {
		// Create multiple validators with the same power but different consensus addresses
		samePower := int64(250)
		var validators []struct {
			consAddr sdk.ConsAddress
			moniker  string
		}

		for i := 0; i < 3; i++ {
			pubKey := ed25519.GenPrivKey().PubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)

			consAddr := sdk.GetConsAddress(pubKey)
			moniker := sdk.AccAddress([]byte{byte(i + 10)}).String()

			validator := types.Validator{
				PubKey: pubKeyAny,
				Power:  samePower,
				Metadata: &types.ValidatorMetadata{
					Moniker:         moniker,
					OperatorAddress: moniker,
				},
			}

			err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
			require.NoError(t, err)

			validators = append(validators, struct {
				consAddr sdk.ConsAddress
				moniker  string
			}{consAddr, moniker})
		}

		// Verify all validators exist with the same power
		for _, v := range validators {
			power, err := f.poaKeeper.GetValidatorPower(f.ctx, v.consAddr)
			require.NoError(t, err)
			require.Equal(t, samePower, power)

			// Verify we can retrieve each validator individually
			retrieved, err := f.poaKeeper.GetValidator(f.ctx, v.consAddr)
			require.NoError(t, err)
			require.Equal(t, samePower, retrieved.Power)
			require.Equal(t, v.moniker, retrieved.Metadata.Moniker)
		}

		// Verify all validators appear in iteration
		ranger := new(collections.Range[collections.Pair[int64, string]]).Descending()
		var foundValidators []string
		err := f.poaKeeper.IterateValidators(f.ctx, ranger, func(power int64, val types.Validator) (bool, error) {
			if power == samePower {
				foundValidators = append(foundValidators, val.Metadata.Moniker)
			}
			return false, nil
		})
		require.NoError(t, err)

		// Should find all 3 validators with the same power
		require.Len(t, foundValidators, 3)
		for _, v := range validators {
			require.Contains(t, foundValidators, v.moniker)
		}
	})

	t.Run("cannot create validators with different consensus addresses but same operator address", func(t *testing.T) {
		// Create first validator with a specific operator address
		sameOperatorAddr := sdk.AccAddress("shared-operator")

		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKeyAny1, err := codectypes.NewAnyWithValue(pubKey1)
		require.NoError(t, err)
		consAddr1 := sdk.GetConsAddress(pubKey1)

		validator1 := types.Validator{
			PubKey: pubKeyAny1,
			Power:  100,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "validator-1",
				OperatorAddress: sameOperatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr1, validator1, true)
		require.NoError(t, err)

		// Try to create second validator with different consensus address but same operator address
		pubKey2 := ed25519.GenPrivKey().PubKey()
		pubKeyAny2, err := codectypes.NewAnyWithValue(pubKey2)
		require.NoError(t, err)
		consAddr2 := sdk.GetConsAddress(pubKey2)

		// Ensure consensus addresses are different
		require.NotEqual(t, consAddr1.String(), consAddr2.String())

		validator2 := types.Validator{
			PubKey: pubKeyAny2,
			Power:  200,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "validator-2",
				OperatorAddress: sameOperatorAddr.String(), // Same operator address
			},
		}

		// This should fail due to operator address uniqueness constraint
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr2, validator2, true)
		require.Error(t, err)
		require.Contains(t, err.Error(), "uniqueness")

		// Verify only the first validator exists
		retrieved, err := f.poaKeeper.GetValidatorByOperatorAddress(f.ctx, sameOperatorAddr)
		require.NoError(t, err)
		require.Equal(t, validator1.Metadata.Moniker, retrieved.Metadata.Moniker)
		require.Equal(t, int64(100), retrieved.Power)

		// Verify second validator was not created
		_, err = f.poaKeeper.GetValidator(f.ctx, consAddr2)
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("cannot overwrite validator with same primary key (same power and consensus address)", func(t *testing.T) {
		// This test verifies that the HasValidator check in CreateValidator prevents
		// the edge case where Set() with the same primary key (power, consensus_address)
		// would overwrite without checking unique indexes, allowing operator address to change

		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)
		consAddr := sdk.GetConsAddress(pubKey)
		operatorAddr1 := sdk.AccAddress("same-key-test-operator1")

		// Create first validator with power=0
		validator1 := types.Validator{
			PubKey: pubKeyAny,
			Power:  0,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "same-key-validator-1",
				OperatorAddress: operatorAddr1.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator1, true)
		require.NoError(t, err)

		// Try to create second validator with same consensus address AND same power
		// This would have the same primary key: (0, consAddr)
		// Without HasValidator check, Set() would overwrite, changing operator address
		operatorAddr2 := sdk.AccAddress("same-key-test-operator2")
		validator2 := types.Validator{
			PubKey: pubKeyAny,
			Power:  0, // Same power = same primary key
			Metadata: &types.ValidatorMetadata{
				Moniker:         "same-key-validator-2",
				OperatorAddress: operatorAddr2.String(), // Different operator
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator2, true)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot create duplicate validator with same consensus address")

		// Verify original validator is unchanged
		retrieved, err := f.poaKeeper.GetValidator(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(0), retrieved.Power)
		require.Equal(t, "same-key-validator-1", retrieved.Metadata.Moniker)
		require.Equal(t, operatorAddr1.String(), retrieved.Metadata.OperatorAddress)

		// Verify the operator address was NOT changed to operator2
		require.NotEqual(t, operatorAddr2.String(), retrieved.Metadata.OperatorAddress)

		// Verify GetValidatorByOperatorAddress returns the original validator
		retrieved, err = f.poaKeeper.GetValidatorByOperatorAddress(f.ctx, operatorAddr1)
		require.NoError(t, err)
		require.Equal(t, "same-key-validator-1", retrieved.Metadata.Moniker)

		// Verify operator2 does not have a validator
		_, err = f.poaKeeper.GetValidatorByOperatorAddress(f.ctx, operatorAddr2)
		require.ErrorIs(t, err, collections.ErrNotFound)
	})
}

func TestValidatorIteration(t *testing.T) {
	f := setupTest(t)

	// Create validators with different power levels
	type validatorInfo struct {
		consAddr sdk.ConsAddress
		power    int64
		moniker  string
	}

	validators := []validatorInfo{
		{power: 300, moniker: "validator-300"},
		{power: 100, moniker: "validator-100"},
		{power: 500, moniker: "validator-500"},
		{power: 200, moniker: "validator-200"},
		{power: 0, moniker: "validator-0"},
	}

	// Create all validators
	for i, v := range validators {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		consAddr := sdk.GetConsAddress(pubKey)
		validators[i].consAddr = consAddr

		operatorAddr := sdk.AccAddress(v.moniker)
		validator := types.Validator{
			PubKey: pubKeyAny,
			Power:  v.power,
			Metadata: &types.ValidatorMetadata{
				Moniker:         v.moniker,
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)
	}

	t.Run("iterates validators in descending power order", func(t *testing.T) {
		// Iterate in descending order (highest power first)
		ranger := new(collections.Range[collections.Pair[int64, string]]).Descending()

		var iteratedValidators []validatorInfo
		err := f.poaKeeper.IterateValidators(f.ctx, ranger, func(power int64, val types.Validator) (bool, error) {
			iteratedValidators = append(iteratedValidators, validatorInfo{
				power:   power,
				moniker: val.Metadata.Moniker,
			})
			return false, nil
		})
		require.NoError(t, err)

		// Verify order: 500, 300, 200, 100, 0
		require.Len(t, iteratedValidators, 5)
		require.Equal(t, int64(500), iteratedValidators[0].power)
		require.Equal(t, "validator-500", iteratedValidators[0].moniker)
		require.Equal(t, int64(300), iteratedValidators[1].power)
		require.Equal(t, "validator-300", iteratedValidators[1].moniker)
		require.Equal(t, int64(200), iteratedValidators[2].power)
		require.Equal(t, "validator-200", iteratedValidators[2].moniker)
		require.Equal(t, int64(100), iteratedValidators[3].power)
		require.Equal(t, "validator-100", iteratedValidators[3].moniker)
		require.Equal(t, int64(0), iteratedValidators[4].power)
		require.Equal(t, "validator-0", iteratedValidators[4].moniker)
	})

	t.Run("stops iteration at power 0", func(t *testing.T) {
		// Iterate but stop when we hit power 0
		ranger := new(collections.Range[collections.Pair[int64, string]]).Descending()

		var activeValidators []validatorInfo
		err := f.poaKeeper.IterateValidators(f.ctx, ranger, func(power int64, val types.Validator) (bool, error) {
			// Stop when we hit power 0
			if power == 0 {
				return true, nil
			}

			activeValidators = append(activeValidators, validatorInfo{
				power:   power,
				moniker: val.Metadata.Moniker,
			})
			return false, nil
		})
		require.NoError(t, err)

		// Should only have validators with power > 0
		require.Len(t, activeValidators, 4)
		require.Equal(t, int64(500), activeValidators[0].power)
		require.Equal(t, int64(300), activeValidators[1].power)
		require.Equal(t, int64(200), activeValidators[2].power)
		require.Equal(t, int64(100), activeValidators[3].power)
	})
}

func TestValidatorDemotion(t *testing.T) {
	f := setupTest(t)

	// Create validators with different power levels
	type validatorInfo struct {
		consAddr sdk.ConsAddress
		pubKey   *codectypes.Any
		power    int64
		moniker  string
	}

	validators := []validatorInfo{
		{power: 300, moniker: "validator-300"},
		{power: 100, moniker: "validator-100"},
		{power: 500, moniker: "validator-500"}, // This will be demoted
		{power: 200, moniker: "validator-200"},
	}

	// Create all validators
	for i, v := range validators {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		consAddr := sdk.GetConsAddress(pubKey)
		validators[i].consAddr = consAddr
		validators[i].pubKey = pubKeyAny

		operatorAddr := sdk.AccAddress(v.moniker)
		validator := types.Validator{
			PubKey: pubKeyAny,
			Power:  v.power,
			Metadata: &types.ValidatorMetadata{
				Moniker:         v.moniker,
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)
	}

	t.Run("demote highest power validator to 0 and verify iteration", func(t *testing.T) {
		// Find the validator-500 (highest power)
		var validator500 validatorInfo
		for _, v := range validators {
			if v.moniker == "validator-500" {
				validator500 = v
				break
			}
		}

		// Demote validator-500 to power 0
		err := f.poaKeeper.SetValidatorPower(f.ctx, validator500.consAddr, 0)
		require.NoError(t, err)

		// Verify power was updated
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, validator500.consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(0), power)

		// Iterate in descending order
		ranger := new(collections.Range[collections.Pair[int64, string]]).Descending()

		var iteratedValidators []validatorInfo
		err = f.poaKeeper.IterateValidators(f.ctx, ranger, func(power int64, val types.Validator) (bool, error) {
			iteratedValidators = append(iteratedValidators, validatorInfo{
				power:   power,
				moniker: val.Metadata.Moniker,
			})
			return false, nil
		})
		require.NoError(t, err)

		// Verify order: 300, 200, 100, 0 (validator-500 now at the end)
		require.Len(t, iteratedValidators, 4)
		require.Equal(t, int64(300), iteratedValidators[0].power)
		require.Equal(t, "validator-300", iteratedValidators[0].moniker)
		require.Equal(t, int64(200), iteratedValidators[1].power)
		require.Equal(t, "validator-200", iteratedValidators[1].moniker)
		require.Equal(t, int64(100), iteratedValidators[2].power)
		require.Equal(t, "validator-100", iteratedValidators[2].moniker)
		require.Equal(t, int64(0), iteratedValidators[3].power)
		require.Equal(t, "validator-500", iteratedValidators[3].moniker) // Demoted validator at end

		// Now iterate with stop at power 0
		var activeValidators []validatorInfo
		err = f.poaKeeper.IterateValidators(f.ctx, ranger, func(power int64, val types.Validator) (bool, error) {
			// Stop when we hit power 0
			if power == 0 {
				return true, nil
			}

			activeValidators = append(activeValidators, validatorInfo{
				power:   power,
				moniker: val.Metadata.Moniker,
			})
			return false, nil
		})
		require.NoError(t, err)

		// Should only have 3 active validators (validator-500 excluded)
		require.Len(t, activeValidators, 3)
		require.Equal(t, int64(300), activeValidators[0].power)
		require.Equal(t, int64(200), activeValidators[1].power)
		require.Equal(t, int64(100), activeValidators[2].power)

		// Verify validator-500 is NOT in the active list
		for _, v := range activeValidators {
			require.NotEqual(t, "validator-500", v.moniker)
		}
	})
}

func TestAdjustTotalPower(t *testing.T) {
	t.Run("zero delta does nothing", func(t *testing.T) {
		f := setupTest(t)

		// Set initial total power
		err := f.poaKeeper.AdjustTotalPower(f.ctx, 100)
		require.NoError(t, err)

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)

		// Adjust by zero
		err = f.poaKeeper.AdjustTotalPower(f.ctx, 0)
		require.NoError(t, err)

		// Total should be unchanged
		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)
	})

	t.Run("positive delta increases total power", func(t *testing.T) {
		f := setupTest(t)

		// Set initial total power
		err := f.poaKeeper.AdjustTotalPower(f.ctx, 100)
		require.NoError(t, err)

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)

		// Increase by 50
		err = f.poaKeeper.AdjustTotalPower(f.ctx, 50)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(150), totalPower)

		// Increase by 200
		err = f.poaKeeper.AdjustTotalPower(f.ctx, 200)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(350), totalPower)
	})

	t.Run("negative delta decreases total power", func(t *testing.T) {
		f := setupTest(t)

		// Set initial total power
		err := f.poaKeeper.AdjustTotalPower(f.ctx, 500)
		require.NoError(t, err)

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(500), totalPower)

		// Decrease by 100
		err = f.poaKeeper.AdjustTotalPower(f.ctx, -100)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(400), totalPower)

		// Decrease by 250
		err = f.poaKeeper.AdjustTotalPower(f.ctx, -250)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(150), totalPower)
	})

	t.Run("error when total power would become negative", func(t *testing.T) {
		f := setupTest(t)

		// Set initial total power
		err := f.poaKeeper.AdjustTotalPower(f.ctx, 100)
		require.NoError(t, err)

		// Try to decrease by more than current total
		err = f.poaKeeper.AdjustTotalPower(f.ctx, -150)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidTotalPower)
		require.Contains(t, err.Error(), "total power would become negative")

		// Verify total power unchanged
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)
	})

	t.Run("error when total power would become zero", func(t *testing.T) {
		f := setupTest(t)

		// Set initial total power
		err := f.poaKeeper.AdjustTotalPower(f.ctx, 100)
		require.NoError(t, err)

		// Try to decrease to exactly zero
		err = f.poaKeeper.AdjustTotalPower(f.ctx, -100)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidTotalPower)
		require.Contains(t, err.Error(), "total power cannot be zero")

		// Verify total power unchanged
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)
	})

	t.Run("multiple sequential adjustments", func(t *testing.T) {
		f := setupTest(t)

		// Start with 100
		err := f.poaKeeper.AdjustTotalPower(f.ctx, 100)
		require.NoError(t, err)

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)

		// Add 50
		err = f.poaKeeper.AdjustTotalPower(f.ctx, 50)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(150), totalPower)

		// Subtract 30
		err = f.poaKeeper.AdjustTotalPower(f.ctx, -30)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(120), totalPower)

		// Add 200
		err = f.poaKeeper.AdjustTotalPower(f.ctx, 200)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(320), totalPower)

		// Subtract 100
		err = f.poaKeeper.AdjustTotalPower(f.ctx, -100)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(220), totalPower)
	})

	t.Run("large positive and negative deltas", func(t *testing.T) {
		f := setupTest(t)

		// Start with large value
		err := f.poaKeeper.AdjustTotalPower(f.ctx, 1000000)
		require.NoError(t, err)

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(1000000), totalPower)

		// Large decrease
		err = f.poaKeeper.AdjustTotalPower(f.ctx, -500000)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(500000), totalPower)

		// Large increase
		err = f.poaKeeper.AdjustTotalPower(f.ctx, 2000000)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(2500000), totalPower)
	})

	t.Run("error cases preserve state", func(t *testing.T) {
		f := setupTest(t)

		// Set initial total power
		err := f.poaKeeper.AdjustTotalPower(f.ctx, 100)
		require.NoError(t, err)

		// Failed adjustment (would go negative)
		err = f.poaKeeper.AdjustTotalPower(f.ctx, -200)
		require.Error(t, err)

		// State should be unchanged
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)

		// Failed adjustment (would go to zero)
		err = f.poaKeeper.AdjustTotalPower(f.ctx, -100)
		require.Error(t, err)

		// State should still be unchanged
		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)

		// Successful adjustment should work
		err = f.poaKeeper.AdjustTotalPower(f.ctx, 50)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(150), totalPower)
	})

	t.Run("sum of validator powers equals total power", func(t *testing.T) {
		f := setupTest(t)

		// Create multiple validators with different powers
		validatorPowers := []int64{100, 250, 75, 300, 50}
		var consAddrs []sdk.ConsAddress

		for i, power := range validatorPowers {
			pubKey := ed25519.GenPrivKey().PubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)
			consAddr := sdk.GetConsAddress(pubKey)
			consAddrs = append(consAddrs, consAddr)

			validator := types.Validator{
				PubKey: pubKeyAny,
				Power:  power,
				Metadata: &types.ValidatorMetadata{
					Moniker:         sdk.AccAddress([]byte{byte(i)}).String(),
					OperatorAddress: sdk.AccAddress([]byte{byte(i)}).String(),
				},
			}

			err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
			require.NoError(t, err)
		}

		// Calculate sum of all validator powers
		var sum int64
		for _, consAddr := range consAddrs {
			power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
			require.NoError(t, err)
			sum += power
		}

		// Verify sum equals total power
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, sum, totalPower)
		require.Equal(t, int64(775), totalPower) // 100+250+75+300+50

		// Change some validator powers
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddrs[0], 150) // 100 -> 150
		require.NoError(t, err)

		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddrs[2], 0) // 75 -> 0
		require.NoError(t, err)

		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddrs[3], 400) // 300 -> 400
		require.NoError(t, err)

		// Recalculate sum after changes
		sum = 0
		for _, consAddr := range consAddrs {
			power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
			require.NoError(t, err)
			sum += power
		}

		// Verify sum still equals total power after changes
		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, sum, totalPower)
		require.Equal(t, int64(850), totalPower) // 150+250+0+400+50
	})
}

func TestGetTotalPower(t *testing.T) {
	f := setupTest(t)

	t.Run("creating validators updates total power", func(t *testing.T) {
		// Create first validator with power 100
		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKeyAny1, err := codectypes.NewAnyWithValue(pubKey1)
		require.NoError(t, err)
		consAddr1 := sdk.GetConsAddress(pubKey1)

		validator1 := types.Validator{
			PubKey: pubKeyAny1,
			Power:  100,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "validator-1",
				OperatorAddress: sdk.AccAddress("operator1").String(),
			},
		}
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr1, validator1, true)
		require.NoError(t, err)

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)

		// Create second validator with power 200
		pubKey2 := ed25519.GenPrivKey().PubKey()
		pubKeyAny2, err := codectypes.NewAnyWithValue(pubKey2)
		require.NoError(t, err)
		consAddr2 := sdk.GetConsAddress(pubKey2)

		validator2 := types.Validator{
			PubKey: pubKeyAny2,
			Power:  200,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "validator-2",
				OperatorAddress: sdk.AccAddress("operator2").String(),
			},
		}
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr2, validator2, true)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(300), totalPower)

		// Create third validator with power 0 (should not change total)
		pubKey3 := ed25519.GenPrivKey().PubKey()
		pubKeyAny3, err := codectypes.NewAnyWithValue(pubKey3)
		require.NoError(t, err)
		consAddr3 := sdk.GetConsAddress(pubKey3)

		validator3 := types.Validator{
			PubKey: pubKeyAny3,
			Power:  0,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "validator-3",
				OperatorAddress: sdk.AccAddress("operator3").String(),
			},
		}
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr3, validator3, true)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(300), totalPower)
	})

	t.Run("updating validator power adjusts total power", func(t *testing.T) {
		f := setupTest(t)

		// Create validator with power 150
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)
		consAddr := sdk.GetConsAddress(pubKey)

		validator := types.Validator{
			PubKey: pubKeyAny,
			Power:  150,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: sdk.AccAddress("operator").String(),
			},
		}
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(150), totalPower)

		// Update power to 250
		validator.Power = 250
		err = f.poaKeeper.UpdateValidator(f.ctx, consAddr, validator)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(250), totalPower) // Increased by 100

		// Update power to 50
		validator.Power = 50
		err = f.poaKeeper.UpdateValidator(f.ctx, consAddr, validator)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(50), totalPower) // Decreased by 200
	})

	t.Run("setting validator power to zero adjusts total power", func(t *testing.T) {
		f := setupTest(t)

		// Create two validators
		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKeyAny1, err := codectypes.NewAnyWithValue(pubKey1)
		require.NoError(t, err)
		consAddr1 := sdk.GetConsAddress(pubKey1)

		validator1 := types.Validator{
			PubKey: pubKeyAny1,
			Power:  100,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "validator-1",
				OperatorAddress: sdk.AccAddress("operator1").String(),
			},
		}
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr1, validator1, true)
		require.NoError(t, err)

		pubKey2 := ed25519.GenPrivKey().PubKey()
		pubKeyAny2, err := codectypes.NewAnyWithValue(pubKey2)
		require.NoError(t, err)
		consAddr2 := sdk.GetConsAddress(pubKey2)

		validator2 := types.Validator{
			PubKey: pubKeyAny2,
			Power:  200,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "validator-2",
				OperatorAddress: sdk.AccAddress("operator2").String(),
			},
		}
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr2, validator2, true)
		require.NoError(t, err)

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(300), totalPower)

		// Set first validator power to 0
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr1, 0)
		require.NoError(t, err)

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(200), totalPower) // Only validator2's power remains

		// Try to set second validator power to 0 - should fail as it's the last one
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr2, 0)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidTotalPower)

		// Total power should still be 200
		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(200), totalPower)
	})

	t.Run("complex scenario with multiple power changes", func(t *testing.T) {
		f := setupTest(t)

		// Create three validators with different powers
		var validators []struct {
			consAddr sdk.ConsAddress
			power    int64
		}

		initialPowers := []int64{100, 200, 300}
		expectedTotal := int64(600)

		for i, power := range initialPowers {
			pubKey := ed25519.GenPrivKey().PubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)
			consAddr := sdk.GetConsAddress(pubKey)

			validator := types.Validator{
				PubKey: pubKeyAny,
				Power:  power,
				Metadata: &types.ValidatorMetadata{
					Moniker:         sdk.AccAddress([]byte{byte(i)}).String(),
					OperatorAddress: sdk.AccAddress([]byte{byte(i)}).String(),
				},
			}
			err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
			require.NoError(t, err)

			validators = append(validators, struct {
				consAddr sdk.ConsAddress
				power    int64
			}{consAddr, power})
		}

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, expectedTotal, totalPower)

		// Increase first validator power by 50
		err = f.poaKeeper.SetValidatorPower(f.ctx, validators[0].consAddr, 150)
		require.NoError(t, err)
		expectedTotal += 50

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, expectedTotal, totalPower)

		// Decrease second validator power by 100
		err = f.poaKeeper.SetValidatorPower(f.ctx, validators[1].consAddr, 100)
		require.NoError(t, err)
		expectedTotal -= 100

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, expectedTotal, totalPower)

		// Set third validator power to 0
		err = f.poaKeeper.SetValidatorPower(f.ctx, validators[2].consAddr, 0)
		require.NoError(t, err)
		expectedTotal -= 300

		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, expectedTotal, totalPower)

		// Final total should be 150 + 100 + 0 = 250
		require.Equal(t, int64(250), expectedTotal)
	})

	t.Run("error when total power would become negative", func(t *testing.T) {
		f := setupTest(t)

		// Create validator with power 100
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)
		consAddr := sdk.GetConsAddress(pubKey)

		validator := types.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: sdk.AccAddress("operator").String(),
			},
		}
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)

		// Try to set power to a value that would make total negative
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr, -100)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidTotalPower)

		// Verify total power unchanged
		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)
	})

	t.Run("error when setting last validator power to zero", func(t *testing.T) {
		f := setupTest(t)

		// Create a single validator with power 100
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)
		consAddr := sdk.GetConsAddress(pubKey)

		validator := types.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "only-validator",
				OperatorAddress: sdk.AccAddress("operator").String(),
			},
		}
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)

		// Try to set the only validator's power to 0 - should fail
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr, 0)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidTotalPower)

		// Verify validator power was NOT changed (validation failed before changes)
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(100), power)

		// Verify total power unchanged
		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)
	})

	t.Run("can set validator power to zero if other validators exist", func(t *testing.T) {
		f := setupTest(t)

		// Create two validators
		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKeyAny1, err := codectypes.NewAnyWithValue(pubKey1)
		require.NoError(t, err)
		consAddr1 := sdk.GetConsAddress(pubKey1)

		validator1 := types.Validator{
			PubKey: pubKeyAny1,
			Power:  100,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "validator-1",
				OperatorAddress: sdk.AccAddress("operator1").String(),
			},
		}
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr1, validator1, true)
		require.NoError(t, err)

		pubKey2 := ed25519.GenPrivKey().PubKey()
		pubKeyAny2, err := codectypes.NewAnyWithValue(pubKey2)
		require.NoError(t, err)
		consAddr2 := sdk.GetConsAddress(pubKey2)

		validator2 := types.Validator{
			PubKey: pubKeyAny2,
			Power:  200,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "validator-2",
				OperatorAddress: sdk.AccAddress("operator2").String(),
			},
		}
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr2, validator2, true)
		require.NoError(t, err)

		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(300), totalPower)

		// Set first validator power to 0 - should succeed because validator2 still has power
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr1, 0)
		require.NoError(t, err)

		// Verify total power is now 200
		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(200), totalPower)

		// Try to set second validator to 0 - should fail as it's the last one
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr2, 0)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidTotalPower)

		// Verify total power still 200
		totalPower, err = f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(200), totalPower)
	})
}

func TestGetAllValidators(t *testing.T) {
	f := setupTest(t)

	t.Run("when no validators exist", func(t *testing.T) {
		// Initially, there should be no validators
		validators, err := f.poaKeeper.GetAllValidators(f.ctx)
		require.NoError(t, err)
		require.Empty(t, validators)
	})

	t.Run("with multiple validators", func(t *testing.T) {
		// Create multiple validators with different powers
		numValidators := 5
		createdValidators := make([]types.Validator, numValidators)
		consAddrs := make([]sdk.ConsAddress, numValidators)

		for i := 0; i < numValidators; i++ {
			pubKey := ed25519.GenPrivKey().PubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)

			consAddr := sdk.GetConsAddress(pubKey)
			consAddrs[i] = consAddr
			operatorAddr := sdk.AccAddress(fmt.Sprintf("operator%d", i))

			validator := types.Validator{
				PubKey: pubKeyAny,
				Power:  int64((i + 1) * 10), // Different powers: 10, 20, 30, 40, 50
				Metadata: &types.ValidatorMetadata{
					Moniker:         fmt.Sprintf("validator-%d", i),
					OperatorAddress: operatorAddr.String(),
				},
			}

			err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
			require.NoError(t, err)
			createdValidators[i] = validator
		}

		// Get all validators
		allValidators, err := f.poaKeeper.GetAllValidators(f.ctx)
		require.NoError(t, err)
		require.Len(t, allValidators, numValidators)

		// Verify validators are returned in descending power order
		for i := 0; i < len(allValidators)-1; i++ {
			require.GreaterOrEqual(t, allValidators[i].Power, allValidators[i+1].Power,
				"validators should be in descending power order: validator at index %d has power %d, validator at index %d has power %d",
				i, allValidators[i].Power, i+1, allValidators[i+1].Power)
		}

		// Verify all validators are present
		// Create a map of operator addresses for easy lookup
		validatorMap := make(map[string]types.Validator)
		for _, v := range allValidators {
			validatorMap[v.Metadata.OperatorAddress] = v
		}

		// Check that all created validators are in the result
		for i, created := range createdValidators {
			found, exists := validatorMap[created.Metadata.OperatorAddress]
			require.True(t, exists, "validator %d should be found", i)
			require.Equal(t, created.Power, found.Power)
			require.Equal(t, created.Metadata.Moniker, found.Metadata.Moniker)
			require.Equal(t, created.Metadata.OperatorAddress, found.Metadata.OperatorAddress)
		}
	})

	t.Run("validators returned in descending power order", func(t *testing.T) {
		// Use a fresh fixture to avoid state pollution from previous tests
		fOrder := setupTest(t)

		// Create validators with different powers in random order
		powers := []int64{100, 50, 200, 0, 150, 75}
		validators := make([]types.Validator, len(powers))
		consAddrs := make([]sdk.ConsAddress, len(powers))

		for i, power := range powers {
			pubKey := ed25519.GenPrivKey().PubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)

			consAddr := sdk.GetConsAddress(pubKey)
			consAddrs[i] = consAddr
			operatorAddr := sdk.AccAddress(fmt.Sprintf("operator-order-%d", i))

			validators[i] = types.Validator{
				PubKey: pubKeyAny,
				Power:  power,
				Metadata: &types.ValidatorMetadata{
					Moniker:         fmt.Sprintf("validator-order-%d", i),
					OperatorAddress: operatorAddr.String(),
					Description:     fmt.Sprintf("Validator with power %d", power),
				},
			}

			err = fOrder.poaKeeper.CreateValidator(fOrder.ctx, consAddr, validators[i], true)
			require.NoError(t, err)
		}

		// Get all validators
		allValidators, err := fOrder.poaKeeper.GetAllValidators(fOrder.ctx)
		require.NoError(t, err)
		require.Len(t, allValidators, len(powers))

		// Verify validators are returned in descending power order
		// Expected order: 200, 150, 100, 75, 50, 0
		expectedPowers := []int64{200, 150, 100, 75, 50, 0}
		for i, expectedPower := range expectedPowers {
			require.Equal(t, expectedPower, allValidators[i].Power,
				"validator at index %d should have power %d, but has %d", i, expectedPower, allValidators[i].Power)
		}

		// Also verify that each validator's power is >= the next validator's power
		for i := 0; i < len(allValidators)-1; i++ {
			require.GreaterOrEqual(t, allValidators[i].Power, allValidators[i+1].Power,
				"validators should be in descending power order: validator at index %d has power %d, validator at index %d has power %d",
				i, allValidators[i].Power, i+1, allValidators[i+1].Power)
		}
	})

	t.Run("including zero power validator", func(t *testing.T) {
		// Test with a validator with power 0
		pubKeyZero := ed25519.GenPrivKey().PubKey()
		pubKeyAnyZero, err := codectypes.NewAnyWithValue(pubKeyZero)
		require.NoError(t, err)
		consAddrZero := sdk.GetConsAddress(pubKeyZero)
		operatorAddrZero := sdk.AccAddress("operator-zero")

		validatorZero := types.Validator{
			PubKey: pubKeyAnyZero,
			Power:  0,
			Metadata: &types.ValidatorMetadata{
				Moniker:         "validator-zero",
				OperatorAddress: operatorAddrZero.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddrZero, validatorZero, true)
		require.NoError(t, err)

		// GetAllValidators should include the validator with power 0
		allValidators, err := f.poaKeeper.GetAllValidators(f.ctx)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(allValidators), 6) // At least 5 from previous test + 1 zero power

		// Verify the zero-power validator is included
		foundZero := false
		for _, v := range allValidators {
			if v.Metadata.OperatorAddress == validatorZero.Metadata.OperatorAddress {
				foundZero = true
				require.Equal(t, int64(0), v.Power)
				break
			}
		}
		require.True(t, foundZero, "validator with power 0 should be included")
	})
}

func TestCreateABCIValidatorUpdate(t *testing.T) {
	f := setupTest(t)

	t.Run("creates validator update with ed25519 key", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		update, err := f.poaKeeper.createABCIValidatorUpdate(pubKeyAny, 100)
		require.NoError(t, err)
		require.Equal(t, int64(100), update.Power)
		require.NotNil(t, update.PubKey)
	})

	t.Run("creates validator update with zero power", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		update, err := f.poaKeeper.createABCIValidatorUpdate(pubKeyAny, 0)
		require.NoError(t, err)
		require.Equal(t, int64(0), update.Power)
		require.NotNil(t, update.PubKey)
	})

	t.Run("creates validator update with high power", func(t *testing.T) {
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		update, err := f.poaKeeper.createABCIValidatorUpdate(pubKeyAny, 1000000)
		require.NoError(t, err)
		require.Equal(t, int64(1000000), update.Power)
		require.NotNil(t, update.PubKey)
	})

	t.Run("returns error for invalid pubkey", func(t *testing.T) {
		invalidPubKeyAny := codectypes.UnsafePackAny("not-a-pubkey")

		_, err := f.poaKeeper.createABCIValidatorUpdate(invalidPubKeyAny, 100)
		require.Error(t, err)
	})
}
