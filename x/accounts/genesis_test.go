package accounts

import (
	"testing"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
)

func TestGenesis(t *testing.T) {
	const testAccountType = "test"
	k, ctx := newKeeper(t, func(deps implementation.Dependencies) (string, implementation.Account, error) {
		acc, err := NewTestAccount(deps)
		return testAccountType, acc, err
	})
	// we init two accounts of the same type

	// we set counter to 10
	_, addr1, err := k.Init(ctx, testAccountType, []byte("sender"), &types.Empty{}, nil, nil)
	require.NoError(t, err)
	_, err = k.Execute(ctx, addr1, []byte("sender"), &types.UInt64Value{Value: 10}, nil)
	require.NoError(t, err)

	// we set counter to 20
	_, addr2, err := k.Init(ctx, testAccountType, []byte("sender"), &types.Empty{}, nil, nil)
	require.NoError(t, err)
	_, err = k.Execute(ctx, addr2, []byte("sender"), &types.UInt64Value{Value: 20}, nil)
	require.NoError(t, err)

	// export state
	state, err := k.ExportState(ctx)
	require.NoError(t, err)

	// reset state
	k, ctx = newKeeper(t, func(deps implementation.Dependencies) (string, implementation.Account, error) {
		acc, err := NewTestAccount(deps)
		return testAccountType, acc, err
	})
	// add to state a genesis account init msg.
	initMsg, err := implementation.PackAny(&types.Empty{})
	require.NoError(t, err)
	state.InitAccountMsgs = append(state.InitAccountMsgs, &v1.MsgInit{
		Sender:      "sender-2",
		AccountType: testAccountType,
		Message:     initMsg,
		Funds:       nil,
	})
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

	// check initted on genesis account
	addr3, err := k.makeAddress([]byte("sender-2"), 2, nil)
	require.NoError(t, err)
	resp, err = k.Query(ctx, addr3, &types.DoubleValue{})
	require.NoError(t, err)
	require.Equal(t, &types.UInt64Value{Value: 0}, resp)
	// reset state
	k, ctx = newKeeper(t, func(deps implementation.Dependencies) (string, implementation.Account, error) {
		acc, err := NewTestAccount(deps)
		return testAccountType, acc, err
	})

	// modify the accounts account number
	state.Accounts[0].AccountNumber = 99

	err = k.ImportState(ctx, state)
	require.NoError(t, err)

	currentAccNum, err := k.AccountNumber.Peek(ctx)
	require.NoError(t, err)
	// AccountNumber should be set to the highest account number in the genesis state + 2
	// (one is the sequence offset, the other is the genesis account being added through init msg)
	require.Equal(t, state.Accounts[0].AccountNumber+2, currentAccNum)

	// Test when init with empty accounts list, account number is not modified
	// make genesis state accounts empty
	state.Accounts = []*v1.GenesisAccount{}

	// set another value for account number
	err = k.AccountNumber.Set(ctx, uint64(10))
	require.NoError(t, err)

	err = k.ImportState(ctx, state)
	require.NoError(t, err)

	currentAccNum, err = k.AccountNumber.Peek(ctx)
	require.NoError(t, err)
	// AccountNumber should be 10 + 1
	// (one is the genesis account being added through init msg)
	require.Equal(t, uint64(11), currentAccNum)
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
