package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	authtypes "cosmossdk.io/x/auth/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
)

// require invalid vesting account fails validation
func TestValidateGenesisInvalidAccounts(t *testing.T) {
	ac := codectestutil.CodecOptions{}.GetAddressCodec()
	addr1Str, err := ac.BytesToString(addr1)
	require.NoError(t, err)
	acc1 := authtypes.NewBaseAccountWithAddress(addr1Str)
	acc1Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))
	baseVestingAcc, err := NewBaseVestingAccount(acc1, acc1Balance, 1548775410)
	require.NoError(t, err)

	// invalid delegated vesting
	baseVestingAcc.DelegatedVesting = acc1Balance.Add(acc1Balance...)

	addr2Str, err := ac.BytesToString(addr2)
	require.NoError(t, err)
	acc2 := authtypes.NewBaseAccountWithAddress(addr2Str)
	// acc2Balance := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150))

	genAccs := make([]authtypes.GenesisAccount, 2)
	genAccs[0] = baseVestingAcc
	genAccs[1] = acc2

	require.Error(t, authtypes.ValidateGenAccounts(genAccs, ac))
	baseVestingAcc.DelegatedVesting = acc1Balance
	genAccs[0] = baseVestingAcc
	require.NoError(t, authtypes.ValidateGenAccounts(genAccs, ac))
	// invalid start time
	genAccs[0] = NewContinuousVestingAccountRaw(baseVestingAcc, 1548888000)
	require.Error(t, authtypes.ValidateGenAccounts(genAccs, ac))
}
