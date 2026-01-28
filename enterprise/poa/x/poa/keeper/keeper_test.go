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
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
)

func TestAdmin(t *testing.T) {
	f := setupTest(t)

	t.Run("when params are not set", func(t *testing.T) {
		// Test getting admin when params are not set (should return error)
		_, err := f.poaKeeper.Admin(f.ctx)
		require.Error(t, err)
	})

	t.Run("when params are set", func(t *testing.T) {
		// Set params with admin address
		adminAddr := sdk.AccAddress("admin")
		params := poatypes.Params{
			Admin: adminAddr.String(),
		}
		err := f.poaKeeper.UpdateParams(f.ctx, params)
		require.NoError(t, err)

		// Test getting admin address
		admin, err := f.poaKeeper.Admin(f.ctx)
		require.NoError(t, err)
		require.Equal(t, adminAddr.String(), admin)
	})
}

func TestReapValidatorUpdates(t *testing.T) {
	f := setupTest(t)

	t.Run("initially empty", func(t *testing.T) {
		// Initially, there should be no updates
		updates := f.poaKeeper.ReapValidatorUpdates(f.ctx)
		require.Empty(t, updates)
	})

	t.Run("after creating validator", func(t *testing.T) {
		// Create a validator with power > 0 to generate an update
		// CreateValidator queues an update when power > 0
		pubKey := ed25519.GenPrivKey().PubKey()
		consAddr := sdk.GetConsAddress(pubKey)
		operatorAddr := sdk.AccAddress("operator1")

		validator := poatypes.Validator{
			PubKey: codectypes.UnsafePackAny(pubKey),
			Power:  100,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         "test-validator",
				OperatorAddress: operatorAddr.String(),
			},
		}

		err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
		require.NoError(t, err)

		// ReapValidatorUpdates should return the queued updates
		updates := f.poaKeeper.ReapValidatorUpdates(f.ctx)
		require.Len(t, updates, 1)
		require.Equal(t, int64(100), updates[0].Power)
	})
}

func TestGetParams(t *testing.T) {
	f := setupTest(t)

	t.Run("when params are not set", func(t *testing.T) {
		// Test getting params when not set (should return error)
		_, err := f.poaKeeper.GetParams(f.ctx)
		require.Error(t, err)
	})

	t.Run("when params are set", func(t *testing.T) {
		// Set params
		adminAddr := sdk.AccAddress("admin")
		params := poatypes.Params{
			Admin: adminAddr.String(),
		}
		err := f.poaKeeper.UpdateParams(f.ctx, params)
		require.NoError(t, err)

		// Test getting params
		retrievedParams, err := f.poaKeeper.GetParams(f.ctx)
		require.NoError(t, err)
		require.Equal(t, params.Admin, retrievedParams.Admin)
	})
}

func TestUpdateParams(t *testing.T) {
	f := setupTest(t)

	t.Run("initial update", func(t *testing.T) {
		// Test updating params
		adminAddr1 := sdk.AccAddress("admin1")
		params1 := poatypes.Params{
			Admin: adminAddr1.String(),
		}
		err := f.poaKeeper.UpdateParams(f.ctx, params1)
		require.NoError(t, err)

		// Verify params were set
		retrievedParams, err := f.poaKeeper.GetParams(f.ctx)
		require.NoError(t, err)
		require.Equal(t, params1.Admin, retrievedParams.Admin)
	})

	t.Run("updating again with different values", func(t *testing.T) {
		// Test updating params again with different values
		adminAddr2 := sdk.AccAddress("admin2")
		params2 := poatypes.Params{
			Admin: adminAddr2.String(),
		}
		err := f.poaKeeper.UpdateParams(f.ctx, params2)
		require.NoError(t, err)

		// Verify params were updated
		retrievedParams, err := f.poaKeeper.GetParams(f.ctx)
		require.NoError(t, err)
		require.Equal(t, params2.Admin, retrievedParams.Admin)

		// Verify the old admin is different
		adminAddr1 := sdk.AccAddress("admin1")
		require.NotEqual(t, adminAddr1.String(), retrievedParams.Admin)
	})
}

