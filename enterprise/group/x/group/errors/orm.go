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

// ormCodespace is the codespace for all errors defined in orm package
const ormCodespace = "legacy_orm"

var (
	// ErrORMIteratorDone defines an error when an iterator is done
	ErrORMIteratorDone = errors.Register(ormCodespace, 11, "iterator done")

	// ErrORMInvalidIterator defines an error for an invalid iterator
	ErrORMInvalidIterator = errors.Register(ormCodespace, 12, "invalid iterator")

	// ErrORMUniqueConstraint defines an error when a value already exists at a given key
	ErrORMUniqueConstraint = errors.Register(ormCodespace, 13, "unique constraint violation")

	// ErrORMInvalidArgument defines an error when an invalid argument is provided as part of ORM functions
	ErrORMInvalidArgument = errors.Register(ormCodespace, 14, "invalid argument")

	// ErrORMKeyMaxLength defines an error when a key exceeds max length
	ErrORMKeyMaxLength = errors.Register(ormCodespace, 15, "key exceeds max length")

	// ErrORMEmptyKey defines an error for an empty key
	ErrORMEmptyKey = errors.Register(ormCodespace, 47, "cannot use empty key")
)
