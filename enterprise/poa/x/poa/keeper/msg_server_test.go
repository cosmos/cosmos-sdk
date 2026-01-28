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
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
)

var adminAddr = sdk.AccAddress("admin").String()

func TestMsgServerUpdateParams(t *testing.T) {
	t.Run("successfully updates params with valid admin", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// Update params with valid admin
		newParams := poatypes.Params{Admin: sdk.AccAddress("newadmin").String()}
		msg := &poatypes.MsgUpdateParams{
			Admin:  adminAddr,
			Params: newParams,
		}

		resp, err := msgServer.UpdateParams(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify params were updated
		updatedParams, err := f.poaKeeper.Params(f.ctx, &poatypes.QueryParamsRequest{})
		require.NoError(t, err)
		require.Equal(t, newParams.Admin, updatedParams.Params.Admin)

		// Verify event was emitted
		events := f.ctx.EventManager().Events()
		require.Len(t, events, 1)
		require.Equal(t, poatypes.EventTypeUpdateParams, events[0].Type)
		require.Len(t, events[0].Attributes, 2)
		require.Equal(t, poatypes.AttributeKeyAdmin, events[0].Attributes[0].Key)
		require.Equal(t, adminAddr, events[0].Attributes[0].Value)
		require.Equal(t, poatypes.AttributeKeyParams, events[0].Attributes[1].Key)
		require.Contains(t, events[0].Attributes[1].Value, newParams.Admin)
	})

	t.Run("fails with invalid admin", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)
		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// Try to update params with wrong admin
		msg := &poatypes.MsgUpdateParams{
			Admin:  sdk.AccAddress("wrongadmin").String(),
			Params: poatypes.Params{Admin: sdk.AccAddress("newadmin").String()},
		}

		_, err = msgServer.UpdateParams(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid authority")
	})
}

