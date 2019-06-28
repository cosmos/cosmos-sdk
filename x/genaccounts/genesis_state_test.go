package genaccounts

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestSanitize(t *testing.T) {

	addr1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	authAcc1 := auth.NewBaseAccountWithAddress(addr1)
	authAcc1.SetCoins(sdk.Coins{
		sdk.NewInt64Coin("bcoin", 150),
		sdk.NewInt64Coin("acoin", 150),
	})
	authAcc1.SetAccountNumber(1)
	genAcc1 := NewGenesisAccount(&authAcc1)

	addr2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	authAcc2 := auth.NewBaseAccountWithAddress(addr2)
	authAcc2.SetCoins(sdk.Coins{
		sdk.NewInt64Coin("acoin", 150),
		sdk.NewInt64Coin("bcoin", 150),
	})
	genAcc2 := NewGenesisAccount(&authAcc2)

	genesisState := GenesisState([]GenesisAccount{genAcc1, genAcc2})
	require.NoError(t, ValidateGenesis(genesisState))
	require.True(t, genesisState[0].AccountNumber > genesisState[1].AccountNumber)
	require.Equal(t, genesisState[0].Coins[0].Denom, "bcoin")
	require.Equal(t, genesisState[0].Coins[1].Denom, "acoin")
	require.Equal(t, genesisState[1].Address, addr2)
	genesisState.Sanitize()
	require.False(t, genesisState[0].AccountNumber > genesisState[1].AccountNumber)
	require.Equal(t, genesisState[1].Address, addr1)
	require.Equal(t, genesisState[1].Coins[0].Denom, "acoin")
	require.Equal(t, genesisState[1].Coins[1].Denom, "bcoin")
}

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
)

// require duplicate accounts fails validation
func TestValidateGenesisDuplicateAccounts(t *testing.T) {
	acc1 := auth.NewBaseAccountWithAddress(sdk.AccAddress(addr1))
	acc1.Coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))

	genAccs := make([]GenesisAccount, 2)
	genAccs[0] = NewGenesisAccount(&acc1)
	genAccs[1] = NewGenesisAccount(&acc1)

	genesisState := GenesisState(genAccs)
	err := ValidateGenesis(genesisState)
	require.Error(t, err)
}

// require invalid vesting account fails validation (invalid end time)
func TestValidateGenesisInvalidAccounts(t *testing.T) {
	acc1 := auth.NewBaseAccountWithAddress(sdk.AccAddress(addr1))
	acc1.Coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))
	acc2 := auth.NewBaseAccountWithAddress(sdk.AccAddress(addr2))
	acc2.Coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))

	genAccs := make([]GenesisAccount, 2)
	genAccs[0] = NewGenesisAccount(&acc1)
	genAccs[1] = NewGenesisAccount(&acc2)

	genesisState := GenesisState(genAccs)
	genesisState[0].OriginalVesting = genesisState[0].Coins
	err := ValidateGenesis(genesisState)
	require.Error(t, err)

	genesisState[0].StartTime = 1548888000
	genesisState[0].EndTime = 1548775410
	err = ValidateGenesis(genesisState)
	require.Error(t, err)
}
