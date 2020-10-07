package v039

import (
	"fmt"

	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v038"
)

// Migrate accepts exported genesis state from v0.38 and migrates it to v0.39
// genesis state.
func Migrate(oldAuthGenState v038auth.GenesisState) GenesisState {
	accounts := make(v038auth.GenesisAccounts, len(oldAuthGenState.Accounts))

	for i, acc := range oldAuthGenState.Accounts {
		switch t := acc.(type) {
		case *v038auth.BaseAccount:
			accounts[i] = NewBaseAccount(t.Address, t.Coins, t.PubKey, t.AccountNumber, t.Sequence)

		case *v038auth.BaseVestingAccount:
			accounts[i] = NewBaseVestingAccount(
				NewBaseAccount(t.Address, t.Coins, t.PubKey, t.AccountNumber, t.Sequence),
				t.OriginalVesting, t.DelegatedFree, t.DelegatedVesting, t.EndTime,
			)

		case *v038auth.ContinuousVestingAccount:
			accounts[i] = NewContinuousVestingAccountRaw(
				NewBaseVestingAccount(
					NewBaseAccount(t.Address, t.Coins, t.PubKey, t.AccountNumber, t.Sequence),
					t.OriginalVesting, t.DelegatedFree, t.DelegatedVesting, t.EndTime,
				),
				t.StartTime,
			)

		case *v038auth.DelayedVestingAccount:
			accounts[i] = NewDelayedVestingAccountRaw(
				NewBaseVestingAccount(
					NewBaseAccount(t.Address, t.Coins, t.PubKey, t.AccountNumber, t.Sequence),
					t.OriginalVesting, t.DelegatedFree, t.DelegatedVesting, t.EndTime,
				),
			)

		case *v038auth.ModuleAccount:
			accounts[i] = NewModuleAccount(
				NewBaseAccount(t.Address, t.Coins, t.PubKey, t.AccountNumber, t.Sequence),
				t.Name, t.Permissions...,
			)

		default:
			panic(fmt.Sprintf("unexpected account type: %T", acc))
		}
	}

	accounts = v038auth.SanitizeGenesisAccounts(accounts)

	if err := v038auth.ValidateGenAccounts(accounts); err != nil {
		panic(err)
	}

	return NewGenesisState(oldAuthGenState.Params, accounts)
}
