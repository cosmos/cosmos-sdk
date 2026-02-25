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

package orm

import (
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

// genTableModel generates a new table model. At the moment it doesn't
// generate empty strings for Name.
var genTableModel = rapid.Custom(func(t *rapid.T) *testdata.TableModel {
	return &testdata.TableModel{
		Id:       rapid.Uint64().Draw(t, "id"),
		Name:     rapid.StringN(1, 100, 150).Draw(t, "name"),
		Number:   rapid.Uint64().Draw(t, "number "),
		Metadata: []byte(rapid.StringN(1, 100, 150).Draw(t, "metadata")),
	}
})
