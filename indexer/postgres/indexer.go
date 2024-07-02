package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/indexing"
	"cosmossdk.io/schema/logutil"
)

type Indexer struct {
	ctx     context.Context
	db      *sql.DB
	tx      *sql.Tx
	options Options

	modules map[string]*ModuleManager
}

func (i *Indexer) Initialize(ctx context.Context, data indexing.InitializationData) (indexing.InitializationResult, error) {
	i.options.Logger.Info("Starting Postgres Indexer")

	go func() {
		<-ctx.Done()
		err := i.db.Close()
		if err != nil {
			panic(fmt.Sprintf("failed to close database: %v", err))
		}
	}()

	i.ctx = ctx

	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return indexing.InitializationResult{}, fmt.Errorf("failed to start transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, BaseSQL)
	if err != nil {
		return indexing.InitializationResult{}, err
	}

	i.tx = tx

	return indexing.InitializationResult{
		Listener: i.listener(),
	}, nil
}

type configOptions struct {
	DatabaseDriver  string `json:"database_driver"`
	DatabaseURL     string `json:"database_url"`
	RetainDeletions bool   `json:"retain_deletions"`
}

func init() {
	indexing.RegisterIndexer("postgres", func(rawOpts map[string]interface{}, resources indexing.IndexerResources) (indexing.Indexer, error) {
		bz, err := json.Marshal(rawOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal options: %w", err)
		}

		var opts configOptions
		err = json.Unmarshal(bz, &opts)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal options: %w", err)
		}

		if opts.DatabaseDriver == "" {
			opts.DatabaseDriver = "pgx"
		}

		if opts.DatabaseURL == "" {
			return nil, fmt.Errorf("connection URL not set")
		}

		db, err := sql.Open(opts.DatabaseDriver, opts.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}

		return NewIndexer(db, Options{
			RetainDeletions: opts.RetainDeletions,
			Logger:          resources.Logger,
		})
	})
}

type Options struct {
	RetainDeletions bool
	Logger          logutil.Logger
}

func NewIndexer(db *sql.DB, opts Options) (*Indexer, error) {
	return &Indexer{
		db:      db,
		modules: map[string]*ModuleManager{},
		options: opts,
	}, nil
}

func (i *Indexer) listener() appdata.Listener {
	return appdata.Listener{
		InitializeModuleData: i.initModuleSchema,
		OnObjectUpdate:       i.onObjectUpdate,
		Commit:               i.commit,
	}
}

func (i *Indexer) initModuleSchema(data appdata.ModuleInitializationData) error {
	moduleName := data.ModuleName
	modSchema := data.Schema
	_, ok := i.modules[moduleName]
	if ok {
		return fmt.Errorf("module %s already initialized", moduleName)
	}

	mm := newModuleManager(moduleName, modSchema, i.options)
	i.modules[moduleName] = mm

	return mm.InitializeSchema(i.ctx, i.tx)
}

func (i *Indexer) onObjectUpdate(data appdata.ObjectUpdateData) error {
	module := data.ModuleName
	mod, ok := i.modules[module]
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
			err = tm.Delete(i.ctx, i.tx, update.Key)
		} else {
			err = tm.InsertUpdate(i.ctx, i.tx, update.Key, update.Value)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Indexer) commit(_ appdata.CommitData) error {
	err := i.tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	i.tx, err = i.db.BeginTx(i.ctx, nil)
	return err
}

func (i *Indexer) Modules() map[string]*ModuleManager {
	return i.modules
}