func TestMsgServerCreateValidator(t *testing.T) {
	t.Run("successfully creates validator", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Generate pubkey
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)

		// Create validator
		operatorAddr := sdk.AccAddress("operator1")
		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator",
			Description:     "test description",
			OperatorAddress: operatorAddr.String(),
		}

		resp, err := msgServer.CreateValidator(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify validator was created
		consAddr := sdk.GetConsAddress(pubKey)
		validator, err := f.poaKeeper.GetValidator(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(0), validator.Power) // New validators start with 0 power
		require.Equal(t, "test-validator", validator.Metadata.Moniker)
		require.Equal(t, operatorAddr.String(), validator.Metadata.OperatorAddress)

		// Verify event was emitted
		events := f.ctx.EventManager().Events()
		require.Len(t, events, 1)
		require.Equal(t, poatypes.EventTypeCreateValidator, events[0].Type)
		require.Len(t, events[0].Attributes, 4)

		// Check attributes
		attrs := make(map[string]string)
		for _, attr := range events[0].Attributes {
			attrs[attr.Key] = attr.Value
		}
		require.Equal(t, operatorAddr.String(), attrs[poatypes.AttributeKeyOperatorAddress])
		require.Equal(t, consAddr.String(), attrs[poatypes.AttributeKeyConsensusAddress])
		require.Equal(t, "test-validator", attrs[poatypes.AttributeKeyMoniker])
		require.Equal(t, "0", attrs[poatypes.AttributeKeyPower])
	})

	t.Run("creates validator with complete metadata", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Generate pubkey
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)

		// Create validator with full metadata
		operatorAddr := sdk.AccAddress("operator1")
		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator",
			Description:     "A test validator for unit tests",
			OperatorAddress: operatorAddr.String(),
		}

		resp, err := msgServer.CreateValidator(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify all metadata was stored
		consAddr := sdk.GetConsAddress(pubKey)
		validator, err := f.poaKeeper.GetValidator(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, "test-validator", validator.Metadata.Moniker)
		require.Equal(t, "A test validator for unit tests", validator.Metadata.Description)
		require.Equal(t, operatorAddr.String(), validator.Metadata.OperatorAddress)
	})

	t.Run("fails validation with empty moniker", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		operatorAddr := sdk.AccAddress("operator1")

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "", // Empty moniker
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		_, err := msgServer.CreateValidator(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "moniker cannot be empty")
	})

	t.Run("fails validation with moniker too long", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		operatorAddr := sdk.AccAddress("operator1")

		// Create a moniker longer than 256 characters
		longMoniker := string(make([]byte, 257))

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         longMoniker,
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		_, err := msgServer.CreateValidator(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "moniker too long")
	})

	t.Run("fails validation with description too long", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		operatorAddr := sdk.AccAddress("operator1")

		// Create a description longer than 256 characters
		longDescription := string(make([]byte, 257))

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator",
			Description:     longDescription,
			OperatorAddress: operatorAddr.String(),
		}

		_, err := msgServer.CreateValidator(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "description too long")
	})

	t.Run("fails validation with missing operator address", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator",
			Description:     "test",
			OperatorAddress: "", // Missing operator address
		}

		_, err := msgServer.CreateValidator(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing validator operator address")
	})

	t.Run("fails validation with invalid operator address", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator",
			Description:     "test",
			OperatorAddress: "invalid-address",
		}

		_, err := msgServer.CreateValidator(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "operator address is invalid")
	})

	t.Run("fails when creating validator with duplicate consensus address - same primary key", func(t *testing.T) {
		f := setupTest(t)

		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		consAddr := sdk.GetConsAddress(pubKey)
		operatorAddr1 := sdk.AccAddress("operator1")

		// Create first validator with power 0
		validator1 := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  0,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator-1",
				Description:     "first validator",
				OperatorAddress: operatorAddr1.String(),
			},
		}

		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator1, true)
		require.NoError(t, err)

		// Try to create second validator with same consensus address AND same power
		// This would have the same primary key (power, consensus_address)
		// Without HasValidator check, Set() would overwrite and allow operator address to change
		// The HasValidator check prevents this
		operatorAddr2 := sdk.AccAddress("operator2")
		validator2 := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  0, // Same power = same primary key
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator-2",
				Description:     "second validator",
				OperatorAddress: operatorAddr2.String(), // Different operator
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator2, true)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})

	t.Run("fails when creating validator with duplicate consensus address - different primary key", func(t *testing.T) {
		f := setupTest(t)

		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		consAddr := sdk.GetConsAddress(pubKey)
		operatorAddr1 := sdk.AccAddress("operator1")

		// Create first validator with power 0
		validator1 := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  0,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator-1",
				Description:     "first validator",
				OperatorAddress: operatorAddr1.String(),
			},
		}

		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator1, true)
		require.NoError(t, err)

		// Try to create second validator with same consensus address but different power
		// This will have a different primary key (power, consensus_address)
		// The HasValidator check will catch this first, but if removed,
		// the unique index on consensus address would also catch it
		operatorAddr2 := sdk.AccAddress("operator2")
		validator2 := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  100, // Different power = different primary key
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator-2",
				Description:     "second validator",
				OperatorAddress: operatorAddr2.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator2, true)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})

	t.Run("fails when creating validator with duplicate operator address", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Use same operator address for both validators
		operatorAddr := sdk.AccAddress("operator1")

		// Create first validator
		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKeyAny1 := types.UnsafePackAny(pubKey1)

		msg1 := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny1,
			Moniker:         "test-validator-1",
			Description:     "first validator",
			OperatorAddress: operatorAddr.String(),
		}

		_, err := msgServer.CreateValidator(f.ctx, msg1)
		require.NoError(t, err)

		// Try to create second validator with different pubkey but same operator address
		pubKey2 := ed25519.GenPrivKey().PubKey()
		pubKeyAny2 := types.UnsafePackAny(pubKey2)

		msg2 := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny2, // Different pubkey
			Moniker:         "test-validator-2",
			Description:     "second validator",
			OperatorAddress: operatorAddr.String(), // Same operator address
		}

		_, err = msgServer.CreateValidator(f.ctx, msg2)
		require.Error(t, err)
		// Error will be a uniqueness constraint violation on the operator address index
		require.Contains(t, err.Error(), "uniqueness constrain violation")
	})

	t.Run("rejects same key for operator and consensus", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Use same key for both operator and consensus
		sameKey := ed25519.GenPrivKey()
		operatorAddr := sdk.AccAddress(sameKey.PubKey().Address())

		pubKeyAny := types.UnsafePackAny(sameKey.PubKey())

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator",
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		_, err := msgServer.CreateValidator(f.ctx, msg)
		require.Error(t, err)
		require.ErrorIs(t, err, poatypes.ErrSameKeyForOperatorAndConsensus)
		require.Contains(t, err.Error(), "operator address and consensus pubkey must use different keys")
	})

	t.Run("accepts different keys for operator and consensus", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Use different keys
		operatorKey := ed25519.GenPrivKey()
		operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address())

		consensusKey := ed25519.GenPrivKey()
		pubKeyAny := types.UnsafePackAny(consensusKey.PubKey())

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator",
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		resp, err := msgServer.CreateValidator(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify validator was created
		consAddr := sdk.GetConsAddress(consensusKey.PubKey())
		validator, err := f.poaKeeper.GetValidator(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, "test-validator", validator.Metadata.Moniker)
	})

	t.Run("rejects pubkey type not in consensus params", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Set consensus params to only allow secp256k1 (not ed25519)
		f.ctx = f.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{"secp256k1"},
			},
		})

		// Try to create validator with ed25519 key
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		operatorAddr := sdk.AccAddress("operator-reject-type")

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator",
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		_, err := msgServer.CreateValidator(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "public key type ed25519 is not in the consensus parameters")
	})

	t.Run("accepts pubkey type in consensus params", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Set consensus params to allow ed25519
		f.ctx = f.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
		})

		// Create validator with ed25519 key
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		operatorAddr := sdk.AccAddress("operator-accept-type")

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator",
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		resp, err := msgServer.CreateValidator(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify validator was created
		consAddr := sdk.GetConsAddress(pubKey)
		validator, err := f.poaKeeper.GetValidator(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, "test-validator", validator.Metadata.Moniker)
	})

	t.Run("rejects secp256k1 pubkey when only ed25519 is allowed", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Set consensus params to only allow ed25519
		f.ctx = f.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
		})

		// Try to create validator with secp256k1 key
		consensusKey := secp256k1.GenPrivKey()
		pubKeyAny := types.UnsafePackAny(consensusKey.PubKey())
		// Use different key for operator to avoid same-key error
		operatorKey := secp256k1.GenPrivKey()
		operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address())

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator",
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		_, err := msgServer.CreateValidator(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "public key type secp256k1 is not in the consensus parameters")
	})
}

