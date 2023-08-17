package sqllite

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"cosmossdk.io/store/v2"
	_ "modernc.org/sqlite"
)

const (
	driverName = "sqlite"
)

var _ store.VersionedDatabase = (*Database)(nil)

type Database struct {
	storage *sql.DB
}

func New(dataDir string) (*Database, error) {
	db, err := sql.Open(driverName, filepath.Join(dataDir, "ss.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite DB: %w", err)
	}

	return &Database{
		storage: db,
	}, nil
}
