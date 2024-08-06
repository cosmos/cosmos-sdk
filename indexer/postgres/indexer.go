package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/indexer"
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

func StartIndexer(params indexer.InitParams) (indexer.InitResult, error) {
	config, err := decodeConfig(params.Config.Config)
	if err != nil {
		return indexer.InitResult{}, err
	}

	ctx := params.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if config.DatabaseURL == "" {
		return indexer.InitResult{}, errors.New("missing database URL")
	}

	driver := config.DatabaseDriver
	if driver == "" {
		driver = "pgx"
	}

	db, err := sql.Open(driver, config.DatabaseURL)
	if err != nil {
		return indexer.InitResult{}, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return indexer.InitResult{}, err
	}

	// commit base schema
	_, err = tx.Exec(BaseSQL)
	if err != nil {
		return indexer.InitResult{}, err
	}

	moduleIndexers := map[string]*ModuleIndexer{}
	var sqlLogger func(msg, sql string, params ...interface{})
	if logger := params.Logger; logger != nil {
		sqlLogger = func(msg, sql string, params ...interface{}) {
			params = append(params, "sql", sql)
			logger.Debug(msg, params...)
		}
	}
	opts := Options{
		DisableRetainDeletions: config.DisableRetainDeletions,
		Logger:                 sqlLogger,
		AddressCodec:           params.AddressCodec,
	}

	listener := appdata.Listener{
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
		OnObjectUpdate: func(data appdata.ObjectUpdateData) error {
			module := data.ModuleName
			mod, ok := moduleIndexers[module]
			if !ok {
				return fmt.Errorf("module %s not initialized", module)
			}

			for _, update := range data.Updates {
				tm, ok := mod.tables[update.TypeName]
				if !ok {
					return fmt.Errorf("object type %s not found in schema for module %s", update.TypeName, module)
				}

				var err error
				if update.Delete {
					err = tm.Delete(ctx, tx, update.Key)
				} else {
					err = tm.InsertUpdate(ctx, tx, update.Key, update.Value)
				}
				if err != nil {
					return err
				}
			}
			return nil
		},
		Commit: func(data appdata.CommitData) error {
			err = tx.Commit()
			if err != nil {
				return err
			}

			tx, err = db.BeginTx(ctx, nil)
			return err
		},
	}

	return indexer.InitResult{
		Listener: listener,
	}, nil
}

func decodeConfig(rawConfig map[string]interface{}) (*Config, error) {
	bz, err := json.Marshal(rawConfig)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(bz, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
