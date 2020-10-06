package v038

import (
	v036auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v036"
	v036genaccounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v036"
)

// Migrate accepts exported genesis state from v0.34 and migrates it to v0.38
// genesis state.
func Migrate(authGenState v036auth.GenesisState, genAccountsGenState v036genaccounts.GenesisState) GenesisState {
	accounts := make(GenesisAccounts, len(genAccountsGenState))

	for i, acc := range genAccountsGenState {
		var genAccount GenesisAccount

		baseAccount := NewBaseAccount(acc.Address, acc.Coins.Sort(), nil, acc.AccountNumber, acc.Sequence)

		switch {
		case !acc.OriginalVesting.IsZero():
			baseVestingAccount := NewBaseVestingAccount(
				baseAccount, acc.OriginalVesting.Sort(), acc.DelegatedFree.Sort(),
				acc.DelegatedVesting.Sort(), acc.EndTime,
			)

			if acc.StartTime != 0 && acc.EndTime != 0 {
				// continuous vesting account type
				genAccount = NewContinuousVestingAccountRaw(baseVestingAccount, acc.StartTime)
			} else if acc.EndTime != 0 {
				// delayed vesting account type
				genAccount = NewDelayedVestingAccountRaw(baseVestingAccount)
			}

		case acc.ModuleName != "":
			// module account type
			genAccount = NewModuleAccount(baseAccount, acc.ModuleName, acc.ModulePermissions...)

		default:
			// standard account type
			genAccount = baseAccount
		}

		accounts[i] = genAccount
	}

	accounts = SanitizeGenesisAccounts(accounts)

	if err := ValidateGenAccounts(accounts); err != nil {
		panic(err)
	}

	return NewGenesisState(authGenState.Params, accounts)
}
