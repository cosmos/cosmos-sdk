package testing

import (
	"context"
	"os"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/hashicorp/consul/sdk/freeport"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/indexer/postgres"
	"cosmossdk.io/schema"
	"cosmossdk.io/schema/appdata"
	indexertesting "cosmossdk.io/schema/testing"
	appdatatest "cosmossdk.io/schema/testing/appdata"
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

	indexer, err := postgres.NewIndexer(ctx, postgres.Options{
		Driver:          "pgx",
		ConnectionURL:   dbUrl,
		RetainDeletions: retainDeletions,
	})
	require.NoError(t, err)

	fixture := appdatatest.NewSimulator(appdatatest.SimulatorOptions{
		Listener: appdata.ListenerMux(
			appdata.DebugListener(os.Stdout),
			indexer.Listener(),
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
		require.NoError(t, fixture.AppState().ScanModuleSchemas(func(modName string, modSchema schema.ModuleSchema) error {
			modState, ok := fixture.AppState().GetModule(modName)
			require.True(t, ok)
			modMgr, ok := indexer.Modules[modName]
			require.True(t, ok)
			for _, objType := range modSchema.ObjectTypes {
				objColl, ok := modState.GetObjectCollection(objType.Name)
				require.True(t, ok)
				tblMgr, ok := modMgr.Tables[objType.Name]
				require.True(t, ok)

				expectedCount := objColl.Len()
				actualCount, err := tblMgr.Count(context.Background(), indexer.Tx)
				require.NoError(t, err)
				require.Equalf(t, expectedCount, actualCount, "table %s %s count mismatch", modName, objType.Name)

				objColl.ScanState(func(update schema.ObjectUpdate) bool {
					found, err := tblMgr.Equals(
						context.Background(),
						indexer.Tx, update.Key, update.Value)
					require.NoError(t, err)
					require.Truef(t, found, "object not found in table %s %s %v %v", modName, objType.Name, update.Key, update.Value)
					return true
				})
			}
			return nil
		}))
	}
}
