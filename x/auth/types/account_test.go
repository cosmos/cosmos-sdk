package types_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestBaseAddressPubKey(t *testing.T) {
	_, pub1, addr1 := types.KeyTestPubAddr()
	_, pub2, addr2 := types.KeyTestPubAddr()
	acc := types.NewBaseAccountWithAddress(addr1)

	// check the address (set) and pubkey (not set)
	require.EqualValues(t, addr1, acc.GetAddress())
	require.EqualValues(t, nil, acc.GetPubKey())

	// can't override address
	err := acc.SetAddress(addr2)
	require.NotNil(t, err)
	require.EqualValues(t, addr1, acc.GetAddress())

	// set the pubkey
	err = acc.SetPubKey(pub1)
	require.Nil(t, err)
	require.Equal(t, pub1, acc.GetPubKey())

	// can override pubkey
	err = acc.SetPubKey(pub2)
	require.Nil(t, err)
	require.Equal(t, pub2, acc.GetPubKey())

	//------------------------------------

	// can set address on empty account
	acc2 := types.BaseAccount{}
	err = acc2.SetAddress(addr2)
	require.Nil(t, err)
	require.EqualValues(t, addr2, acc2.GetAddress())
}

func TestBaseAccountSequence(t *testing.T) {
	_, _, addr := types.KeyTestPubAddr()
	acc := types.NewBaseAccountWithAddress(addr)
	seq := uint64(7)

	err := acc.SetSequence(seq)
	require.Nil(t, err)
	require.Equal(t, seq, acc.GetSequence())
}

func TestBaseAccountMarshal(t *testing.T) {
	_, pub, addr := types.KeyTestPubAddr()
	acc := types.NewBaseAccountWithAddress(addr)
	seq := uint64(7)

	// set everything on the account
	err := acc.SetPubKey(pub)
	require.Nil(t, err)
	err = acc.SetSequence(seq)
	require.Nil(t, err)

	bz, err := app.AccountKeeper.MarshalAccount(acc)
	require.Nil(t, err)

	acc2, err := app.AccountKeeper.UnmarshalAccount(bz)
	require.Nil(t, err)
	require.Equal(t, acc, acc2)

	// error on bad bytes
	_, err = app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	require.NotNil(t, err)
}

func TestGenesisAccountValidate(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	baseAcc := types.NewBaseAccount(addr, pubkey, 0, 0)

	tests := []struct {
		name   string
		acc    types.GenesisAccount
		expErr bool
	}{
		{
			"valid base account",
			baseAcc,
			false,
		},
		{
			"invalid base valid account",
			types.NewBaseAccount(addr, secp256k1.GenPrivKey().PubKey(), 0, 0),
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expErr, tt.acc.Validate() != nil)
		})
	}
}

func TestModuleAccountMarshalYAML(t *testing.T) {
	name := "test"
	moduleAcc := types.NewEmptyModuleAccount(name, types.Minter, types.Burner, types.Staking)
	bs, err := yaml.Marshal(moduleAcc)
	require.NoError(t, err)

	want := "|\n  address: cosmos1n7rdpqvgf37ktx30a2sv2kkszk3m7ncmg5drhe\n  public_key: \"\"\n  account_number: 0\n  sequence: 0\n  name: test\n  permissions:\n  - minter\n  - burner\n  - staking\n"
	require.Equal(t, want, string(bs))
}

func TestHasPermissions(t *testing.T) {
	name := "test"
	macc := types.NewEmptyModuleAccount(name, types.Staking, types.Minter, types.Burner)
	cases := []struct {
		permission string
		expectHas  bool
	}{
		{types.Staking, true},
		{types.Minter, true},
		{types.Burner, true},
		{"other", false},
	}

	for i, tc := range cases {
		hasPerm := macc.HasPermission(tc.permission)
		if tc.expectHas {
			require.True(t, hasPerm, "test case #%d", i)
		} else {
			require.False(t, hasPerm, "test case #%d", i)
		}
	}
}

func TestValidate(t *testing.T) {
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	baseAcc := types.NewBaseAccount(addr, nil, 0, 0)
	tests := []struct {
		name   string
		acc    types.GenesisAccount
		expErr error
	}{
		{
			"valid module account",
			types.NewEmptyModuleAccount("test"),
			nil,
		},
		{
			"invalid name and address pair",
			types.NewModuleAccount(baseAcc, "test"),
			fmt.Errorf("address %s cannot be derived from the module name 'test'", addr),
		},
		{
			"empty module account name",
			types.NewModuleAccount(baseAcc, "    "),
			errors.New("module account name cannot be blank"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.acc.Validate()
			require.Equal(t, tt.expErr, err)
		})
	}
}

func TestModuleAccountJSON(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	baseAcc := types.NewBaseAccount(addr, nil, 10, 50)
	acc := types.NewModuleAccount(baseAcc, "test", "burner")

	bz, err := json.Marshal(acc)
	require.NoError(t, err)

	bz1, err := acc.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz1), string(bz))

	var a types.ModuleAccount
	require.NoError(t, json.Unmarshal(bz, &a))
	require.Equal(t, acc.String(), a.String())
}

func TestGenesisAccountsContains(t *testing.T) {
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	acc := types.NewBaseAccount(addr, secp256k1.GenPrivKey().PubKey(), 0, 0)

	genAccounts := types.GenesisAccounts{}
	require.False(t, genAccounts.Contains(acc.GetAddress()))

	genAccounts = append(genAccounts, acc)
	require.True(t, genAccounts.Contains(acc.GetAddress()))
}
