package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
)

// require invalid vesting account fails validation
func TestValidateGenesisInvalidAccounts(t *testing.T) {
	acc1 := authtypes.NewBaseAccountWithAddress(sdk.AccAddress(addr1))
	acc1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))
	endTime := int64(1548775410)
	baseVestingAcc := NewBaseVestingAccount(acc1, acc1Balance, endTime)

	// invalid delegated vesting
	baseVestingAcc.DelegatedVesting = acc1Balance.Add(acc1Balance...)

	acc2 := authtypes.NewBaseAccountWithAddress(sdk.AccAddress(addr2))
	// acc2Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))

	genAccs := make([]authtypes.GenesisAccount, 2)
	genAccs[0] = baseVestingAcc
	genAccs[1] = acc2

	require.Error(t, authtypes.ValidateGenAccounts(genAccs))
	baseVestingAcc.DelegatedVesting = acc1Balance
	genAccs[0] = baseVestingAcc
	require.NoError(t, authtypes.ValidateGenAccounts(genAccs))
	// invalid start time
	genAccs[0] = NewContinuousVestingAccountRaw(baseVestingAcc, 1548888000)
	require.Error(t, authtypes.ValidateGenAccounts(genAccs))
	// invalid period: duration
	genAccs[0] = NewPeriodicVestingAccountRaw(baseVestingAcc, endTime-100000, []Period{
		{Length: 100000 + 20, Amount: acc1Balance},
	})
	require.Error(t, authtypes.ValidateGenAccounts(genAccs))
	// invalid period: amount
	genAccs[0] = NewPeriodicVestingAccountRaw(baseVestingAcc, endTime-100000, []Period{
		{Length: 100000, Amount: acc1Balance.Add(acc1Balance...)},
	})
	require.Error(t, authtypes.ValidateGenAccounts(genAccs))
	// Passing case
	genAccs[0] = NewPeriodicVestingAccountRaw(baseVestingAcc, endTime-100000, []Period{
		{Length: 100000, Amount: acc1Balance},
	})
	require.NoError(t, authtypes.ValidateGenAccounts(genAccs))
}
