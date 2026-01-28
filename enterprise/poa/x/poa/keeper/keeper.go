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
	"slices"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/store"
	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper maintains the POA module state and provides methods for validator management and fee distribution.
type Keeper struct {
	params collections.Item[types.Params]
	// Validators indexed by (power, consensus_address) for sorted iteration
	validators *collections.IndexedMap[collections.Pair[int64, string], types.Validator, ValidatorIndexes]
	// totalPower of all validators
	totalPower collections.Item[int64]
	// totalAllocatedFees tracks the sum of all allocated fees across validators
	totalAllocatedFees collections.Item[types.ValidatorFees]
	// queuedUpdates stores pending validator updates for the block in transient store.
	// The store gets wiped every block, so this is only used for same-block udpates.
	queuedUpdates collections.Vec[abci.ValidatorUpdate]

	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper

	cdc    codec.Codec
	schema collections.Schema
}

// ValidatorIndexes defines secondary indexes for validators to enable efficient lookups by different keys.
type ValidatorIndexes struct {
	// ConsensusAddress maps consensus address -> (power, consensus_address)
	ConsensusAddress *indexes.Unique[string, collections.Pair[int64, string], types.Validator]
	// OperatorAddress maps operator address -> (power, consensus_address)
	OperatorAddress *indexes.Unique[string, collections.Pair[int64, string], types.Validator]
}

// IndexesList returns the list of indexes for the validators collection.
func (a ValidatorIndexes) IndexesList() []collections.Index[collections.Pair[int64, string], types.Validator] {
	return []collections.Index[collections.Pair[int64, string], types.Validator]{
		a.ConsensusAddress,
		a.OperatorAddress,
	}
}

// NewKeeper creates a new POA Keeper instance with the provided dependencies.
func NewKeeper(cdc codec.Codec, storeService store.KVStoreService, transientStoreService store.TransientStoreService, authKeeper types.AccountKeeper, bankKeeper types.BankKeeper) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	tsb := collections.NewSchemaBuilder(types.TransientStoreService{TransientStoreService: transientStoreService})

	k := &Keeper{
		params: collections.NewItem(
			sb,
			types.ParamsKey,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		// Validators indexed by (power, consensus_address) for sorted iteration
		validators: collections.NewIndexedMap(
			sb,
			types.ValidatorsKey,
			"validators",
			collections.PairKeyCodec(collections.Int64Key, collections.StringKey),
			codec.CollValue[types.Validator](cdc),
			ValidatorIndexes{
				ConsensusAddress: indexes.NewUnique(
					sb,
					types.ValidatorConsensusAddressIndex,
					"validator_by_consensus",
					collections.StringKey,
					collections.PairKeyCodec(collections.Int64Key, collections.StringKey),
					func(key collections.Pair[int64, string], v types.Validator) (string, error) {
						// Extract consensus address from composite key
						return key.K2(), nil
					},
				),
				OperatorAddress: indexes.NewUnique(
					sb,
					types.ValidatorOperatorAddressIndex,
					"validator_by_operator",
					collections.StringKey,
					collections.PairKeyCodec(collections.Int64Key, collections.StringKey),
					func(_ collections.Pair[int64, string], v types.Validator) (string, error) {
						return v.Metadata.OperatorAddress, nil
					},
				),
			},
		),
		totalPower: collections.NewItem(
			sb,
			types.TotalPowerKey,
			"total_power",
			collections.Int64Value,
		),
		totalAllocatedFees: collections.NewItem(
			sb,
			types.TotalAllocatedKey,
			"total_allocated",
			codec.CollValue[types.ValidatorFees](cdc),
		),
		queuedUpdates: collections.NewVec(
			tsb,
			types.QueuedUpdatesKey,
			"queued_updates",
			codec.CollValue[abci.ValidatorUpdate](cdc),
		),
		authKeeper: authKeeper,
		bankKeeper: bankKeeper,
		cdc:        cdc,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.schema = schema

	return k
}

// Admin returns the current admin address from module parameters.
func (k *Keeper) Admin(ctx sdk.Context) (string, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return "", err
	}

	return params.Admin, nil
}

// ReapValidatorUpdates returns and clears all queued validator updates for CometBFT.
func (k *Keeper) ReapValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	var updates []abci.ValidatorUpdate
	err := k.queuedUpdates.Walk(ctx, nil, func(_ uint64, value abci.ValidatorUpdate) (bool, error) {
		updates = append(updates, value)
		return false, nil
	})
	if err != nil {
		return []abci.ValidatorUpdate{}
	}

	if len(updates) > 0 {
		ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName)).Info("applying ABCI validator updates", "total", len(updates))
	}
	return updates
}

// GetParams retrieves the current module parameters.
func (k *Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	return k.params.Get(ctx)
}

// UpdateParams updates the module parameters with the provided values.
func (k *Keeper) UpdateParams(ctx sdk.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}

// ValidateOperatorAndConsensusPubKeyDifferent validates that the operator address
// and consensus pubkey are derived from different keys.
//
// This is critical when secp256k1 is enabled for consensus, as users might accidentally
// use their account key for consensus signing, which would be a security risk.
//
// Returns an error if:
// - The operator address is invalid
// - The pubkey cannot be unpacked
// - The operator address derives from the same key as the consensus pubkey
func (k *Keeper) ValidateOperatorAndConsensusPubKeyDifferent(operatorAddress string, pubKeyAny *codectypes.Any) error {
	operatorAddr, err := sdk.AccAddressFromBech32(operatorAddress)
	if err != nil {
		return errors.Wrap(err, "invalid operator address")
	}

	var pubKey cryptotypes.PubKey
	if err := k.cdc.UnpackAny(pubKeyAny, &pubKey); err != nil {
		return err
	}

	// Derive account address from consensus pubkey and compare
	accountAddrFromConsPubkey := sdk.AccAddress(pubKey.Address())
	if accountAddrFromConsPubkey.Equals(operatorAddr) {
		return errors.Wrapf(
			types.ErrSameKeyForOperatorAndConsensus,
			"operator address %s derives from the consensus pubkey - these must be different keys",
			operatorAddress,
		)
	}

	return nil
}

func (k *Keeper) validatePubkeyType(ctx sdk.Context, key cryptotypes.PubKey) error {
	validPkTypes := ctx.ConsensusParams().Validator.GetPubKeyTypes()

	if !slices.Contains(validPkTypes, key.Type()) {
		return fmt.Errorf("public key type %s is not in the consensus parameters validator public key types: [%s]", key.Type(), strings.Join(validPkTypes, ","))
	}
	return nil
}
