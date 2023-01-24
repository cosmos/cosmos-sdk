package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

const balanceInvariant = "Funds-Holder-Balance"

// RegisterInvariants registers all quarantine invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper) {
	ir.RegisterRoute(quarantine.ModuleName, balanceInvariant, FundsHolderBalanceInvariant(keeper))
}

// FundsHolderBalanceInvariant checks that the funds-holder account has enough funds to cover all quarantined funds.
func FundsHolderBalanceInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg, broken := fundsHolderBalanceInvariantHelper(ctx, keeper)
		return sdk.FormatInvariant(quarantine.ModuleName, balanceInvariant, msg), broken
	}
}

func fundsHolderBalanceInvariantHelper(ctx sdk.Context, keeper Keeper) (string, bool) {
	totalQuarantined := sdk.Coins{}
	accumulator := func(_, _ sdk.AccAddress, record *quarantine.QuarantineRecord) bool {
		totalQuarantined = totalQuarantined.Add(record.Coins...)
		return false
	}
	keeper.IterateQuarantineRecords(ctx, nil, accumulator)

	if totalQuarantined.IsZero() {
		return "total funds quarantined is zero", false
	}

	holder := keeper.GetFundsHolder()
	fundsHolderBalance := keeper.bankKeeper.GetAllBalances(ctx, holder)

	problems := make([]string, 0, len(totalQuarantined))
	for _, needed := range totalQuarantined {
		have, holding := fundsHolderBalance.Find(needed.Denom)
		switch {
		case !have:
			problems = append(problems, fmt.Sprintf("zero is less than %s", needed))
		case holding.IsLT(needed):
			problems = append(problems, fmt.Sprintf("%s is less than %s", holding, needed))
		}
	}

	broken := len(problems) > 0
	msg := fmt.Sprintf("quarantine funds holder account %s balance ", holder)
	if broken {
		msg += "insufficient"
	} else {
		msg += "sufficient"
	}

	if fundsHolderBalance.IsZero() {
		msg += ", have: zero balance"
	} else {
		msg += fmt.Sprintf(", have: %s", fundsHolderBalance)
	}
	msg += fmt.Sprintf(", need: %s", totalQuarantined)

	if broken {
		msg += fmt.Sprintf(", %s: insufficient funds", strings.Join(problems, ", "))
	}

	return msg, broken
}
