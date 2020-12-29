package v040

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	v040vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// convertBaseAccount converts a 0.39 BaseAccount to a 0.40 BaseAccount.
func convertBaseAccount(old *v039auth.BaseAccount) *v040auth.BaseAccount {
	var any *codectypes.Any
	// If the old genesis had a pubkey, we pack it inside an Any. Or else, we
	// just leave it nil.
	if old.PubKey != nil {
		var err error
		any, err = codectypes.NewAnyWithValue(old.PubKey)
		if err != nil {
			panic(err)
		}
	}

	return &v040auth.BaseAccount{
		Address:       old.Address.String(),
		PubKey:        any,
		AccountNumber: old.AccountNumber,
		Sequence:      old.Sequence,
	}
}

// convertBaseVestingAccount converts a 0.39 BaseVestingAccount to a 0.40 BaseVestingAccount.
func convertBaseVestingAccount(old *v039auth.BaseVestingAccount) *v040vesting.BaseVestingAccount {
	baseAccount := convertBaseAccount(old.BaseAccount)

	return &v040vesting.BaseVestingAccount{
		BaseAccount:      baseAccount,
		OriginalVesting:  old.OriginalVesting,
		DelegatedFree:    old.DelegatedFree,
		DelegatedVesting: old.DelegatedVesting,
		EndTime:          old.EndTime,
	}
}

// Migrate accepts exported x/auth genesis state from v0.38/v0.39 and migrates
// it to v0.40 x/auth genesis state. The migration includes:
//
// - Removing coins from account encoding.
// - Re-encode in v0.40 GenesisState.
func Migrate(authGenState v039auth.GenesisState) *v040auth.GenesisState {
	// Convert v0.39 accounts to v0.40 ones.
	var v040Accounts = make([]v040auth.GenesisAccount, len(authGenState.Accounts))
	for i, v039Account := range authGenState.Accounts {
		switch v039Account := v039Account.(type) {
		case *v039auth.BaseAccount:
			{
				v040Accounts[i] = convertBaseAccount(v039Account)
			}
		case *v039auth.ModuleAccount:
			{
				v040Accounts[i] = &v040auth.ModuleAccount{
					BaseAccount: convertBaseAccount(v039Account.BaseAccount),
					Name:        v039Account.Name,
					Permissions: v039Account.Permissions,
				}
			}
		case *v039auth.BaseVestingAccount:
			{
				v040Accounts[i] = convertBaseVestingAccount(v039Account)
			}
		case *v039auth.ContinuousVestingAccount:
			{
				v040Accounts[i] = &v040vesting.ContinuousVestingAccount{
					BaseVestingAccount: convertBaseVestingAccount(v039Account.BaseVestingAccount),
					StartTime:          v039Account.StartTime,
				}
			}
		case *v039auth.DelayedVestingAccount:
			{
				v040Accounts[i] = &v040vesting.DelayedVestingAccount{
					BaseVestingAccount: convertBaseVestingAccount(v039Account.BaseVestingAccount),
				}
			}
		case *v039auth.PeriodicVestingAccount:
			{
				vestingPeriods := make([]v040vesting.Period, len(v039Account.VestingPeriods))
				for j, period := range v039Account.VestingPeriods {
					vestingPeriods[j] = v040vesting.Period{
						Length: period.Length,
						Amount: period.Amount,
					}
				}
				v040Accounts[i] = &v040vesting.PeriodicVestingAccount{
					BaseVestingAccount: convertBaseVestingAccount(v039Account.BaseVestingAccount),
					StartTime:          v039Account.StartTime,
					VestingPeriods:     vestingPeriods,
				}
			}
		default:
			panic(sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "got invalid type %T", v039Account))
		}

	}

	// Convert v0.40 accounts into Anys.
	anys := make([]*codectypes.Any, len(v040Accounts))
	for i, v040Account := range v040Accounts {
		any, err := codectypes.NewAnyWithValue(v040Account)
		if err != nil {
			panic(err)
		}

		anys[i] = any
	}

	return &v040auth.GenesisState{
		Params: v040auth.Params{
			MaxMemoCharacters:      authGenState.Params.MaxMemoCharacters,
			TxSigLimit:             authGenState.Params.TxSigLimit,
			TxSizeCostPerByte:      authGenState.Params.TxSizeCostPerByte,
			SigVerifyCostED25519:   authGenState.Params.SigVerifyCostED25519,
			SigVerifyCostSecp256k1: authGenState.Params.SigVerifyCostSecp256k1,
		},
		Accounts: anys,
	}
}
