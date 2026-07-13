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
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper maintains the POA module state and provides methods for validator management and fee distribution.
type Keeper struct {
	params collections.Item[types.Params]
	// Validators keyed by consensus address
	validators *collections.IndexedMap[sdk.ConsAddress, types.Validator, ValidatorIndexes]
	// totalPower of all validators
	totalPower collections.Item[int64]
	// totalAllocatedFees tracks the sum of all allocated fees across validators
	totalAllocatedFees collections.Item[types.ValidatorFees]
	// validatorAllocatedFees tracks per-validator allocated fees, keyed by consensus address
	validatorAllocatedFees collections.Map[string, types.ValidatorFees]
	// queuedUpdates stores pending validator updates for the block in transient store.
	// The store gets wiped every block, so this is only used for same-block updates.
	queuedUpdates collections.Vec[abci.ValidatorUpdate]
	// lastCommittedPower records the power last handed to CometBFT per consensus
	// address.
	lastCommittedPower collections.Map[sdk.ConsAddress, int64]

	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper

	cdc    codec.Codec
	schema collections.Schema
}

// ValidatorIndexes defines secondary indexes for validators to enable efficient lookups by different keys.
type ValidatorIndexes struct {
	// OperatorAddress maps operator address -> consensus address
	OperatorAddress *indexes.Unique[string, sdk.ConsAddress, types.Validator]
	// Power maps (power, consensus_address) for sorted iteration by power
	Power *indexes.Multi[int64, sdk.ConsAddress, types.Validator]
}

// IndexesList returns the list of indexes for the validators collection.
func (a ValidatorIndexes) IndexesList() []collections.Index[sdk.ConsAddress, types.Validator] {
	return []collections.Index[sdk.ConsAddress, types.Validator]{
		a.OperatorAddress,
		a.Power,
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
		// Validators keyed by consensus address
		validators: collections.NewIndexedMap(
			sb,
			types.ValidatorsKey,
			"validators",
			sdk.ConsAddressKey,
			codec.CollValue[types.Validator](cdc),
			ValidatorIndexes{
				OperatorAddress: indexes.NewUnique(
					sb,
					types.ValidatorOperatorAddressIndex,
					"validator_by_operator",
					collections.StringKey,
					sdk.ConsAddressKey,
					func(_ sdk.ConsAddress, v types.Validator) (string, error) {
						return v.Metadata.OperatorAddress, nil
					},
				),
				Power: indexes.NewMulti(
					sb,
					types.ValidatorPowerIndex,
					"validator_by_power",
					collections.Int64Key,
					sdk.ConsAddressKey,
					func(_ sdk.ConsAddress, v types.Validator) (int64, error) {
						return v.Power, nil
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
		validatorAllocatedFees: collections.NewMap(
			sb,
			types.ValidatorAllocatedFeesKey,
			"validator_allocated_fees",
			collections.StringKey,
			codec.CollValue[types.ValidatorFees](cdc),
		),
		queuedUpdates: collections.NewVec(
			tsb,
			types.QueuedUpdatesKey,
			"queued_updates",
			codec.CollValue[abci.ValidatorUpdate](cdc),
		),
		lastCommittedPower: collections.NewMap(
			sb,
			types.LastCommittedPowerKey,
			"last_committed_power",
			sdk.ConsAddressKey,
			collections.Int64Value,
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

// ReapValidatorUpdates coalesces the queued updates into the changeset for
// CometBFT, one entry per consensus address.
func (k *Keeper) ReapValidatorUpdates(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	// order preserves the first appearance order of each cons pub key so the
	// changeset is deterministic
	var order []string

	// latestUpdates records the latestUpdates validator update that should be used for a
	// cons pub key as we are walking the queued updates
	latestUpdates := make(map[string]abci.ValidatorUpdate)

	err := k.queuedUpdates.Walk(ctx, nil, func(_ uint64, queuedUpdate abci.ValidatorUpdate) (bool, error) {
		// Key by the marshaled pubkey; the struct itself isn't map-comparable.
		key, err := queuedUpdate.PubKey.Marshal()
		if err != nil {
			return true, err
		}

		id := string(key)

		// record order if we have not seen this pubkey before
		if _, seen := latestUpdates[id]; !seen {
			order = append(order, id)
		}

		// record this update the latest update for the pubkey, potentially
		// overriding another update, we use last write wins semantics here. if
		// we have updates for pubkey X, X->20 and then X->10, we only care
		// about X->10 and can disregard the X->20.
		latestUpdates[id] = queuedUpdate
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	// updates is the final update list
	updates := make([]abci.ValidatorUpdate, 0, len(order))

	// do a final past over the last recorded updates for the pubkeys being
	// updated. if an update is sending a pubkey to 0 power, and it is not an
	// active set pubkey (i.e. it has no power in lastCommittedPower), we must
	// remove it from the list of updates as this would look to Comet as if we
	// are removing a validator that does not exist, which is invalid.
	//
	// NOTE: this has to happen in a second pass over the updates since we must
	// operate on the final update for a pubkey.
	for _, id := range order {
		update := latestUpdates[id]
		if update.Power != 0 {
			updates = append(updates, update)
			continue
		}

		// zero power update, must ensure that the update is for a pubkey that
		// exists already
		consAddr, err := consAddrFromUpdate(update)
		if err != nil {
			return nil, err
		}

		committed, err := k.lastCommittedPower.Has(ctx, consAddr)
		if err != nil {
			return nil, err
		}
		if !committed {
			// comet doesnt know about this pubkey, and we are setting it to 0
			// with this update (i.e. a delete), this is an invalid operation
			// so we drop it (the pubkey is __already__ at zero power in
			// comet's eyes).
			continue
		}

		// we are setting an existing validators pubkey to 0 power, this is
		// valid, keep the update.
		updates = append(updates, update)
	}

	if len(updates) > 0 {
		ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName)).Info("applying ABCI validator updates", "total", len(updates))
	}
	return updates, nil
}

// SetLastCommittedPower updates the lastCommittedPower snapshot to reflect an
// ABCI updates changeset.
func (k *Keeper) SetLastCommittedPower(ctx sdk.Context, updates []abci.ValidatorUpdate) error {
	for _, update := range updates {
		consAddr, err := consAddrFromUpdate(update)
		if err != nil {
			return err
		}

		if update.Power == 0 {
			// remove vals when they go to 0 power instead of recording them at
			// 0 power so the collection does not grow indefinitely
			if err := k.lastCommittedPower.Remove(ctx, consAddr); err != nil {
				return err
			}
			continue
		}

		if err := k.lastCommittedPower.Set(ctx, consAddr, update.Power); err != nil {
			return err
		}
	}
	return nil
}

// consAddrFromUpdate derives the consensus address from an ABCI update's pubkey.
func consAddrFromUpdate(u abci.ValidatorUpdate) (sdk.ConsAddress, error) {
	pk, err := cryptocodec.FromCmtProtoPublicKey(u.PubKey)
	if err != nil {
		return nil, err
	}
	return sdk.GetConsAddress(pk), nil
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
