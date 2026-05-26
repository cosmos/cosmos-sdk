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

package group

import "time"

// Config is a config struct used for initializing the group module to avoid using globals.
type Config struct {
	// MaxExecutionPeriod defines the max duration after a proposal's voting
	// period ends that members can send a MsgExec to execute the proposal.
	MaxExecutionPeriod time.Duration
	// MaxMetadataLen defines the max length of the metadata bytes field for various entities within the group module. Defaults to 255 if not explicitly set.
	MaxMetadataLen uint64
}

// DefaultConfig returns the default config for group.
func DefaultConfig() Config {
	return Config{
		MaxExecutionPeriod: 2 * time.Hour * 24 * 7, // Two weeks.
		MaxMetadataLen:     255,
	}
}
