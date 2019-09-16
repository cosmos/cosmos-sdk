package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

func TestSanitize(t *testing.T) {
	addr1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	authAcc1 := NewBaseAccountWithAddress(addr1)
	authAcc1.SetCoins(sdk.Coins{
		sdk.NewInt64Coin("bcoin", 150),
		sdk.NewInt64Coin("acoin", 150),
	})
	authAcc1.SetAccountNumber(1)

	addr2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	authAcc2 := NewBaseAccountWithAddress(addr2)
	authAcc2.SetCoins(sdk.Coins{
		sdk.NewInt64Coin("acoin", 150),
		sdk.NewInt64Coin("bcoin", 150),
	})

	genAccs := exported.GenesisAccounts{&authAcc1, &authAcc2}

	require.True(t, genAccs[0].GetAccountNumber() > genAccs[1].GetAccountNumber())
	require.Equal(t, genAccs[0].GetCoins()[0].Denom, "bcoin")
	require.Equal(t, genAccs[0].GetCoins()[1].Denom, "acoin")
	require.Equal(t, genAccs[1].GetAddress(), addr2)
	genAccs = SanitizeGenesisAccounts(genAccs)

	require.False(t, genAccs[0].GetAccountNumber() > genAccs[1].GetAccountNumber())
	require.Equal(t, genAccs[1].GetAddress(), addr1)
	require.Equal(t, genAccs[1].GetCoins()[0].Denom, "acoin")
	require.Equal(t, genAccs[1].GetCoins()[1].Denom, "bcoin")
}

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
)

// require duplicate accounts fails validation
func TestValidateGenesisDuplicateAccounts(t *testing.T) {
	acc1 := NewBaseAccountWithAddress(sdk.AccAddress(addr1))
	acc1.Coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))

	genAccs := make(exported.GenesisAccounts, 2)
	genAccs[0] = &acc1
	genAccs[1] = &acc1

	require.Error(t, ValidateGenAccounts(genAccs))
}

func TestGenesisAccountIterator(t *testing.T) {
	acc1 := NewBaseAccountWithAddress(sdk.AccAddress(addr1))
	acc1.Coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))

	acc2 := NewBaseAccountWithAddress(sdk.AccAddress(addr2))
	acc2.Coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))

	genAccounts := exported.GenesisAccounts{&acc1, &acc2}

	authGenState := DefaultGenesisState()
	authGenState.Accounts = genAccounts

	appGenesis := make(map[string]json.RawMessage)
	authGenStateBz, err := ModuleCdc.MarshalJSON(authGenState)
	require.NoError(t, err)

	appGenesis[ModuleName] = authGenStateBz

	var addresses []sdk.AccAddress
	GenesisAccountIterator{}.IterateGenesisAccounts(
		ModuleCdc, appGenesis, func(acc exported.Account) (stop bool) {
			addresses = append(addresses, acc.GetAddress())
			return false
		},
	)

	require.Len(t, addresses, 2)
	require.Equal(t, addresses[0], acc1.GetAddress())
	require.Equal(t, addresses[1], acc2.GetAddress())
}
