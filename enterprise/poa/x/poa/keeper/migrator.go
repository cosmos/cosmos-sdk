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
	v2 "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/migrations/v2"
	v3 "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/migrations/v3"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrator handles in-place store migrations for the POA module.
type Migrator struct {
	keeper *Keeper
}

// NewMigrator returns a new Migrator instance.
func NewMigrator(keeper *Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates the POA module state from consensus version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.Migrate(ctx, m.keeper.validators.Indexes.Power, m.keeper.lastCommittedPower)
}

// Migrate2to3 migrates the POA module state from consensus version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	moduleAddr := m.keeper.authKeeper.GetModuleAddress(types.ModuleName)
	balance := sdk.NewDecCoinsFromCoins(m.keeper.bankKeeper.GetAllBalances(ctx, moduleAddr)...)
	return v3.Migrate(ctx, balance, m.keeper.validatorAllocatedFees, m.keeper.totalAllocatedFees)
}
