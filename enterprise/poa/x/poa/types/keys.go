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
// See ./enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package types

import "cosmossdk.io/collections"

const (
	ModuleName        = "poa"
	StoreKey          = ModuleName
	TransientStoreKey = "transient_" + StoreKey
)

var (
	ParamsKey                      = collections.NewPrefix(0)
	ValidatorsKey                  = collections.NewPrefix(1)
	ValidatorConsensusAddressIndex = collections.NewPrefix(2)
	ValidatorOperatorAddressIndex  = collections.NewPrefix(3)
	TotalPowerKey                  = collections.NewPrefix(4)
	TotalAllocatedKey              = collections.NewPrefix(5)
	QueuedUpdatesKey               = collections.NewPrefix(6)
)
