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

package cmd

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"

	cfg "github.com/cometbft/cometbft/config"

	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// GenAppStateFromConfig gets the genesis app state from the config
func GenAppStateFromConfig(config *cfg.Config, genesis *types.AppGenesis, persistentPeers []string,
) (appState json.RawMessage, err error) {
	// process genesis transactions, else create default genesis.json
	// TODO: appGenTxs was collected here. do we need that??
	slices.Sort(persistentPeers)
	config.P2P.PersistentPeers = strings.Join(persistentPeers, ",")
	cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

	// create the app state
	appGenesisState, err := types.GenesisStateFromAppGenesis(genesis)
	if err != nil {
		return appState, err
	}

	appState, err = json.MarshalIndent(appGenesisState, "", "  ")
	if err != nil {
		return appState, err
	}

	genesis.AppState = appState
	err = ExportGenesisFile(genesis, config.GenesisFile())

	return appState, err
}

// ExportGenesisFile creates and writes the genesis configuration to disk. An
// error is returned if building or writing the configuration to file fails.
func ExportGenesisFile(genesis *types.AppGenesis, genFile string) error {
	if err := genesis.ValidateAndComplete(); err != nil {
		return err
	}

	return genesis.SaveAs(genFile)
}
