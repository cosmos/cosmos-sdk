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

	require.NoError(t, app.CheckStateDecodes(ctx))

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

	err = app.CheckStateDecodes(ctx)
	require.ErrorIs(t, err, collections.ErrEncoding)
	require.ErrorContains(t, err, "feegrant/allowances_queue")
}
