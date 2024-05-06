package accounts

import (
	"testing"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
)

func TestGenesis(t *testing.T) {
	k, ctx := newKeeper(t, func(deps implementation.Dependencies) (string, implementation.Account, error) {
		acc, err := NewTestAccount(deps)
		return "test", acc, err
	})
	// we init two accounts of the same type

	// we set counter to 10
	_, addr1, err := k.Init(ctx, "test", []byte("sender"), &types.Empty{}, nil)
	require.NoError(t, err)
	_, err = k.Execute(ctx, addr1, []byte("sender"), &types.UInt64Value{Value: 10}, nil)
	require.NoError(t, err)

	// we set counter to 20
	_, addr2, err := k.Init(ctx, "test", []byte("sender"), &types.Empty{}, nil)
	require.NoError(t, err)
	_, err = k.Execute(ctx, addr2, []byte("sender"), &types.UInt64Value{Value: 20}, nil)
	require.NoError(t, err)

	// export state
	state, err := k.ExportState(ctx)
	require.NoError(t, err)

	// reset state
	_, ctx = colltest.MockStore()
	err = k.ImportState(ctx, state)
	require.NoError(t, err)

	// if genesis import went fine, we should be able to query the accounts
	// and get the expected values.
	resp, err := k.Query(ctx, addr1, &types.DoubleValue{})
	require.NoError(t, err)
	require.Equal(t, &types.UInt64Value{Value: 10}, resp)

	resp, err = k.Query(ctx, addr2, &types.DoubleValue{})
	require.NoError(t, err)
	require.Equal(t, &types.UInt64Value{Value: 20}, resp)
}

func TestImportAccountError(t *testing.T) {
	// Initialize the keeper and context for testing
	k, ctx := newKeeper(t, func(deps implementation.Dependencies) (string, implementation.Account, error) {
		acc, err := NewTestAccount(deps)
		return "test", acc, err
	})

	// Define a mock GenesisAccount with a non-existent account type
	acc := &v1.GenesisAccount{
		Address:       "test-address",
		AccountType:   "non-existent-type",
		AccountNumber: 1,
		State:         nil,
	}

	// Attempt to import the mock GenesisAccount into the state
	err := k.importAccount(ctx, acc)

	// Assert that an error is returned
	require.Error(t, err)

	// Assert that the error message contains the expected substring
	require.Contains(t, err.Error(), "account type non-existent-type not found in the registered accounts")
}
