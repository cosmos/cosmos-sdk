package testing

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/log"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/hashicorp/consul/sdk/freeport"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/indexer/postgres"
	"cosmossdk.io/schema"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/indexing"
	indexertesting "cosmossdk.io/schema/testing"
	"cosmossdk.io/schema/testing/appdatasim"
	"cosmossdk.io/schema/testing/statesim"
)

func TestPostgresIndexer(t *testing.T) {
	t.Run("RetainDeletions", func(t *testing.T) {
		testPostgresIndexer(t, true)
	})
	t.Run("NoRetainDeletions", func(t *testing.T) {
		testPostgresIndexer(t, false)
	})
}

func testPostgresIndexer(t *testing.T, retainDeletions bool) {
	tempDir, err := os.MkdirTemp("", "postgres-indexer-test")
	require.NoError(t, err)

	dbPort := freeport.GetOne(t)
	pgConfig := embeddedpostgres.DefaultConfig().
		Port(uint32(dbPort)).
		DataPath(tempDir)

	dbUrl := pgConfig.GetConnectionURL()
	pg := embeddedpostgres.NewDatabase(pgConfig)
	require.NoError(t, pg.Start())

	ctx, cancel := context.WithCancel(context.Background())

	t.Cleanup(func() {
		cancel()
		require.NoError(t, pg.Stop())
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	})

	db, err := sql.Open("pgx", dbUrl)
	require.NoError(t, err)

	indexer, err := postgres.NewIndexer(db, postgres.Options{
		RetainDeletions: retainDeletions,
		Logger:          log.NewTestLogger(t),
	})
	require.NoError(t, err)

	res, err := indexer.Initialize(ctx, indexing.InitializationData{})
	require.NoError(t, err)

	fixture := appdatasim.NewSimulator(appdatasim.Options{
		Listener: appdata.ListenerMux(
			appdata.DebugListener(os.Stdout),
			res.Listener,
		),
		AppSchema: indexertesting.ExampleAppSchema,
		StateSimOptions: statesim.Options{
			CanRetainDeletions: retainDeletions,
		},
	})

	require.NoError(t, fixture.Initialize())

	blockDataGen := fixture.BlockDataGenN(1000)
	for i := 0; i < 1000; i++ {
		blockData := blockDataGen.Example(i)
		require.NoError(t, fixture.ProcessBlockData(blockData))

		require.NoError(t, fixture.AppState().ScanModules(func(moduleName string, mod *statesim.Module) error {
			modMgr, ok := indexer.Modules()[moduleName]
			require.True(t, ok)

			return mod.ScanObjectCollections(func(collection *statesim.ObjectCollection) error {
				tblMgr, ok := modMgr.Tables()[collection.ObjectType().Name]
				require.True(t, ok)

				expectedCount := collection.Len()
				actualCount, err := tblMgr.Count(context.Background(), db)
				require.NoError(t, err)
				require.Equalf(t, expectedCount, actualCount, "table %s %s count mismatch", moduleName, collection.ObjectType().Name)

				return collection.ScanState(func(update schema.ObjectUpdate) error {
					found, err := tblMgr.Equals(
						context.Background(),
						db, update.Key, update.Value)
					if err != nil {
						return err
					}

					if !found {
						return fmt.Errorf("object not found in table %s %s %v %v", moduleName, collection.ObjectType().Name, update.Key, update.Value)
					}

					return nil
				})
			})
		}))
	}
}
