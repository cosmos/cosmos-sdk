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

package simapp

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/msgguard"
)

// TestMsgServersCheckCircuitBreakerFirst calls every msg server method this
// app registers directly, as an EVM precompile calls them, and requires each
// one to check the circuit breaker before anything else. Only the msgs
// x/safety always allows may skip the check. A new msg server method without
// the guard fails here.
func TestMsgServersCheckCircuitBreakerFirst(t *testing.T) {
	// The sweep needs only the module manager and router, not a chain.
	app, _ := setup(false, 0)
	msgguard.RequireAppGuarded(t, app.AppCodec(), app.ModuleManager, app.InterfaceRegistry(), app.MsgServiceRouter())
}
