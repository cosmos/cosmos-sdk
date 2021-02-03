package v042_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v040"
	v042distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v042"
)

func TestStoreMigration(t *testing.T) {
	distributionKey := sdk.NewKVStoreKey("distribution")
	ctx := testutil.DefaultContext(distributionKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(distributionKey)

	_, _, addr := testdata.KeyTestPubAddr()
	denom := []byte("foo")
	value := []byte("bar")

	oldKey := append(append(v040distribution.BalancesPrefix, addr...), denom...)
	store.Set(oldKey, value)

	err := v042distribution.StoreMigration(store)
	require.NoError(t, err)

	newKey := append(v042distribution.CreateAccountBalancesPrefix(addr), denom...)
	// -7 because we replaced "balances" with 0x02,
	// +1 because we added length-prefix to address.
	require.Equal(t, len(oldKey)-7+1, len(newKey))
	require.Nil(t, store.Get(oldKey))
	require.Equal(t, value, store.Get(newKey))
}
