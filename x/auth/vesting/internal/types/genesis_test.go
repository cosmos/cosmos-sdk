package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
)

// require invalid vesting account fails validation (invalid end time)
func TestValidateGenesisInvalidAccounts(t *testing.T) {
	acc1 := auth.NewBaseAccountWithAddress(sdk.AccAddress(addr1))
	acc1.Coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))
	baseVestingAcc := NewBaseVestingAccount(&acc1, acc1.Coins.Add(acc1.Coins), nil, nil, 1548775410)

	acc2 := auth.NewBaseAccountWithAddress(sdk.AccAddress(addr2))
	acc2.Coins = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))

	genAccs := make([]exported.GenesisAccount, 2)
	genAccs[0] = baseVestingAcc
	genAccs[1] = &acc2

	require.Error(t, auth.ValidateGenAccounts(genAccs))
	baseVestingAcc.OriginalVesting = acc1.Coins
	genAccs[0] = baseVestingAcc
	require.NoError(t, auth.ValidateGenAccounts(genAccs))

	genAccs[0] = NewContinuousVestingAccountRaw(baseVestingAcc, 1548888000)
	require.Error(t, auth.ValidateGenAccounts(genAccs))
}