func TestMsgServerUpdateValidators(t *testing.T) {
	t.Run("successfully updates validators with valid admin", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// Create validator with proper pubkey-derived consensus address
		operatorAddr := sdk.AccAddress("operator1")
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		consAddr := sdk.GetConsAddress(pubKey)

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Update validator power
		validator.Power = 200
		msg := &poatypes.MsgUpdateValidators{
			Admin:      adminAddr,
			Validators: []poatypes.Validator{validator},
		}

		resp, err := msgServer.UpdateValidators(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify validator was updated
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(200), power)
	})

	t.Run("fails with invalid admin", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// Create validator with proper pubkey-derived consensus address
		operatorAddr := sdk.AccAddress("operator1")
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		consAddr := sdk.GetConsAddress(pubKey)

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Try to update with wrong admin
		msg := &poatypes.MsgUpdateValidators{
			Admin:      sdk.AccAddress("wrongadmin").String(),
			Validators: []poatypes.Validator{validator},
		}

		_, err = msgServer.UpdateValidators(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid authority")
	})

	t.Run("updates multiple validators at once", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// Create validators with proper pubkey-derived consensus addresses
		operatorAddr1 := sdk.AccAddress("operator1")
		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKeyAny1 := types.UnsafePackAny(pubKey1)
		consAddr1 := sdk.GetConsAddress(pubKey1)

		validator1 := poatypes.Validator{
			PubKey: pubKeyAny1,
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "validator-1",
				OperatorAddress: operatorAddr1.String(),
			},
		}

		operatorAddr2 := sdk.AccAddress("operator2")
		pubKey2 := ed25519.GenPrivKey().PubKey()
		pubKeyAny2 := types.UnsafePackAny(pubKey2)
		consAddr2 := sdk.GetConsAddress(pubKey2)

		validator2 := poatypes.Validator{
			PubKey: pubKeyAny2,
			Power:  200,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "validator-2",
				OperatorAddress: operatorAddr2.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr1, validator1, true)
		require.NoError(t, err)
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr2, validator2, true)
		require.NoError(t, err)

		// Update both validators
		validator1.Power = 300
		validator2.Power = 400
		msg := &poatypes.MsgUpdateValidators{
			Admin:      adminAddr,
			Validators: []poatypes.Validator{validator1, validator2},
		}

		resp, err := msgServer.UpdateValidators(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify both validators were updated
		power1, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr1)
		require.NoError(t, err)
		require.Equal(t, int64(300), power1)

		power2, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr2)
		require.NoError(t, err)
		require.Equal(t, int64(400), power2)

		// Verify total power
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(700), totalPower)

		// Verify events were emitted for each validator
		events := f.ctx.EventManager().Events()
		require.Len(t, events, 2)

		// Check first validator event
		require.Equal(t, poatypes.EventTypeUpdateValidator, events[0].Type)
		attrs1 := make(map[string]string)
		for _, attr := range events[0].Attributes {
			attrs1[attr.Key] = attr.Value
		}
		require.Equal(t, "300", attrs1[poatypes.AttributeKeyPower])

		// Check second validator event
		require.Equal(t, poatypes.EventTypeUpdateValidator, events[1].Type)
		attrs2 := make(map[string]string)
		for _, attr := range events[1].Attributes {
			attrs2[attr.Key] = attr.Value
		}
		require.Equal(t, "400", attrs2[poatypes.AttributeKeyPower])
	})

	t.Run("sets validator power to zero with other validators present", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// Create two validators so total power doesn't go to zero
		operatorAddr1 := sdk.AccAddress("operator1")
		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKeyAny1 := types.UnsafePackAny(pubKey1)
		consAddr1 := sdk.GetConsAddress(pubKey1)

		validator1 := poatypes.Validator{
			PubKey: pubKeyAny1,
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator-1",
				OperatorAddress: operatorAddr1.String(),
			},
		}

		operatorAddr2 := sdk.AccAddress("operator2")
		pubKey2 := ed25519.GenPrivKey().PubKey()
		pubKeyAny2 := types.UnsafePackAny(pubKey2)
		consAddr2 := sdk.GetConsAddress(pubKey2)

		validator2 := poatypes.Validator{
			PubKey: pubKeyAny2,
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator-2",
				OperatorAddress: operatorAddr2.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr1, validator1, true)
		require.NoError(t, err)
		err = f.poaKeeper.CreateValidator(f.ctx, consAddr2, validator2, true)
		require.NoError(t, err)

		// Set first validator power to zero (total power will still be 100)
		validator1.Power = 0
		msg := &poatypes.MsgUpdateValidators{
			Admin:      adminAddr,
			Validators: []poatypes.Validator{validator1},
		}

		resp, err := msgServer.UpdateValidators(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify validator power is zero
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr1)
		require.NoError(t, err)
		require.Equal(t, int64(0), power)

		// Verify total power is still 100
		totalPower, err := f.poaKeeper.GetTotalPower(f.ctx)
		require.NoError(t, err)
		require.Equal(t, int64(100), totalPower)
	})

	t.Run("fails validation with negative power", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// Create validator
		operatorAddr := sdk.AccAddress("operator1")
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		consAddr := sdk.GetConsAddress(pubKey)

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Try to update with negative power
		validator.Power = -100
		msg := &poatypes.MsgUpdateValidators{
			Admin:      adminAddr,
			Validators: []poatypes.Validator{validator},
		}

		_, err = msgServer.UpdateValidators(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "negative validator power")
	})

	t.Run("fails validation with missing operator address", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// Create validator with missing operator address
		// This will fail because the validator doesn't exist (can't look it up without valid metadata)
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: "", // Missing operator address
			},
		}

		msg := &poatypes.MsgUpdateValidators{
			Admin:      adminAddr,
			Validators: []poatypes.Validator{validator},
		}

		_, err = msgServer.UpdateValidators(f.ctx, msg)
		require.Error(t, err)
		// The error will be "unknown validator" because it can't find it, not a validation error
		require.Contains(t, err.Error(), "unknown validator")
	})

	t.Run("rejects update with operator address matching consensus key", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// Create validator with different keys initially
		operatorKey := ed25519.GenPrivKey()
		operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address())

		consensusKey := ed25519.GenPrivKey()
		pubKeyAny := types.UnsafePackAny(consensusKey.PubKey())
		consAddr := sdk.GetConsAddress(consensusKey.PubKey())

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Try to update operator address to one derived from consensus key
		badOperatorAddr := sdk.AccAddress(consensusKey.PubKey().Address())

		badValidator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  200,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "bad-validator",
				OperatorAddress: badOperatorAddr.String(),
			},
		}

		msg := &poatypes.MsgUpdateValidators{
			Admin:      adminAddr,
			Validators: []poatypes.Validator{badValidator},
		}

		_, err = msgServer.UpdateValidators(f.ctx, msg)
		require.Error(t, err)
		require.ErrorIs(t, err, poatypes.ErrSameKeyForOperatorAndConsensus)
		require.Contains(t, err.Error(), "operator address and consensus pubkey must use different keys")
	})

	t.Run("accepts update with different operator address", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// Create validator with different keys
		operatorKey := ed25519.GenPrivKey()
		operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address())

		consensusKey := ed25519.GenPrivKey()
		pubKeyAny := types.UnsafePackAny(consensusKey.PubKey())
		consAddr := sdk.GetConsAddress(consensusKey.PubKey())

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Update with a different operator address (still different from consensus key)
		newOperatorKey := ed25519.GenPrivKey()
		newOperatorAddr := sdk.AccAddress(newOperatorKey.PubKey().Address())

		updatedValidator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  200,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "updated-validator",
				OperatorAddress: newOperatorAddr.String(),
			},
		}

		msg := &poatypes.MsgUpdateValidators{
			Admin:      adminAddr,
			Validators: []poatypes.Validator{updatedValidator},
		}

		resp, err := msgServer.UpdateValidators(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify update was applied
		updated, err := f.poaKeeper.GetValidator(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(200), updated.Power)
		require.Equal(t, "updated-validator", updated.Metadata.Moniker)
	})

	t.Run("rejects update when pubkey type not in consensus params", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// First, create a validator with ed25519 when it's allowed
		f.ctx = f.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
		})

		operatorAddr := sdk.AccAddress("operator-update-type")
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		consAddr := sdk.GetConsAddress(pubKey)

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Now change consensus params to only allow secp256k1
		f.ctx = f.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{"secp256k1"},
			},
		})

		// Try to update the validator (which has ed25519 key) - should fail
		validator.Power = 200
		msg := &poatypes.MsgUpdateValidators{
			Admin:      adminAddr,
			Validators: []poatypes.Validator{validator},
		}

		_, err = msgServer.UpdateValidators(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "public key type ed25519 is not in the consensus parameters")
	})

	t.Run("accepts update when pubkey type is in consensus params", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		err := f.poaKeeper.UpdateParams(f.ctx, poatypes.Params{Admin: adminAddr})
		require.NoError(t, err)

		// Set consensus params to allow ed25519
		f.ctx = f.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
		})

		operatorAddr := sdk.AccAddress("operator-update-ok")
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny := types.UnsafePackAny(pubKey)
		consAddr := sdk.GetConsAddress(pubKey)

		validator := poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		err = f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// Update the validator power - should succeed since ed25519 is still allowed
		validator.Power = 200
		msg := &poatypes.MsgUpdateValidators{
			Admin:      adminAddr,
			Validators: []poatypes.Validator{validator},
		}

		resp, err := msgServer.UpdateValidators(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify update was applied
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(200), power)
	})
}

