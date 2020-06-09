package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestSanitize(t *testing.T) {
	addr1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	authAcc1 := types.NewBaseAccountWithAddress(addr1)
	err := authAcc1.SetAccountNumber(1)
	require.NoError(t, err)

	addr2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	authAcc2 := types.NewBaseAccountWithAddress(addr2)

	genAccs := types.GenesisAccounts{authAcc1, authAcc2}

	require.True(t, genAccs[0].GetAccountNumber() > genAccs[1].GetAccountNumber())
	require.Equal(t, genAccs[1].GetAddress(), addr2)
	genAccs = types.SanitizeGenesisAccounts(genAccs)

	require.False(t, genAccs[0].GetAccountNumber() > genAccs[1].GetAccountNumber())
	require.Equal(t, genAccs[1].GetAddress(), addr1)
}

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.ValAddress(pk1.Address())
	addr2 = sdk.ValAddress(pk2.Address())
)

// require duplicate accounts fails validation
func TestValidateGenesisDuplicateAccounts(t *testing.T) {
	acc1 := types.NewBaseAccountWithAddress(sdk.AccAddress(addr1))

	genAccs := make(types.GenesisAccounts, 2)
	genAccs[0] = acc1
	genAccs[1] = acc1

	require.Error(t, types.ValidateGenAccounts(genAccs))
}

func TestGenesisAccountIterator(t *testing.T) {
	acc1 := types.NewBaseAccountWithAddress(sdk.AccAddress(addr1))
	acc2 := types.NewBaseAccountWithAddress(sdk.AccAddress(addr2))

	genAccounts := types.GenesisAccounts{acc1, acc2}

	authGenState := types.DefaultGenesisState()
	authGenState.Accounts = genAccounts

	appGenesis := make(map[string]json.RawMessage)
	authGenStateBz, err := appCodec.MarshalJSON(authGenState)
	require.NoError(t, err)

	appGenesis[types.ModuleName] = authGenStateBz

	var addresses []sdk.AccAddress
	types.GenesisAccountIterator{}.IterateGenesisAccounts(
		appCodec, appGenesis, func(acc types.AccountI) (stop bool) {
			addresses = append(addresses, acc.GetAddress())
			return false
		},
	)

	require.Len(t, addresses, 2)
	require.Equal(t, addresses[0], acc1.GetAddress())
	require.Equal(t, addresses[1], acc2.GetAddress())
}
