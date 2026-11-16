package simapp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

func TestCheckStateDecodes(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(true)

	_, err := app.CheckStateDecodes(ctx)
	require.NoError(t, err)

	// A fee allowance queue entry with an empty value, as older feegrant code
	// wrote it, must fail the check.
	exp := time.Unix(1_000_000, 0).UTC()
	addr := app.AccountKeeper.GetModuleAddress(feegrant.ModuleName)
	queueKey, err := collections.EncodeKeyWithPrefix(
		feegrant.FeeAllowanceQueueKeyPrefix,
		app.FeeGrantKeeper.FeeAllowanceQueue.KeyCodec(),
		collections.Join3(exp, addr, addr),
	)
	require.NoError(t, err)
	ctx.KVStore(app.GetKey(feegrant.StoreKey)).Set(queueKey, []byte{})

	_, err = app.CheckStateDecodes(ctx)
	require.ErrorIs(t, err, collections.ErrEncoding)
	require.ErrorContains(t, err, "feegrant/allowances_queue")
}

// TestCollectionsModulesCoverStores requires every KV store of the app to be
// either checked by CheckStateDecodes or listed in StateCheckExcludedStores,
// so a store added later cannot go unchecked without notice.
func TestCollectionsModulesCoverStores(t *testing.T) {
	app := Setup(t, false)
	excluded := StateCheckExcludedStores()

	checked := map[string]bool{}
	for _, m := range app.CollectionsModules() {
		require.NotNil(t, m.StoreKey, "module %s has no mounted store", m.Name)
		checked[m.StoreKey.Name()] = true
		require.NotContains(t, excluded, m.StoreKey.Name(), "store %s is both checked and excluded", m.StoreKey.Name())
	}
	for _, key := range app.GetStoreKeys() {
		name := key.Name()
		if !checked[name] {
			require.Contains(t, excluded, name, "store %s is neither checked by CheckStateDecodes nor listed in StateCheckExcludedStores", name)
		}
	}
	for name := range excluded {
		require.NotNil(t, app.GetKey(name), "excluded store %s is not mounted", name)
	}
}
