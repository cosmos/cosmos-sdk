package baseapp_test

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// closeOncePanicDB mimics a PebbleDB handle: the first Close() succeeds, any
// subsequent Close() panics ("pebble: closed"). It is used to prove BaseApp.Close
// releases the underlying DB exactly once even though server.startInProcess calls
// app.Close() twice on shutdown (startCmtNode's cleanupFn and startApp's
// appCleanupFn). A MemDB cannot reproduce this because its Close() is a no-op.
type closeOncePanicDB struct {
	dbm.DB
	closes int
}

func (d *closeOncePanicDB) Close() error {
	d.closes++
	if d.closes > 1 {
		panic("pebble: closed")
	}
	return d.DB.Close()
}

// TestBaseAppCloseIsIdempotent guards against the double-close regression:
// calling Close() more than once must not re-close (and panic on) the DB.
func TestBaseAppCloseIsIdempotent(t *testing.T) {
	db := &closeOncePanicDB{DB: dbm.NewMemDB()}
	app := baseapp.NewBaseApp(t.Name(), log.NewTestLogger(t), db, nil)

	require.NoError(t, app.Close())
	require.NotPanics(t, func() {
		require.NoError(t, app.Close())
	}, "second Close must not re-close the underlying DB")
	require.Equal(t, 1, db.closes, "underlying DB must be closed exactly once")
}
