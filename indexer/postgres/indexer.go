package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"cosmossdk.io/schema/appdata"
)

type Indexer struct {
	ctx     context.Context
	db      *sql.DB
	tx      *sql.Tx
	modules map[string]*moduleManager
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

func (i *Indexer) Listener() appdata.Listener {
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

	mm := newModuleManager(moduleName, modSchema)
	i.modules[moduleName] = mm

	return mm.Init(i.ctx, i.tx)
}

func (i *Indexer) onObjectUpdate(data appdata.ObjectUpdateData) error {
	module := data.ModuleName
	update := data.Update
	mod, ok := i.modules[module]
	if !ok {
		return fmt.Errorf("module %s not initialized", module)
	}

	tm, ok := mod.tables[update.TypeName]
	if !ok {
		return fmt.Errorf("object type %s not found in schema for module %s", update.TypeName, module)
	}

	if update.Delete {
		return tm.Delete(i.ctx, i.tx, update.Key)
	} else {
		return tm.InsertUpdate(i.ctx, i.tx, update.Key, update.Value)
	}
}

func (i *Indexer) commit() error {
	err := i.tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	i.tx, err = i.db.BeginTx(i.ctx, nil)
	return err
}
