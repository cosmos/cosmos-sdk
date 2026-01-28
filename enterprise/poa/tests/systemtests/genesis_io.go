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

//go:build system_test

package systemtests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"

	"cosmossdk.io/systemtests"
)

// SetConsensusValidatorPubKeyTypes sets the allowed validator public key types in consensus params
func SetConsensusValidatorPubKeyTypes(t *testing.T, pubKeyTypes []string) systemtests.GenesisMutator {
	t.Helper()
	return func(genesis []byte) []byte {
		// Create JSON array string from pubKeyTypes
		jsonArray := "["
		for i, keyType := range pubKeyTypes {
			if i > 0 {
				jsonArray += ","
			}
			jsonArray += fmt.Sprintf(`"%s"`, keyType)
		}
		jsonArray += "]"

		state, err := sjson.SetRaw(string(genesis), "consensus.params.validator.pub_key_types", jsonArray)
		require.NoError(t, err)
		return []byte(state)
	}
}
