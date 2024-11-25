package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cosmossdk.io/schema/indexer"
	"cosmossdk.io/schema/logutil"
)

type Config struct {
	// DatabaseURL is the PostgreSQL connection URL to use to connect to the database.
	DatabaseURL string `json:"database_url"`

	// DatabaseDriver is the PostgreSQL database/sql driver to use. This defaults to "pgx".
	DatabaseDriver string `json:"database_driver"`

	// DisableRetainDeletions disables the retain deletions functionality even if it is set in an object type schema.
	DisableRetainDeletions bool `json:"disable_retain_deletions"`
}

type indexerImpl struct {
	ctx     context.Context
	db      *sql.DB
	tx      *sql.Tx
	opts    options
	modules map[string]*moduleIndexer
	logger  logutil.Logger
}

func init() {
	indexer.Register("postgres", indexer.Initializer{
		InitFunc:   startIndexer,
		ConfigType: Config{},
	})
}

func startIndexer(params indexer.InitParams) (indexer.InitResult, error) {
	config, ok := params.Config.Config.(Config)
	if !ok {
		return indexer.InitResult{}, fmt.Errorf("invalid config type, expected %T got %T", Config{}, params.Config.Config)
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
	opts := options{
		disableRetainDeletions: config.DisableRetainDeletions,
		logger:                 params.Logger,
		addressCodec:           params.AddressCodec,
	}

	idx := &indexerImpl{
		ctx:     ctx,
		db:      db,
		tx:      tx,
		opts:    opts,
		modules: moduleIndexers,
		logger:  params.Logger,
	}

	return indexer.InitResult{
		Listener: idx.listener(),
		View:     idx,
	}, nil
}
