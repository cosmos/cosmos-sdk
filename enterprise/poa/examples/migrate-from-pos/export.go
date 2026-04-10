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
	cmttypes "github.com/cometbft/cometbft/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	validators, err := writeValidators(ctx, app)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          height,
		ConsensusParams: app.GetConsensusParams(ctx),
	}, nil
}

// writeValidators exports POA validators as CometBFT genesis validators.
func writeValidators(ctx sdk.Context, app *SimApp) ([]cmttypes.GenesisValidator, error) {
	poaVals, err := app.POAKeeper.GetAllValidators(ctx)
	if err != nil {
		return nil, fmt.Errorf("get POA validators: %w", err)
	}

	vals := make([]cmttypes.GenesisValidator, 0, len(poaVals))
	for _, v := range poaVals {
		var pk cryptotypes.PubKey
		if err := app.encodingConfig.InterfaceRegistry.UnpackAny(v.PubKey, &pk); err != nil {
			return nil, fmt.Errorf("unpack pubkey for validator %s: %w", v.Metadata.GetMoniker(), err)
		}

		cmtPk, err := cryptocodec.ToCmtPubKeyInterface(pk)
		if err != nil {
			return nil, fmt.Errorf("convert pubkey for validator %s: %w", v.Metadata.GetMoniker(), err)
		}

		vals = append(vals, cmttypes.GenesisValidator{
			Address: sdk.ConsAddress(cmtPk.Address()).Bytes(),
			PubKey:  cmtPk,
			Power:   v.Power,
			Name:    v.Metadata.GetMoniker(),
		})
	}
	return vals, nil
}
