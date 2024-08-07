package testing

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"cosmossdk.io/log"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/hashicorp/consul/sdk/freeport"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/indexer/postgres"
	"cosmossdk.io/schema/addressutil"
	"cosmossdk.io/schema/indexer"
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

	cfg := postgres.Config{
		DatabaseURL:            dbUrl,
		DisableRetainDeletions: !retainDeletions,
	}
	cfgBz, err := json.Marshal(cfg)
	require.NoError(t, err)

	var cfgMap map[string]interface{}
	err = json.Unmarshal(cfgBz, &cfgMap)

	pgIndexer, err := postgres.StartIndexer(indexer.InitParams{
		Config: indexer.Config{
			Type:   "postgres",
			Config: cfgMap,
		},
		Context:      ctx,
		Logger:       log.NewTestLogger(t),
		AddressCodec: addressutil.HexAddressCodec{},
	})
	require.NoError(t, err)

	sim, err := appdatasim.NewSimulator(appdatasim.Options{
		Listener:  pgIndexer.Listener(),
		AppSchema: indexertesting.ExampleAppSchema,
		StateSimOptions: statesim.Options{
			CanRetainDeletions: retainDeletions,
		},
	})
	require.NoError(t, err)

	blockDataGen := sim.BlockDataGenN(10, 100)
	for i := 0; i < 10; i++ {
		// using Example generates a deterministic data set based
		// on a seed so that regression tests can be created OR rapid.Check can
		// be used for fully random property-based testing
		blockData := blockDataGen.Example(i)

		// process the generated block data with the simulator which will also
		// send it to the indexer
		require.NoError(t, sim.ProcessBlockData(blockData))

		// compare the expected state in the simulator to the actual state in the indexer and expect the diff to be empty
		require.Empty(t, appdatasim.DiffAppData(sim, pgIndexer))
	}
}
