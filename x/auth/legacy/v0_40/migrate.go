package v040

import (
	"fmt"

	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_39"
)

func convertBaseAccount(old *v039auth.BaseAccount) BaseAccount {
	return BaseAccount{
		Address:       old.Address,
		PubKey:        old.PubKey,
		AccountNumber: old.AccountNumber,
		Sequence:      old.Sequence,
	}
}

// Migrate accepts exported x/auth genesis state from v0.38/v0.39 and migrates
// it to v0.40 x/auth genesis state. The migration includes:
//
// - Removing coins from account encoding.
func Migrate(authGenState v039auth.GenesisState) *GenesisState {
	// Convert v0.39 BaseAccounts and ModuleAccounts to v0.40 ones.
	var ba = make([]BaseAccount, 0)
	var ma = make([]ModuleAccount, 0)
	var cva = make([]ContinuousVestingAccount, 0)
	var dva = make([]ContinuousVestingAccount, 0)
	for _, account := range authGenState.Accounts {
		// set coins to nil and allow the JSON encoding to omit coins.
		if err := account.SetCoins(nil); err != nil {
			panic(fmt.Sprintf("failed to set account coins to nil: %s", err))
		}

		switch account.(type) {
		case *v039auth.BaseAccount:
			{
				ba = append(ba, convertBaseAccount(account.(*v039auth.BaseAccount)))
			}
		case *v039auth.ModuleAccount:
			{
				v039Account := account.(*v039auth.ModuleAccount)
				v040BaseAccount := convertBaseAccount(v039Account.BaseAccount)
				ma = append(ma, ModuleAccount{
					BaseAccount: &v040BaseAccount,
					Name:        v039Account.Name,
					Permissions: v039Account.Permissions,
				})
			}
		}
		// Other account types are handled by the vesting module.
	}

	// Convert v0.40 BaseAccounts and ModuleAccounts into Anys.
	// var baAny = make([]BaseAccount, 0)
	// var maAny = make([]ModuleAccount, 0)

	return &GenesisState{
		Params: Params{
			MaxMemoCharacters:      authGenState.Params.MaxMemoCharacters,
			TxSigLimit:             authGenState.Params.TxSigLimit,
			TxSizeCostPerByte:      authGenState.Params.TxSizeCostPerByte,
			SigVerifyCostED25519:   authGenState.Params.SigVerifyCostED25519,
			SigVerifyCostSecp256k1: authGenState.Params.SigVerifyCostSecp256k1,
		},
	}
}
