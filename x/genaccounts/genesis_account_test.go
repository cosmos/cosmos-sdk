package genaccounts

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

func TestGenesisAccountValidate(t *testing.T) {
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	tests := []struct {
		name   string
		acc    GenesisAccount
		expErr error
	}{
		{
			"valid account",
			NewGenesisAccountRaw(addr, sdk.NewCoins(), sdk.NewCoins(), 0, 0, "", ""),
			nil,
		},
		{
			"valid module account",
			NewGenesisAccountRaw(addr, sdk.NewCoins(), sdk.NewCoins(), 0, 0, "mint", supply.Minter),
			nil,
		},
		{
			"invalid vesting amount",
			NewGenesisAccountRaw(addr, sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
				sdk.NewCoins(sdk.NewInt64Coin("stake", 100)), 0, 0, "", ""),
			errors.New("vesting amount cannot be greater than total amount"),
		},
		{
			"invalid vesting amount with multi coins",
			NewGenesisAccountRaw(addr,
				sdk.NewCoins(sdk.NewInt64Coin("uatom", 50), sdk.NewInt64Coin("eth", 50)),
				sdk.NewCoins(sdk.NewInt64Coin("uatom", 100), sdk.NewInt64Coin("eth", 20)),
				0, 0, "", ""),
			errors.New("vesting amount cannot be greater than total amount"),
		},
		{
			"invalid vesting times",
			NewGenesisAccountRaw(addr, sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
				sdk.NewCoins(sdk.NewInt64Coin("stake", 50)), 1654668078, 1554668078, "", ""),
			errors.New("vesting start-time cannot be before end-time"),
		},
		{
			"invalid module account name",
			NewGenesisAccountRaw(addr, sdk.NewCoins(), sdk.NewCoins(), 0, 0, " ", ""),
			errors.New("module account name cannot be blank"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.acc.Validate()
			require.Equal(t, tt.expErr, err)
		})
	}
}

func TestToAccount(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	// base account
	authAcc := auth.NewBaseAccountWithAddress(addr)
	authAcc.SetCoins(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150)))
	genAcc := NewGenesisAccount(&authAcc)
	acc := genAcc.ToAccount()
	require.IsType(t, &auth.BaseAccount{}, acc)
	require.Equal(t, &authAcc, acc.(*auth.BaseAccount))

	// vesting account
	vacc := auth.NewContinuousVestingAccount(
		&authAcc, time.Now().Unix(), time.Now().Add(24*time.Hour).Unix(),
	)
	genAcc, err := NewGenesisAccountI(vacc)
	require.NoError(t, err)
	acc = genAcc.ToAccount()
	require.IsType(t, &auth.ContinuousVestingAccount{}, acc)
	require.Equal(t, vacc, acc.(*auth.ContinuousVestingAccount))

	// module account
	macc := supply.NewEmptyModuleAccount("mint", supply.Minter)
	genAcc, err = NewGenesisAccountI(macc)
	require.NoError(t, err)
	acc = genAcc.ToAccount()
	require.IsType(t, &supply.ModuleAccount{}, acc)
	require.Equal(t, macc, acc.(*supply.ModuleAccount))
}