func TestValidatePubkeyType(t *testing.T) {
	t.Run("accepts ed25519 key when ed25519 is in consensus params", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)
		operatorAddr := sdk.AccAddress("operator-ed25519")

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator-ed25519",
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		resp, err := msgServer.CreateValidator(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("accepts secp256k1 key when secp256k1 is in consensus params", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Set consensus params to allow secp256k1
		f.ctx = f.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{"secp256k1"},
			},
		})

		consensusKey := secp256k1.GenPrivKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(consensusKey.PubKey())
		require.NoError(t, err)
		// Use a different key for operator to avoid same-key error
		operatorKey := secp256k1.GenPrivKey()
		operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address())

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator-secp256k1",
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		resp, err := msgServer.CreateValidator(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("accepts both key types when both are in consensus params", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Set consensus params to allow both ed25519 and secp256k1
		f.ctx = f.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{"ed25519", "secp256k1"},
			},
		})

		// Create validator with ed25519 key
		ed25519Key := ed25519.GenPrivKey()
		ed25519PubKeyAny, err := codectypes.NewAnyWithValue(ed25519Key.PubKey())
		require.NoError(t, err)
		ed25519OperatorAddr := sdk.AccAddress("operator-ed25519-both")

		msg1 := &poatypes.MsgCreateValidator{
			PubKey:          ed25519PubKeyAny,
			Moniker:         "test-validator-ed25519-both",
			Description:     "test",
			OperatorAddress: ed25519OperatorAddr.String(),
		}

		resp1, err := msgServer.CreateValidator(f.ctx, msg1)
		require.NoError(t, err)
		require.NotNil(t, resp1)

		// Create validator with secp256k1 key
		secp256k1ConsKey := secp256k1.GenPrivKey()
		secp256k1PubKeyAny, err := codectypes.NewAnyWithValue(secp256k1ConsKey.PubKey())
		require.NoError(t, err)
		// Use a different key for operator to avoid same-key error
		secp256k1OperatorKey := secp256k1.GenPrivKey()
		secp256k1OperatorAddr := sdk.AccAddress(secp256k1OperatorKey.PubKey().Address())

		msg2 := &poatypes.MsgCreateValidator{
			PubKey:          secp256k1PubKeyAny,
			Moniker:         "test-validator-secp256k1-both",
			Description:     "test",
			OperatorAddress: secp256k1OperatorAddr.String(),
		}

		resp2, err := msgServer.CreateValidator(f.ctx, msg2)
		require.NoError(t, err)
		require.NotNil(t, resp2)
	})

	t.Run("rejects ed25519 key when only secp256k1 is in consensus params", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Set consensus params to only allow secp256k1
		f.ctx = f.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{"secp256k1"},
			},
		})

		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)
		operatorAddr := sdk.AccAddress("operator-reject-ed25519")

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator-reject",
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		_, err = msgServer.CreateValidator(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "public key type ed25519 is not in the consensus parameters")
	})

	t.Run("rejects secp256k1 key when only ed25519 is in consensus params", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Set consensus params to only allow ed25519
		f.ctx = f.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
		})

		consensusKey := secp256k1.GenPrivKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(consensusKey.PubKey())
		require.NoError(t, err)
		// Use a different key for operator to avoid same-key error
		operatorKey := secp256k1.GenPrivKey()
		operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address())

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator-reject-secp",
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		_, err = msgServer.CreateValidator(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "public key type secp256k1 is not in the consensus parameters")
	})

	t.Run("rejects key when consensus params has empty pubkey types", func(t *testing.T) {
		f := setupTest(t)
		msgServer := NewMsgServer(f.poaKeeper)

		// Set consensus params with no allowed pubkey types
		f.ctx = f.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{},
			},
		})

		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)
		operatorAddr := sdk.AccAddress("operator-empty-types")

		msg := &poatypes.MsgCreateValidator{
			PubKey:          pubKeyAny,
			Moniker:         "test-validator-empty",
			Description:     "test",
			OperatorAddress: operatorAddr.String(),
		}

		_, err = msgServer.CreateValidator(f.ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "public key type ed25519 is not in the consensus parameters")
	})
}

func TestValidateOperatorAndConsensusPubKeyDifferent(t *testing.T) {
	f := setupTest(t)

	t.Run("accepts different keys - ed25519 consensus, secp256k1 operator", func(t *testing.T) {
		operatorKey := secp256k1.GenPrivKey()
		operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address()).String()

		consensusKey := ed25519.GenPrivKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(consensusKey.PubKey())
		require.NoError(t, err)

		err = f.poaKeeper.ValidateOperatorAndConsensusPubKeyDifferent(operatorAddr, pubKeyAny)
		require.NoError(t, err)
	})

	t.Run("accepts different keys - secp256k1 consensus, different secp256k1 operator", func(t *testing.T) {
		operatorKey := secp256k1.GenPrivKey()
		operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address()).String()

		consensusKey := secp256k1.GenPrivKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(consensusKey.PubKey())
		require.NoError(t, err)

		err = f.poaKeeper.ValidateOperatorAndConsensusPubKeyDifferent(operatorAddr, pubKeyAny)
		require.NoError(t, err)
	})

	t.Run("rejects same secp256k1 key for operator and consensus", func(t *testing.T) {
		sameKey := secp256k1.GenPrivKey()
		operatorAddr := sdk.AccAddress(sameKey.PubKey().Address()).String()

		pubKeyAny, err := codectypes.NewAnyWithValue(sameKey.PubKey())
		require.NoError(t, err)

		err = f.poaKeeper.ValidateOperatorAndConsensusPubKeyDifferent(operatorAddr, pubKeyAny)
		require.Error(t, err)
		require.ErrorIs(t, err, poatypes.ErrSameKeyForOperatorAndConsensus)
		require.Contains(t, err.Error(), "operator address and consensus pubkey must use different keys")
	})

	t.Run("rejects same ed25519 key for operator and consensus", func(t *testing.T) {
		sameKey := ed25519.GenPrivKey()
		operatorAddr := sdk.AccAddress(sameKey.PubKey().Address()).String()

		pubKeyAny, err := codectypes.NewAnyWithValue(sameKey.PubKey())
		require.NoError(t, err)

		err = f.poaKeeper.ValidateOperatorAndConsensusPubKeyDifferent(operatorAddr, pubKeyAny)
		require.Error(t, err)
		require.ErrorIs(t, err, poatypes.ErrSameKeyForOperatorAndConsensus)
		require.Contains(t, err.Error(), "operator address and consensus pubkey must use different keys")
	})

	t.Run("returns error for invalid operator address", func(t *testing.T) {
		consensusKey := ed25519.GenPrivKey()
		pubKeyAny, err := codectypes.NewAnyWithValue(consensusKey.PubKey())
		require.NoError(t, err)

		err = f.poaKeeper.ValidateOperatorAndConsensusPubKeyDifferent("invalid-address", pubKeyAny)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid operator address")
	})
}
