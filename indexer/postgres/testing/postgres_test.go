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
		RetainDeletions: true,
	})
	require.NoError(t, err)

	fixture := appdatatest.NewSimulator(appdatatest.SimulatorOptions{
		Listener: appdata.ListenerMux(
			appdata.DebugListener(os.Stdout),
			indexer.Listener(),
		),
		AppSchema: indexertesting.ExampleAppSchema,
		StateSimOptions: statesim.Options{
			CanRetainDeletions: true,
		},
	})

	require.NoError(t, fixture.Initialize())

	blockDataGen := fixture.BlockDataGenN(1000)
	for i := 0; i < 1000; i++ {
		blockData := blockDataGen.Example(i)
		require.NoError(t, fixture.ProcessBlockData(blockData))
	}
}
