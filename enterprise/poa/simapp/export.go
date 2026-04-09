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
	"encoding/json"
	"fmt"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

func (app *SimApp) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs, modulesToExport []string) (servertypes.ExportedApp, error) {
	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})

	height := app.LastBlockHeight() + 1
	if forZeroHeight {
		height = 0
	}

	genState, err := app.ModuleManager.ExportGenesisForModules(ctx, app.encodingConfig.Codec, modulesToExport)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators, err := app.POAKeeper.GetAllValidators(ctx)
	if err != nil {
		return servertypes.ExportedApp{}, fmt.Errorf("failed to get validators from POA module: %w", err)
	}

	genesisValidators := make([]types.GenesisValidator, 0, len(validators))
	for _, v := range validators {
		if v.Power == 0 {
			continue
		}
		var pk cryptotypes.PubKey
		if err := app.InterfaceRegistry().UnpackAny(v.PubKey, &pk); err != nil {
			return servertypes.ExportedApp{}, fmt.Errorf("failed to unpack validator pubkey: %w", err)
		}
		cmtPk, err := cryptocodec.ToCmtPubKeyInterface(pk)
		if err != nil {
			return servertypes.ExportedApp{}, fmt.Errorf("failed to convert validator pubkey: %w", err)
		}
		genesisValidators = append(genesisValidators, types.GenesisValidator{
			Address: sdk.ConsAddress(cmtPk.Address()).Bytes(),
			PubKey:  cmtPk,
			Power:   v.Power,
			Name:    v.Metadata.Moniker,
		})
	}

	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      genesisValidators,
		Height:          height,
		ConsensusParams: app.GetConsensusParams(ctx),
	}, nil
}
