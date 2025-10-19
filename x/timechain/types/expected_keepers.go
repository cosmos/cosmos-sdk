// SPDX-License-Identifier: MPL-2.0
// Copyright © 2025 Timechain-Arweave-LunCoSim Contributors

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/x/auth/types"
)

// AccountKeeper defines the expected account keeper used for simulations (no-op)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}
