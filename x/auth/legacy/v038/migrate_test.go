package v038

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v034auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v034"
	v036auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v036"
	v036genaccounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v036"

	"github.com/stretchr/testify/require"
)

func accAddressFromBech32(t *testing.T, addrStr string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(addrStr)
	require.NoError(t, err)
	return addr
}

func TestMigrate(t *testing.T) {
	var genesisState GenesisState

	params := v034auth.Params{
		MaxMemoCharacters:      10,
		TxSigLimit:             10,
		TxSizeCostPerByte:      10,
		SigVerifyCostED25519:   10,
		SigVerifyCostSecp256k1: 10,
	}

	acc1 := v036genaccounts.GenesisAccount{
		Address:       accAddressFromBech32(t, "cosmos1f9xjhxm0plzrh9cskf4qee4pc2xwp0n0556gh0"),
		Coins:         sdk.NewCoins(sdk.NewInt64Coin("stake", 400000)),
		Sequence:      1,
		AccountNumber: 1,
	}
	acc2 := v036genaccounts.GenesisAccount{
		Address:           accAddressFromBech32(t, "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"),
		Coins:             sdk.NewCoins(sdk.NewInt64Coin("stake", 400000000)),
		Sequence:          4,
		AccountNumber:     2,
		ModuleName:        "bonded_tokens_pool",
		ModulePermissions: []string{"burner", "staking"},
	}
	acc3 := v036genaccounts.GenesisAccount{
		Address:         accAddressFromBech32(t, "cosmos17n9sztlhx32tfy0tg0zc2ttmkeeth50yyuv9he"),
		Coins:           sdk.NewCoins(sdk.NewInt64Coin("stake", 10000205)),
		OriginalVesting: sdk.NewCoins(sdk.NewInt64Coin("stake", 10000205)),
		StartTime:       time.Now().Unix(),
		EndTime:         time.Now().Add(48 * time.Hour).Unix(),
		Sequence:        5,
		AccountNumber:   3,
	}
	acc4 := v036genaccounts.GenesisAccount{
		Address:         accAddressFromBech32(t, "cosmos1fmk5elg4r62mlexd36tqjcwyafs7mek0js5m4d"),
		Coins:           sdk.NewCoins(sdk.NewInt64Coin("stake", 10000205)),
		OriginalVesting: sdk.NewCoins(sdk.NewInt64Coin("stake", 10000205)),
		EndTime:         time.Now().Add(48 * time.Hour).Unix(),
		Sequence:        15,
		AccountNumber:   4,
	}

	require.NotPanics(t, func() {
		genesisState = Migrate(
			v036auth.GenesisState{
				Params: params,
			},
			v036genaccounts.GenesisState{acc1, acc2, acc3, acc4},
		)
	})

	expectedAcc1 := NewBaseAccount(acc1.Address, acc1.Coins, nil, acc1.AccountNumber, acc1.Sequence)
	expectedAcc2 := NewModuleAccount(
		NewBaseAccount(acc2.Address, acc2.Coins, nil, acc2.AccountNumber, acc2.Sequence),
		acc2.ModuleName, acc2.ModulePermissions...,
	)
	expectedAcc3 := NewContinuousVestingAccountRaw(
		NewBaseVestingAccount(
			NewBaseAccount(acc3.Address, acc3.Coins, nil, acc3.AccountNumber, acc3.Sequence),
			acc3.OriginalVesting, acc3.DelegatedFree, acc3.DelegatedVesting, acc3.EndTime,
		),
		acc3.StartTime,
	)
	expectedAcc4 := NewDelayedVestingAccountRaw(
		NewBaseVestingAccount(
			NewBaseAccount(acc4.Address, acc4.Coins, nil, acc4.AccountNumber, acc4.Sequence),
			acc4.OriginalVesting, acc4.DelegatedFree, acc4.DelegatedVesting, acc4.EndTime,
		),
	)

	require.Equal(
		t, genesisState, GenesisState{
			Params:   params,
			Accounts: GenesisAccounts{expectedAcc1, expectedAcc2, expectedAcc3, expectedAcc4},
		},
	)
}

func TestMigrateInvalid(t *testing.T) {
	testCases := []struct {
		name string
		acc  v036genaccounts.GenesisAccount
	}{
		{
			"module account with invalid name",
			v036genaccounts.GenesisAccount{
				Address:           accAddressFromBech32(t, "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"),
				Coins:             sdk.NewCoins(sdk.NewInt64Coin("stake", 400000000)),
				Sequence:          4,
				AccountNumber:     2,
				ModuleName:        "    ",
				ModulePermissions: []string{"burner", "staking"},
			},
		},
		{
			"module account with invalid permissions",
			v036genaccounts.GenesisAccount{
				Address:           accAddressFromBech32(t, "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"),
				Coins:             sdk.NewCoins(sdk.NewInt64Coin("stake", 400000000)),
				Sequence:          4,
				AccountNumber:     2,
				ModuleName:        "bonded_tokens_pool",
				ModulePermissions: []string{""},
			},
		},
		{
			"module account with invalid address",
			v036genaccounts.GenesisAccount{
				Address:           accAddressFromBech32(t, "cosmos17n9sztlhx32tfy0tg0zc2ttmkeeth50yyuv9he"),
				Coins:             sdk.NewCoins(sdk.NewInt64Coin("stake", 400000000)),
				Sequence:          4,
				AccountNumber:     2,
				ModuleName:        "bonded_tokens_pool",
				ModulePermissions: []string{"burner", "staking"},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.Panics(t, func() {
				Migrate(
					v036auth.GenesisState{
						Params: v034auth.Params{
							MaxMemoCharacters:      10,
							TxSigLimit:             10,
							TxSizeCostPerByte:      10,
							SigVerifyCostED25519:   10,
							SigVerifyCostSecp256k1: 10,
						},
					},
					v036genaccounts.GenesisState{tc.acc},
				)
			})
		})
	}
}
