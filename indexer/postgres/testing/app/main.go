package main

import (
	"context"
	"database/sql"
	"os"

	"cosmossdk.io/log"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/hashicorp/consul/sdk/freeport"
	_ "github.com/jackc/pgx/v5/stdlib"

	"cosmossdk.io/indexer/postgres"
	"cosmossdk.io/schema/indexing"
	indexertesting "cosmossdk.io/schema/testing"
	appdatatest "cosmossdk.io/schema/testing/appdata"
	"cosmossdk.io/schema/testing/statesim"
)

func start() error {
	dbUrl, found := os.LookupEnv("DATABASE_URL")
	if !found {
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

		dbUrl = pgConfig.GetConnectionURL()
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
	}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return err
	}

	indexer, err := postgres.NewIndexer(db, postgres.Options{
		Logger: log.NewLogger(os.Stdout),
	})
	if err != nil {
		return err
	}

	res, err := indexer.Initialize(context.Background(), indexing.InitializationData{})
	if err != nil {
		return err
	}

	fixture := appdatatest.NewSimulator(appdatatest.SimulatorOptions{
		Listener:        res.Listener,
		AppSchema:       indexertesting.ExampleAppSchema,
		StateSimOptions: statesim.Options{},
	})

	err = fixture.Initialize()
	if err != nil {
		return err
	}

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
