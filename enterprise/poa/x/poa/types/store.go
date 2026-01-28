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

import (
	"context"

	"cosmossdk.io/core/store"
)

// TransientStoreService adapts an SDK TransientStoreService to a KVStoreService
// for use with collections.NewSchemaBuilder
type TransientStoreService struct {
	store.TransientStoreService
}

func (t TransientStoreService) OpenKVStore(ctx context.Context) store.KVStore {
	return t.OpenTransientStore(ctx)
}
