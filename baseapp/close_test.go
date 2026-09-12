package baseapp_test

import (
	"errors"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// closeCountingDB mimics the backends used in production: pebble panics when
// Close is called on an already-closed database, and goleveldb returns an
// error. Either way, the second Close must never reach the backend.
type closeCountingDB struct {
	dbm.DB
	closes   int
	closeErr error
}

func (db *closeCountingDB) Close() error {
	db.closes++
	if db.closes > 1 {
		panic("pebble: closed")
	}

	if db.closeErr != nil {
		return db.closeErr
	}

	return db.DB.Close()
}

// TestBaseAppCloseIsIdempotent covers the shutdown path in server/start.go,
// where both startCmtNode's cleanup and startApp's cleanup are deferred and
// each calls app.Close().
func TestBaseAppCloseIsIdempotent(t *testing.T) {
	db := &closeCountingDB{DB: dbm.NewMemDB()}
	app := baseapp.NewBaseApp(t.Name(), log.NewTestLogger(t), db, nil)

	require.NoError(t, app.Close())
	require.NoError(t, app.Close())
	require.NoError(t, app.Close())
	require.Equal(t, 1, db.closes)
}

// TestBaseAppCloseReplaysError ensures a failure reported by the first Close is
// not lost by the later, no-op calls.
func TestBaseAppCloseReplaysError(t *testing.T) {
	closeErr := errors.New("boom")
	db := &closeCountingDB{DB: dbm.NewMemDB(), closeErr: closeErr}
	app := baseapp.NewBaseApp(t.Name(), log.NewTestLogger(t), db, nil)

	require.ErrorIs(t, app.Close(), closeErr)
	require.ErrorIs(t, app.Close(), closeErr)
	require.Equal(t, 1, db.closes)
}
