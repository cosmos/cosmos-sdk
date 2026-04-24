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

import (
	context "context"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountKeeper interface {
	AddressCodec() address.Codec

	// NewAccount returns a new account with the next account number. Does not save the new account to the store.
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI

	// GetAccount retrieves an account from the store.
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI

	// SetAccount sets an account in the store.
	SetAccount(context.Context, sdk.AccountI)

	// RemoveAccount Remove an account in the store.
	RemoveAccount(ctx context.Context, acc sdk.AccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}
