package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

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

type indexerImpl struct {
	ctx     context.Context
	db      *sql.DB
	tx      *sql.Tx
	opts    Options
	modules map[string]*moduleIndexer
}

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
	_, err = tx.Exec(baseSQL)
	if err != nil {
		return indexer.InitResult{}, err
	}

	moduleIndexers := map[string]*moduleIndexer{}
	var sqlLogger func(msg, sql string, params ...interface{})
	if logger := params.Logger; logger != nil {
		sqlLogger = func(msg, sql string, params ...interface{}) {
			if len(params) == 0 {
				logger.Debug(msg, "sql", sql)
			} else {
				logger.Debug(msg, "sql", sql, "params", params)
			}
		}
	}
	opts := Options{
		DisableRetainDeletions: config.DisableRetainDeletions,
		Logger:                 sqlLogger,
		AddressCodec:           params.AddressCodec,
	}

	idx := &indexerImpl{
		ctx:     ctx,
		db:      db,
		tx:      tx,
		opts:    opts,
		modules: moduleIndexers,
	}

	return indexer.InitResult{
		Listener: idx.Listener(),
		View:     idx,
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
