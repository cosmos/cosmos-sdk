package simapp

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// TestSyncAccountNumber tests if accounts module account number is set correctly with the value get from auth.
// Also check if the store entry for auth GlobalAccountNumberKey is successfully deleted.
func TestSyncAccountNumber(t *testing.T) {
	app := Setup(t, true)
	ctx := app.NewUncachedContext(true, cmtproto.Header{})

	bytesKey := authtypes.GlobalAccountNumberKey
	store := app.AuthKeeper.KVStoreService.OpenKVStore(ctx)

	// initially there is no value set yet
	v, err := store.Get(bytesKey)
	require.NoError(t, err)
	require.Nil(t, v)

	// set value for legacy account number
	v, err = collections.Uint64Value.Encode(10)
	require.NoError(t, err)
	err = store.Set(bytesKey, v)
	require.NoError(t, err)

	// make sure value are updated
	v, err = store.Get(bytesKey)
	require.NoError(t, err)
	require.NotEmpty(t, v)
	num, err := collections.Uint64Value.Decode(v)
	require.NoError(t, err)
	require.Equal(t, uint64(10), num)

	err = authkeeper.MigrateAccountNumberUnsafe(ctx, &app.AuthKeeper)
	require.NoError(t, err)

	// make sure the DB entry for this key is deleted
	v, err = store.Get(bytesKey)
	require.NoError(t, err)
	require.Nil(t, v)

	// check if accounts's account number is updated
	currentNum, err := app.AccountsKeeper.AccountNumber.Peek(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(10), currentNum)
}
