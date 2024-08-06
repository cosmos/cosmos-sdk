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
	"cosmossdk.io/schema/appdata"
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

	res, err := postgres.StartIndexer(indexer.InitParams{
		Config: indexer.Config{
			Type:   "postgres",
			Config: cfgMap,
		},
		Context:      ctx,
		Logger:       log.NewTestLogger(t),
		AddressCodec: nil,
	})
	require.NoError(t, err)

	_, err = appdatasim.NewSimulator(appdatasim.Options{
		Listener: appdata.ListenerMux(
			appdata.DebugListener(os.Stdout),
			res.Listener,
		),
		AppSchema: indexertesting.ExampleAppSchema,
		StateSimOptions: statesim.Options{
			CanRetainDeletions: retainDeletions,
		},
	})
	require.NoError(t, err)
}
