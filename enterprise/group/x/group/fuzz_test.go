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

package group

import (
	"testing"
)

func FuzzTallyResultGetCounts(f *testing.F) {
	// Seed with valid TallyResult strings
	f.Add("0", "0", "0", "0")
	f.Add("1", "0", "0", "0")
	f.Add("1", "1", "1", "1")
	f.Add("1000000000000000000", "0", "0", "0")

	f.Fuzz(func(t *testing.T, yes, no, abstain, veto string) {
		tr := TallyResult{
			YesCount:        yes,
			NoCount:         no,
			AbstainCount:    abstain,
			NoWithVetoCount: veto,
		}
		_, _ = tr.GetYesCount()
		_, _ = tr.GetNoCount()
		_, _ = tr.GetAbstainCount()
		_, _ = tr.GetNoWithVetoCount()
	})
}

func FuzzTallyResultTotalCounts(f *testing.F) {
	f.Add("0", "0", "0", "0")
	f.Add("1", "1", "1", "1")

	f.Fuzz(func(t *testing.T, yes, no, abstain, veto string) {
		tr := TallyResult{
			YesCount:        yes,
			NoCount:         no,
			AbstainCount:    abstain,
			NoWithVetoCount: veto,
		}
		_, _ = tr.TotalCounts()
	})
}

func FuzzVoteOptionFromString(f *testing.F) {
	f.Add("VOTE_OPTION_YES")
	f.Add("VOTE_OPTION_NO")
	f.Add("VOTE_OPTION_ABSTAIN")
	f.Add("VOTE_OPTION_NO_WITH_VETO")
	f.Add("VOTE_OPTION_UNSPECIFIED")

	f.Fuzz(func(t *testing.T, s string) {
		_, _ = VoteOptionFromString(s)
	})
}
