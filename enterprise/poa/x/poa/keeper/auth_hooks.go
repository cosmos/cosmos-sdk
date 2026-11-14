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
	"context"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ authtypes.PubKeyChangeHooks = AuthHooks{}

// AuthHooks protects PoA invariants when an operator changes its account key.
type AuthHooks struct {
	k *Keeper
}

// NewAuthHooks returns the PoA auth hooks.
func (k *Keeper) NewAuthHooks() authtypes.PubKeyChangeHooks {
	return AuthHooks{k: k}
}

// BeforePubKeyChange rejects changing a validator operator account to its
// consensus key.
func (h AuthHooks) BeforePubKeyChange(
	ctx context.Context,
	address sdk.AccAddress,
	_, newPubKey cryptotypes.PubKey,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	validator, err := h.k.GetValidatorByOperatorAddress(sdkCtx, address)
	if errors.Is(err, collections.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	var consensusPubKey cryptotypes.PubKey
	if err := h.k.cdc.UnpackAny(validator.PubKey, &consensusPubKey); err != nil {
		return err
	}
	if consensusPubKey.Equals(newPubKey) {
		return errorsmod.Wrapf(
			types.ErrSameKeyForOperatorAndConsensus,
			"operator account %s cannot change to its consensus pubkey",
			address,
		)
	}
	return nil
}