func TestMsgServerWithdrawFees(t *testing.T) {
	t.Run("successfully withdraws fees", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Create validator
		opAddr, _ := createValidator(t, f, 1, 100)
		opAddrSdk, err := sdk.AccAddressFromBech32(opAddr)
		require.NoError(t, err)

		// Add fees to fee collector
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Checkpoint to allocate fees
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Get initial operator balance
		initialBalance := f.bankKeeper.GetBalance(f.ctx, opAddrSdk, "stake")

		// Withdraw fees
		msg := &poatypes.MsgWithdrawFees{
			Operator: opAddr,
		}

		resp, err := msgServer.WithdrawFees(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify fees were transferred to operator
		finalBalance := f.bankKeeper.GetBalance(f.ctx, opAddrSdk, "stake")
		require.Equal(t, initialBalance.Amount.Add(math.NewInt(1000)), finalBalance.Amount)

		// Verify withdrawable fees are now zero
		feesResp, err := f.poaKeeper.WithdrawableFees(f.ctx, &poatypes.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr,
		})
		require.NoError(t, err)
		require.True(t, feesResp.Fees.Fees.IsZero())

		// Verify event was emitted
		allEvents := f.ctx.EventManager().Events()
		var withdrawEvent *sdk.Event
		for _, event := range allEvents {
			if event.Type == poatypes.EventTypeWithdrawFees {
				withdrawEvent = &event
				break
			}
		}
		require.NotNil(t, withdrawEvent, "withdraw_fees event should be emitted")
		require.Len(t, withdrawEvent.Attributes, 2)

		// Check attributes
		attrs := make(map[string]string)
		for _, attr := range withdrawEvent.Attributes {
			attrs[attr.Key] = attr.Value
		}
		require.Equal(t, opAddr, attrs[poatypes.AttributeKeyOperatorAddress])
		require.Equal(t, "1000stake", attrs[poatypes.AttributeKeyAmount])
	})

	t.Run("withdraws zero when no fees available", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Create validator
		opAddr, _ := createValidator(t, f, 1, 100)
		opAddrSdk, err := sdk.AccAddressFromBech32(opAddr)
		require.NoError(t, err)

		// Get initial operator balance
		initialBalance := f.bankKeeper.GetBalance(f.ctx, opAddrSdk, "stake")

		// Withdraw fees (should succeed but transfer nothing)
		msg := &poatypes.MsgWithdrawFees{
			Operator: opAddr,
		}

		resp, err := msgServer.WithdrawFees(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify balance unchanged
		finalBalance := f.bankKeeper.GetBalance(f.ctx, opAddrSdk, "stake")
		require.Equal(t, initialBalance.Amount, finalBalance.Amount)
	})

	t.Run("withdraws correct amount with multiple validators", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Create validators with different power
		opAddr1, _ := createValidator(t, f, 1, 100) // 25% power
		opAddr2, _ := createValidator(t, f, 2, 300) // 75% power

		opAddrSdk1, err := sdk.AccAddressFromBech32(opAddr1)
		require.NoError(t, err)
		opAddrSdk2, err := sdk.AccAddressFromBech32(opAddr2)
		require.NoError(t, err)

		// Add fees to fee collector
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Checkpoint to allocate fees
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Withdraw fees for validator 1
		msg1 := &poatypes.MsgWithdrawFees{
			Operator: opAddr1,
		}
		_, err = msgServer.WithdrawFees(f.ctx, msg1)
		require.NoError(t, err)

		// Verify validator 1 received 25% (250 stake)
		balance1 := f.bankKeeper.GetBalance(f.ctx, opAddrSdk1, "stake")
		require.Equal(t, math.NewInt(250), balance1.Amount)

		// Withdraw fees for validator 2
		msg2 := &poatypes.MsgWithdrawFees{
			Operator: opAddr2,
		}
		_, err = msgServer.WithdrawFees(f.ctx, msg2)
		require.NoError(t, err)

		// Verify validator 2 received 75% (750 stake)
		balance2 := f.bankKeeper.GetBalance(f.ctx, opAddrSdk2, "stake")
		require.Equal(t, math.NewInt(750), balance2.Amount)
	})

	t.Run("withdraws multiple denominations", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Create validator
		opAddr, _ := createValidator(t, f, 1, 100)
		opAddrSdk, err := sdk.AccAddressFromBech32(opAddr)
		require.NoError(t, err)

		// Add multiple denominations to fee collector
		fees := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 1000),
			sdk.NewInt64Coin("atom", 500),
		)
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Checkpoint to allocate fees
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Withdraw fees
		msg := &poatypes.MsgWithdrawFees{
			Operator: opAddr,
		}

		resp, err := msgServer.WithdrawFees(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify both denominations were transferred
		stakeBalance := f.bankKeeper.GetBalance(f.ctx, opAddrSdk, "stake")
		require.Equal(t, math.NewInt(1000), stakeBalance.Amount)

		atomBalance := f.bankKeeper.GetBalance(f.ctx, opAddrSdk, "atom")
		require.Equal(t, math.NewInt(500), atomBalance.Amount)

		// Verify event was emitted with correct amount
		allEvents := f.ctx.EventManager().Events()
		var withdrawEvent *sdk.Event
		for _, event := range allEvents {
			if event.Type == poatypes.EventTypeWithdrawFees {
				withdrawEvent = &event
				break
			}
		}
		require.NotNil(t, withdrawEvent, "withdraw_fees event should be emitted")
		require.Len(t, withdrawEvent.Attributes, 2)

		// Check attributes
		attrs := make(map[string]string)
		for _, attr := range withdrawEvent.Attributes {
			attrs[attr.Key] = attr.Value
		}
		require.Equal(t, opAddr, attrs[poatypes.AttributeKeyOperatorAddress])
		// Amount should contain both denominations
		amountStr := attrs[poatypes.AttributeKeyAmount]
		require.Contains(t, amountStr, "1000stake")
		require.Contains(t, amountStr, "500atom")
	})

	t.Run("fails with invalid operator address", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Try to withdraw with invalid address
		msg := &poatypes.MsgWithdrawFees{
			Operator: "invalid-address",
		}

		_, err := msgServer.WithdrawFees(f.ctx, msg)
		require.Error(t, err)
	})

	t.Run("fails validation with empty operator address", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Try to withdraw with empty address (should fail when parsing address)
		msg := &poatypes.MsgWithdrawFees{
			Operator: "",
		}

		_, err := msgServer.WithdrawFees(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty address string is not allowed")
	})

	t.Run("withdraws pending fees without checkpoint", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Create validator
		opAddr, _ := createValidator(t, f, 1, 100)
		opAddrSdk, err := sdk.AccAddressFromBech32(opAddr)
		require.NoError(t, err)

		// Add fees to fee collector (don't checkpoint)
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Withdraw fees (should still work with lazy distribution)
		msg := &poatypes.MsgWithdrawFees{
			Operator: opAddr,
		}

		resp, err := msgServer.WithdrawFees(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify fees were transferred
		balance := f.bankKeeper.GetBalance(f.ctx, opAddrSdk, "stake")
		require.Equal(t, math.NewInt(1000), balance.Amount)
	})
}
