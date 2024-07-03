package testing

import (
	"context"
	"database/sql"
	"os"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/hashicorp/consul/sdk/freeport"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"

	"cosmossdk.io/indexer/postgres"
	"cosmossdk.io/indexer/postgres/internal/testdata"
)

func TestInitSchema(t *testing.T) {
	db := createTestDB(t)
	mm := postgres.NewModuleManager("test", testdata.ExampleSchema, postgres.Options{
		Logger: log.NewTestLogger(t),
	})
	_, err := db.Exec(postgres.BaseSQL)
	require.NoError(t, err)
	require.NoError(t, mm.InitializeSchema(context.Background(), db))
}

func createTestDB(t *testing.T) *sql.DB {
	tempDir, err := os.MkdirTemp("", "postgres-indexer-test")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(tempDir))
	})

	dbPort := freeport.GetOne(t)
	pgConfig := embeddedpostgres.DefaultConfig().
		Port(uint32(dbPort)).
		DataPath(tempDir)

	dbUrl := pgConfig.GetConnectionURL()
	pg := embeddedpostgres.NewDatabase(pgConfig)
	require.NoError(t, pg.Start())
	t.Cleanup(func() {
		require.NoError(t, pg.Stop())
	})

	db, err := sql.Open("pgx", dbUrl)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}
