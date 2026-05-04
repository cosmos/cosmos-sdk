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

package simapp

import (
	samples "github.com/cosmos/cosmos-sdk/enterprise/poa/examples/migrate-from-pos/sample_upgrades"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// UpgradeName is the on-chain upgrade name for POS → POA migration.
const UpgradeName = samples.StandaloneUpgradeName

func (app *SimApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName,
		samples.NewPOSToPOAUpgradeHandler(
			app.AppCodec(),
			app.AccountKeeper,
			*app.StakingKeeper,
			app.BankKeeper,
			app.DistrKeeper,
			app.GovKeeper,
			app.POAKeeper,
		),
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			// POA store is new — the pre-upgrade POS binary doesn't have it.
			Added: []string{poatypes.StoreKey},
			// Staking/distribution/slashing stores are kept mounted so the
			// upgrade handler can read from them during migration.
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
