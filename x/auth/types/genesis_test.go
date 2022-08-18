package types_test

import (
	"encoding/json"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestSanitize(t *testing.T) {
	acc1Addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	acc2Addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	t.Run("accounts are sorted by account number", func(t *testing.T) {
		acc1 := types.NewBaseAccountWithAddress(acc1Addr)
		acc1.AccountNumber = 2
		acc2 := types.NewBaseAccountWithAddress(acc2Addr)
		acc2.AccountNumber = 0
		acc3 := types.NewEmptyModuleAccount("testing")
		acc3.BaseAccount.AccountNumber = 1

		input := types.GenesisAccounts{acc1, acc2, acc3}
		expected := types.GenesisAccounts{acc2, acc3, acc1}
		actual := types.SanitizeGenesisAccounts(input)

		require.Equal(t, expected, actual)
	})

	t.Run("duplicate account numbers are corrected", func(t *testing.T) {
		acc1 := types.NewBaseAccountWithAddress(acc1Addr)
		acc1.AccountNumber = 0
		acc2 := types.NewBaseAccountWithAddress(acc2Addr)
		acc2.AccountNumber = 1
		acc3 := types.NewEmptyModuleAccount("testing")
		acc3.BaseAccount.AccountNumber = 0

		expAcc3 := types.NewEmptyModuleAccount("testing")
		expAcc3.BaseAccount.AccountNumber = 2

		input := types.GenesisAccounts{acc1, acc2, acc3}
		expected := types.GenesisAccounts{acc1, acc2, expAcc3}
		actual := types.SanitizeGenesisAccounts(input)

		require.Equal(t, expected, actual)
	})
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
	accounts, err := types.PackAccounts(genAccounts)
	require.NoError(t, err)
	authGenState.Accounts = accounts

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

func TestPackAccountsAny(t *testing.T) {
	var accounts []*codectypes.Any

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"expected genesis account",
			func() {
				accounts = []*codectypes.Any{{}}
			},
			false,
		},
		{
			"success",
			func() {
				genAccounts := types.GenesisAccounts{&types.BaseAccount{}}
				accounts = make([]*codectypes.Any, len(genAccounts))

				for i, a := range genAccounts {
					msg, ok := a.(proto.Message)
					require.Equal(t, ok, true)
					any, err := codectypes.NewAnyWithValue(msg)
					require.NoError(t, err)
					accounts[i] = any
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			tc.malleate()

			res, err := types.UnpackAccounts(accounts)

			if tc.expPass {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, len(res), len(accounts))
			} else {
				require.Error(t, err)
				require.Nil(t, res)
			}
		})
	}
}
