package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cosmossdk.io/schema/appdata"
)

type Config struct {
	// DatabaseURL is the PostgreSQL connection URL to use to connect to the database.
	DatabaseURL string `json:"database_url"`

	// DatabaseDriver is the PostgreSQL database/sql driver to use. This defaults to "pgx".
	DatabaseDriver string `json:"database_driver"`

	// DisableRetainDeletions disables the retain deletions functionality even if it is set in an object type schema.
	DisableRetainDeletions bool `json:"disable_retain_deletions"`
}

type SqlLogger = func(msg, sql string, params ...interface{})

func StartIndexer(ctx context.Context, logger SqlLogger, config Config) (appdata.Listener, error) {
	if config.DatabaseURL == "" {
		return appdata.Listener{}, errors.New("missing database URL")
	}

	driver := config.DatabaseDriver
	if driver == "" {
		driver = "pgx"
	}

	db, err := sql.Open(driver, config.DatabaseURL)
	if err != nil {
		return appdata.Listener{}, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return appdata.Listener{}, err
	}

	// commit base schema
	_, err = tx.Exec(BaseSQL)
	if err != nil {
		return appdata.Listener{}, err
	}

	moduleIndexers := map[string]*ModuleIndexer{}
	opts := Options{
		DisableRetainDeletions: config.DisableRetainDeletions,
		Logger:                 logger,
	}

	return appdata.Listener{
		InitializeModuleData: func(data appdata.ModuleInitializationData) error {
			moduleName := data.ModuleName
			modSchema := data.Schema
			_, ok := moduleIndexers[moduleName]
			if ok {
				return fmt.Errorf("module %s already initialized", moduleName)
			}

			mm := NewModuleIndexer(moduleName, modSchema, opts)
			moduleIndexers[moduleName] = mm

			return mm.InitializeSchema(ctx, tx)
		},
		Commit: func(data appdata.CommitData) (completionCallback func() error, err error) {
			err = tx.Commit()
			if err != nil {
				return nil, err
			}

			tx, err = db.BeginTx(ctx, nil)
			return nil, err
		},
	}, nil
}
