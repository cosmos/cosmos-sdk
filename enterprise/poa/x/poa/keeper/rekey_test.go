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

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestCreateValidatorRejectsCurrentOperatorKeyAsConsensusKey(t *testing.T) {
	f := setupTest(t)

	originalOperatorKey := secp256k1.GenPrivKey()
	consensusKey := ed25519.GenPrivKey()
	operatorAddr := sdk.AccAddress(originalOperatorKey.PubKey().Address())
	operator := authtypes.NewBaseAccount(operatorAddr, consensusKey.PubKey(), 0, 0)
	f.authKeeper.SetAccount(f.ctx, operator)

	validator := poatypes.Validator{
		PubKey: codectypes.UnsafePackAny(consensusKey.PubKey()),
		Power:  1,
		Metadata: &poatypes.ValidatorMetadata{
			Moniker:         "rekeyed-operator",
			OperatorAddress: operatorAddr.String(),
		},
	}
	err := f.poaKeeper.CreateValidator(f.ctx, sdk.GetConsAddress(consensusKey.PubKey()), validator, false)
	require.ErrorIs(t, err, poatypes.ErrSameKeyForOperatorAndConsensus)
}

func TestOperatorCannotRekeyToConsensusKey(t *testing.T) {
	f := setupTest(t)
	ctx := f.ctx.WithChainID("poa-rekey-test")

	operatorKey := secp256k1.GenPrivKey()
	consensusKey := ed25519.GenPrivKey()
	operatorAddr := sdk.AccAddress(operatorKey.PubKey().Address())
	operator := authtypes.NewBaseAccount(operatorAddr, operatorKey.PubKey(), 0, 0)
	f.authKeeper.SetAccount(ctx, operator)

	validator := poatypes.Validator{
		PubKey: codectypes.UnsafePackAny(consensusKey.PubKey()),
		Power:  1,
		Metadata: &poatypes.ValidatorMetadata{
			Moniker:         "operator",
			OperatorAddress: operatorAddr.String(),
		},
	}
	require.NoError(t, f.poaKeeper.CreateValidator(ctx, sdk.GetConsAddress(consensusKey.PubKey()), validator, false))

	params := authtypes.DefaultParams()
	params.PubKeyChangeEnabled = true
	require.NoError(t, f.authKeeper.Params.Set(ctx, params))

	newPkAny, err := codectypes.NewAnyWithValue(consensusKey.PubKey())
	require.NoError(t, err)
	doc := authtypes.ChangePubKeyProofDoc{
		ChainId:       ctx.ChainID(),
		AccountNumber: operator.GetAccountNumber(),
		Address:       operatorAddr.String(),
		NewPubKey:     newPkAny,
	}
	signBytes, err := authtypes.ChangePubKeyProofSignBytes(authcodec.NewBech32Codec("cosmos"), doc, consensusKey.PubKey())
	require.NoError(t, err)
	sig, err := consensusKey.Sign(signBytes)
	require.NoError(t, err)
	proof, err := signing.SignatureDataToProto(&signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		Signature: sig,
	}).Marshal()
	require.NoError(t, err)

	_, err = authkeeper.NewMsgServerImpl(f.authKeeper).ChangePubKey(ctx, &authtypes.MsgChangePubKey{
		Address:   operatorAddr.String(),
		NewPubKey: newPkAny,
		Proof:     proof,
	})
	require.ErrorIs(t, err, poatypes.ErrSameKeyForOperatorAndConsensus)
	require.True(t, f.authKeeper.GetAccount(ctx, operatorAddr).GetPubKey().Equals(operatorKey.PubKey()))
}
