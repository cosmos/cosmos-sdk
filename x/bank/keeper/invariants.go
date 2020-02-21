package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// RegisterInvariants registers the bank module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, bk ViewKeeper) {
	ir.RegisterRoute(types.ModuleName, "nonnegative-outstanding",
		NonnegativeBalanceInvariant(bk))
}

// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
func NonnegativeBalanceInvariant(bk ViewKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		bk.IterateAllBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
			if balance.IsNegative() {
				count++
				msg += fmt.Sprintf("\t%s has a negative balance of %s\n", addr, balance)
			}

			return false
		})

		broken := count != 0

		return sdk.FormatInvariant(
			types.ModuleName, "nonnegative-outstanding",
			fmt.Sprintf("amount of negative balances found %d\n%s", count, msg),
		), broken
	}
}
