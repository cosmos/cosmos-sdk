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

package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/enterprise/group/x/group"
	"github.com/cosmos/cosmos-sdk/enterprise/group/x/group/keeper"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding group type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], []byte{keeper.GroupTablePrefix}):
			var groupA, groupB group.GroupInfo

			cdc.MustUnmarshal(kvA.Value, &groupA)
			cdc.MustUnmarshal(kvB.Value, &groupB)

			return fmt.Sprintf("%v\n%v", groupA, groupB)
		case bytes.Equal(kvA.Key[:1], []byte{keeper.GroupMemberTablePrefix}):
			var memberA, memberB group.GroupMember

			cdc.MustUnmarshal(kvA.Value, &memberA)
			cdc.MustUnmarshal(kvB.Value, &memberB)

			return fmt.Sprintf("%v\n%v", memberA, memberB)
		case bytes.Equal(kvA.Key[:1], []byte{keeper.GroupPolicyTablePrefix}):
			var accA, accB group.GroupPolicyInfo

			cdc.MustUnmarshal(kvA.Value, &accA)
			cdc.MustUnmarshal(kvB.Value, &accB)

			return fmt.Sprintf("%v\n%v", accA, accB)
		case bytes.Equal(kvA.Key[:1], []byte{keeper.ProposalTablePrefix}):
			var propA, propB group.Proposal

			cdc.MustUnmarshal(kvA.Value, &propA)
			cdc.MustUnmarshal(kvB.Value, &propB)

			return fmt.Sprintf("%v\n%v", propA, propB)
		case bytes.Equal(kvA.Key[:1], []byte{keeper.VoteTablePrefix}):
			var voteA, voteB group.Vote

			cdc.MustUnmarshal(kvA.Value, &voteA)
			cdc.MustUnmarshal(kvB.Value, &voteB)

			return fmt.Sprintf("%v\n%v", voteA, voteB)
		default:
			panic(fmt.Sprintf("invalid group key %X", kvA.Key))
		}
	}
}
