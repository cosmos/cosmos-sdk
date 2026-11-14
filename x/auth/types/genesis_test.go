package types_test

import (
	"encoding/json"
	"testing"

	proto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
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
	encodingConfig := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	cdc := encodingConfig.Codec

	acc1 := types.NewBaseAccountWithAddress(sdk.AccAddress(addr1))
	acc2 := types.NewBaseAccountWithAddress(sdk.AccAddress(addr2))

	genAccounts := types.GenesisAccounts{acc1, acc2}

	authGenState := types.DefaultGenesisState()
	accounts, err := types.PackAccounts(genAccounts)
	require.NoError(t, err)
	authGenState.Accounts = accounts

	appGenesis := make(map[string]json.RawMessage)
	authGenStateBz, err := cdc.MarshalJSON(authGenState)
	require.NoError(t, err)

	appGenesis[types.ModuleName] = authGenStateBz

	var addresses []sdk.AccAddress
	types.GenesisAccountIterator{}.IterateGenesisAccounts(
		cdc, appGenesis, func(acc sdk.AccountI) (stop bool) {
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

func TestValidateGenesis_MismatchedPubKey(t *testing.T) {
	original := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(original.Address())
	rotated := secp256k1.GenPrivKey().PubKey()
	acc := types.NewBaseAccount(addr, rotated, 1, 0)

	genState := func(history []types.GenesisPubKeyHistory, accs ...types.GenesisAccount) types.GenesisState {
		gs := types.NewGenesisState(types.DefaultParams(), accs)
		gs.PubKeyHistory = history
		return *gs
	}

	// No history entry: the pubkey must hash to the address.
	err := types.ValidateGenesis(genState(nil, acc))
	require.ErrorContains(t, err, "pubkey address does not match")

	// A malformed pubkey is an error, not a panic.
	malformed := types.NewBaseAccount(addr, &secp256k1.PubKey{Key: []byte{1, 2, 3}}, 1, 0)
	require.ErrorContains(t, types.ValidateGenesis(genState(nil, malformed)), "malformed pubkey")

	// A history entry for the address makes the account valid.
	oldPkAny, err := codectypes.NewAnyWithValue(original)
	require.NoError(t, err)
	history := []types.GenesisPubKeyHistory{{
		Address: addr.String(),
		Entries: []types.PubKeyHistoryEntry{{
			PubKey:           oldPkAny,
			ReplacedAtHeight: 10,
			NewKeyAddress:    sdk.AccAddress(rotated.Address()).String(),
		}},
	}}
	require.NoError(t, types.ValidateGenesis(genState(history, acc)))

	// History for some other address does not help.
	other := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	otherHistory := []types.GenesisPubKeyHistory{{Address: other.String(), Entries: history[0].Entries}}
	require.Error(t, types.ValidateGenesis(genState(otherHistory, acc)))

	// A ModuleCredential account is valid either way.
	cred, err := types.NewModuleCredential("group", []byte{1})
	require.NoError(t, err)
	credAcc, err := types.NewBaseAccountWithPubKey(cred)
	require.NoError(t, err)
	require.NoError(t, types.ValidateGenesis(genState(nil, credAcc)))
	credMismatch := types.NewBaseAccount(addr, cred, 2, 0)
	require.NoError(t, types.ValidateGenesis(genState(nil, credMismatch)))
	// A ModuleCredential account cannot have rotated, so history for it is invalid.
	require.ErrorContains(t, types.ValidateGenesis(genState(history, credMismatch)), "ModuleCredential")
}

func TestValidateGenesis_PubKeyHistory(t *testing.T) {
	k0 := secp256k1.GenPrivKey().PubKey()
	k1 := secp256k1.GenPrivKey().PubKey()
	k2 := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(k0.Address())
	natural := func(pk cryptotypes.PubKey) string { return sdk.AccAddress(pk.Address()).String() }
	entry := func(old, next cryptotypes.PubKey, height int64) types.PubKeyHistoryEntry {
		pkAny, err := codectypes.NewAnyWithValue(old)
		require.NoError(t, err)
		return types.PubKeyHistoryEntry{PubKey: pkAny, ReplacedAtHeight: height, NewKeyAddress: natural(next)}
	}
	hist := func(entries ...types.PubKeyHistoryEntry) []types.GenesisPubKeyHistory {
		return []types.GenesisPubKeyHistory{{Address: addr.String(), Entries: entries}}
	}
	// acc is the account after rotating k0 -> k1 -> k2.
	acc := types.NewBaseAccount(addr, k2, 1, 0)
	cred, err := types.NewModuleCredential("group", []byte{1})
	require.NoError(t, err)

	testCases := []struct {
		name    string
		accs    []types.GenesisAccount
		history []types.GenesisPubKeyHistory
		expErr  string
	}{
		{"valid", []types.GenesisAccount{acc}, hist(entry(k0, k1, 5), entry(k1, k2, 6)), ""},
		{"bad address", []types.GenesisAccount{acc}, []types.GenesisPubKeyHistory{{Address: "nope", Entries: []types.PubKeyHistoryEntry{entry(k0, k1, 5)}}}, "invalid pubkey history address"},
		{"duplicate address", []types.GenesisAccount{acc}, append(hist(entry(k0, k1, 5), entry(k1, k2, 6)), hist(entry(k1, k2, 6))...), "duplicate pubkey history"},
		{"no entries", []types.GenesisAccount{acc}, hist(), "no entries"},
		// History is keyed by rotation index, so two rotations in one block, or
		// heights that restart after a zero-height export, are valid.
		{"same height twice", []types.GenesisAccount{acc}, hist(entry(k0, k1, 5), entry(k1, k2, 5)), ""},
		{"heights not increasing", []types.GenesisAccount{acc}, hist(entry(k0, k1, 5), entry(k1, k2, 2)), ""},
		{"negative height", []types.GenesisAccount{acc}, hist(entry(k0, k1, -1), entry(k1, k2, 6)), "negative"},
		{"rotated back", []types.GenesisAccount{types.NewBaseAccount(addr, k0, 1, 0)}, hist(entry(k0, k1, 5), entry(k1, k0, 6)), ""},
		{"orphan history", nil, hist(entry(k0, k1, 5), entry(k1, k2, 6)), "no account"},
		{"orphan history for another address", []types.GenesisAccount{types.NewBaseAccount(sdk.AccAddress(k1.Address()), k1, 2, 0)}, hist(entry(k0, k1, 5), entry(k1, k2, 6)), "no account"},
		{"nil entry pubkey", []types.GenesisAccount{acc}, hist(entry(k0, k1, 5), types.PubKeyHistoryEntry{ReplacedAtHeight: 6, NewKeyAddress: natural(k2)}), "no pubkey"},
		{"bad new key address", []types.GenesisAccount{acc}, hist(entry(k0, k1, 5), types.PubKeyHistoryEntry{PubKey: entry(k1, k2, 6).PubKey, NewKeyAddress: "nope"}), "invalid new key address"},
		{"chain broken mid-history", []types.GenesisAccount{acc}, hist(entry(k0, k2, 5), entry(k1, k2, 6)), "does not match the next entry"},
		{"last entry does not match current pubkey", []types.GenesisAccount{acc}, hist(entry(k0, k1, 5), entry(k1, k0, 6)), "does not match the account's current pubkey"},
		// The first replaced key is the account's original key, so it must
		// hash to the account address.
		{"first entry is not the original key", []types.GenesisAccount{acc}, hist(entry(secp256k1.GenPrivKey().PubKey(), k1, 5), entry(k1, k2, 6)), "entry 0 pubkey address"},
		{"account has no pubkey", []types.GenesisAccount{types.NewBaseAccountWithAddress(addr)}, hist(entry(k0, k1, 5)), "has no pubkey"},
		{"ModuleCredential account", []types.GenesisAccount{types.NewBaseAccount(addr, cred, 1, 0)}, hist(entry(k0, cred, 5)), "ModuleCredential"},
		// A ModuleCredential cannot be rekeyed or rekeyed to, so it cannot be
		// a replaced key.
		{"ModuleCredential entry", []types.GenesisAccount{types.NewBaseAccount(addr, k1, 1, 0)}, hist(entry(cred, k1, 5)), "entry 0 is a ModuleCredential"},
		// Address() panics on these keys; validation must return an error.
		{"malformed entry pubkey", []types.GenesisAccount{acc}, hist(entry(&secp256k1.PubKey{Key: []byte{1, 2, 3}}, k1, 5), entry(k1, k2, 6)), "malformed pubkey"},
		{"malformed next entry pubkey", []types.GenesisAccount{acc}, hist(entry(k0, k1, 5), entry(&secp256k1.PubKey{Key: []byte{1, 2, 3}}, k2, 6)), "malformed pubkey"},
		{"malformed current pubkey", []types.GenesisAccount{types.NewBaseAccount(addr, &secp256k1.PubKey{Key: []byte{1, 2, 3}}, 1, 0)}, hist(entry(k0, k1, 5)), "malformed pubkey"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gs := types.NewGenesisState(types.DefaultParams(), tc.accs)
			gs.PubKeyHistory = tc.history
			err := types.ValidateGenesis(*gs)
			if tc.expErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expErr)
			}
		})
	}
}
