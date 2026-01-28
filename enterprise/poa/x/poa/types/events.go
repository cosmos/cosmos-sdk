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

// Event types for the POA module
const (
	EventTypeUpdateParams    = "update_params"
	EventTypeCreateValidator = "create_validator"
	EventTypeUpdateValidator = "update_validator"
	EventTypeWithdrawFees    = "withdraw_fees"
)

// Attribute keys for events
const (
	AttributeKeyAdmin            = "admin"
	AttributeKeyParams           = "params"
	AttributeKeyOperatorAddress  = "operator_address"
	AttributeKeyConsensusAddress = "consensus_address"
	AttributeKeyMoniker          = "moniker"
	AttributeKeyPower            = "power"
	AttributeKeyAmount           = "amount"
)
