package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
)

// register bank invariants
func RegisterInvariants(ir sdk.InvariantRegistry, ak types.AccountKeeper) {
	ir.RegisterRoute(types.ModuleName, "nonnegative-outstanding",
		NonnegativeBalanceInvariant(ak))
}

// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
func NonnegativeBalanceInvariant(ak types.AccountKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var msg string
		var amt int

		accts := ak.GetAllAccounts(ctx)
		for _, acc := range accts {
			coins := acc.GetCoins()
			if coins.IsAnyNegative() {
				amt++
				msg += fmt.Sprintf("\t%s has a negative denomination of %s\n",
					acc.GetAddress().String(),
					coins.String())
			}
		}
		return sdk.PrintInvariant(types.ModuleName, "nonnegative-outstanding",
			fmt.Sprintf("amount of negative accounts found %d\n%s", amt, msg), false), false
	}
}
