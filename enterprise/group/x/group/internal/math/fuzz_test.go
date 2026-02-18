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

//go:build gofuzz || go1.18

package math

import (
	"testing"
)

func FuzzNewDecFromString(f *testing.F) {
	f.Add("0")
	f.Add("1")
	f.Add("1.5")
	f.Add("-1")
	f.Add("0.000000000000000001")
	f.Add("99999999999999999999999999999999999999")

	f.Fuzz(func(t *testing.T, s string) {
		_, _ = NewDecFromString(s)
	})
}

func FuzzNewPositiveDecFromString(f *testing.F) {
	f.Add("1")
	f.Add("1.5")
	f.Add("0.000000000000000001")

	f.Fuzz(func(t *testing.T, s string) {
		_, _ = NewPositiveDecFromString(s)
	})
}

func FuzzNewNonNegativeDecFromString(f *testing.F) {
	f.Add("0")
	f.Add("1")
	f.Add("1.5")
	f.Add("0.000000000000000001")

	f.Fuzz(func(t *testing.T, s string) {
		_, _ = NewNonNegativeDecFromString(s)
	})
}
