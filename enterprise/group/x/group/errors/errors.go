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
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/group/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package errors

import "cosmossdk.io/errors"

// groupCodespace is the codespace for all errors defined in group package
const groupCodespace = "group"

var (
	ErrEmpty        = errors.Register(groupCodespace, 2, "value is empty")
	ErrDuplicate    = errors.Register(groupCodespace, 3, "duplicate value")
	ErrMaxLimit     = errors.Register(groupCodespace, 4, "limit exceeded")
	ErrType         = errors.Register(groupCodespace, 5, "invalid type")
	ErrInvalid      = errors.Register(groupCodespace, 6, "invalid value")
	ErrUnauthorized = errors.Register(groupCodespace, 7, "unauthorized")
	ErrModified     = errors.Register(groupCodespace, 8, "modified")
	ErrExpired      = errors.Register(groupCodespace, 9, "expired")
)
