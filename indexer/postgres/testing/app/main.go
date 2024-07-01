package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/hashicorp/consul/sdk/freeport"
	_ "github.com/jackc/pgx/v5/stdlib"

	"cosmossdk.io/indexer/postgres"
	"cosmossdk.io/schema/appdata"
	indexertesting "cosmossdk.io/schema/testing"
	appdatatest "cosmossdk.io/schema/testing/appdata"
	"cosmossdk.io/schema/testing/statesim"
)

func start() error {
	tempDir, err := os.MkdirTemp("", "postgres-indexer-test")
	if err != nil {
		return err
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			panic(err)
		}
	}(tempDir)

	dbPort := freeport.MustTake(1)[0]
	pgConfig := embeddedpostgres.DefaultConfig().
		Port(uint32(dbPort)).
		DataPath(tempDir)

	dbUrl := pgConfig.GetConnectionURL()
	pg := embeddedpostgres.NewDatabase(pgConfig)
	err = pg.Start()
	if err != nil {
		return err
	}
	defer func(pg *embeddedpostgres.EmbeddedPostgres) {
		err := pg.Stop()
		if err != nil {
			panic(err)
		}
	}(pg)

	indexer, err := postgres.NewIndexer(context.Background(), postgres.configOptions{
		DatabaseDriver: "pgx",
		DatabaseURL:    dbUrl,
	})

	fixture := appdatatest.NewSimulator(appdatatest.SimulatorOptions{
		Listener: appdata.ListenerMux(
			//appdata.DebugListener(os.Stdout),
			indexer.Listener(),
		),
		AppSchema:       indexertesting.ExampleAppSchema,
		StateSimOptions: statesim.Options{},
	})

	err = fixture.Initialize()
	if err != nil {
		return err
	}

	go func() {
		db, err := sql.Open("pgx", pgConfig.GetConnectionURL())
		if err != nil {
			panic(err)
		}
		http.Handle("/graphql", postgres.NewGraphQLHandler(db))
		err = http.ListenAndServe(":8080", nil)
		if err != nil {
			panic(err)
		}
	}()

	blockDataGen := fixture.BlockDataGenN(1000)
	i := 0
	for {
		blockData := blockDataGen.Example(i)
		err = fixture.ProcessBlockData(blockData)
		if err != nil {
			return err
		}
		i++
	}

	return nil
}

func main() {
	err := start()
	if err != nil {
		panic(err)
	}
}
