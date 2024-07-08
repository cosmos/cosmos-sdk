package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/hashicorp/consul/sdk/freeport"
	"gotest.tools/v3/golden"

	// this is where we get our pgx database driver from
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/indexer/postgres"
	"cosmossdk.io/indexer/postgres/internal/testdata"
	"cosmossdk.io/schema/appdata"
)

func TestInitSchema(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		testInitSchema(t, false, "init_schema.txt")
	})

	t.Run("retain deletions disabled", func(t *testing.T) {
		testInitSchema(t, true, "init_schema_no_retain_delete.txt")
	})
}

func testInitSchema(t *testing.T, disableRetainDeletions bool, goldenFileName string) {
	connectionUrl := createTestDB(t)

	buf := &strings.Builder{}
	var logger = func(msg string, sql string, params ...interface{}) {
		_, err := fmt.Fprintln(buf, msg)
		require.NoError(t, err)
		_, err = fmt.Fprintln(buf, sql)
		require.NoError(t, err)
		if len(params) != 0 {
			_, err = fmt.Fprintln(buf, "Params:", params)
			require.NoError(t, err)
		}
		_, err = fmt.Fprintln(buf)
		require.NoError(t, err)
	}
	listener, err := postgres.StartIndexer(context.Background(), logger, postgres.Config{
		DatabaseURL:            connectionUrl,
		DisableRetainDeletions: disableRetainDeletions,
	})
	require.NoError(t, err)

	require.NotNil(t, listener.InitializeModuleData)
	require.NoError(t, listener.InitializeModuleData(appdata.ModuleInitializationData{
		ModuleName: "test",
		Schema:     testdata.ExampleSchema,
	}))

	require.NotNil(t, listener.Commit)
	require.NoError(t, listener.Commit(appdata.CommitData{}))

	golden.Assert(t, buf.String(), goldenFileName)
}

func createTestDB(t *testing.T) (connectionUrl string) {
	tempDir, err := os.MkdirTemp("", "postgres-indexer-test")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(tempDir))
	})

	dbPort := freeport.GetOne(t)
	pgConfig := embeddedpostgres.DefaultConfig().
		Port(uint32(dbPort)).
		DataPath(tempDir)

	connectionUrl = pgConfig.GetConnectionURL()
	pg := embeddedpostgres.NewDatabase(pgConfig)
	require.NoError(t, pg.Start())
	t.Cleanup(func() {
		require.NoError(t, pg.Stop())
	})

	return
}
