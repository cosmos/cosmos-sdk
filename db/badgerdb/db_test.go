package badgerdb

import (
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/dbtest"
)

func TestAll(t *testing.T) {
	dbtest.TestAll(t, func(t *testing.T) dbm.DB {
		dirname := t.TempDir()
		db, err := NewDB(dirname)
		require.NoError(t, err)
		return db
	})
}

// todo: version file
