package postgres

import (
	"context"
	"database/sql"
	"fmt"

	indexerbase "cosmossdk.io/indexer/base"
)

type Indexer struct {
	ctx     context.Context
	db      *sql.DB
	tx      *sql.Tx
	modules map[string]*moduleManager
}

type moduleManager struct {
	moduleName string
	schema     indexerbase.ModuleSchema
	tables     map[string]*TableManager
}

type Options struct {
	Driver        string
	ConnectionURL string
}

func NewIndexer(ctx context.Context, opts Options) (*Indexer, error) {
	if opts.Driver == "" {
		opts.Driver = "pgx"
	}

	if opts.ConnectionURL == "" {
		return nil, fmt.Errorf("connection URL not set")
	}

	db, err := sql.Open(opts.Driver, opts.ConnectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	go func() {
		<-ctx.Done()
		err := db.Close()
		if err != nil {
			panic(fmt.Sprintf("failed to close database: %v", err))
		}
	}()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	return &Indexer{
		ctx:     ctx,
		db:      db,
		tx:      tx,
		modules: map[string]*moduleManager{},
	}, nil
}

func (i *Indexer) Listener() indexerbase.Listener {
	return indexerbase.Listener{
		InitializeModuleSchema: i.initModuleSchema,
	}
}

func (i *Indexer) initModuleSchema(moduleName string, schema indexerbase.ModuleSchema) error {
	_, ok := i.modules[moduleName]
	if ok {
		return fmt.Errorf("module %s already initialized", moduleName)
	}

	mm := &moduleManager{
		moduleName: moduleName,
		schema:     schema,
		tables:     map[string]*TableManager{},
	}

	for _, typ := range schema.ObjectTypes {
		tm := NewTableManager(moduleName, typ)
		mm.tables[typ.Name] = tm
		err := tm.CreateTable(i.ctx, i.tx)
		if err != nil {
			return fmt.Errorf("failed to create table for %s in module %s: %w", typ.Name, moduleName, err)
		}
	}

	return nil
}
