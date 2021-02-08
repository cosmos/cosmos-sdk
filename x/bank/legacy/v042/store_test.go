package v042_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v040"
	v042bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v042"
)

func TestStoreMigration(t *testing.T) {
	bankKey := sdk.NewKVStoreKey("bank")
	ctx := testutil.DefaultContext(bankKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(bankKey)

	_, _, addr := testdata.KeyTestPubAddr()
	denom := []byte("foo")
	value := []byte("bar")

	oldKey := append(append(v040bank.BalancesPrefix, addr...), denom...)
	store.Set(oldKey, value)

	err := v042bank.MigrateStore(ctx, bankKey, nil)
	require.NoError(t, err)

	newKey := append(v042bank.CreateAccountBalancesPrefix(addr), denom...)
	// -7 because we replaced "balances" with 0x02,
	// +1 because we added length-prefix to address.
	require.Equal(t, len(oldKey)-7+1, len(newKey))
	require.Nil(t, store.Get(oldKey))
	require.Equal(t, value, store.Get(newKey))
}
