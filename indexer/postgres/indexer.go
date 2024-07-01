package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"cosmossdk.io/schema/appdata"
)

type Indexer struct {
	ctx context.Context
	// TODO: make private or internal
	Db      *sql.DB
	Tx      *sql.Tx
	Modules map[string]*ModuleManager
	options Options
}

type Options struct {
	Driver          string
	ConnectionURL   string
	RetainDeletions bool
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
		ctx: ctx,
		Db:  db,
		Tx:  tx,
		// TODO: make private or internal
		Modules: map[string]*ModuleManager{},
		options: opts,
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
	_, ok := i.Modules[moduleName]
	if ok {
		return fmt.Errorf("module %s already initialized", moduleName)
	}

	mm := newModuleManager(moduleName, modSchema, i.options)
	i.Modules[moduleName] = mm

	return mm.Init(i.ctx, i.Tx)
}

func (i *Indexer) onObjectUpdate(data appdata.ObjectUpdateData) error {
	module := data.ModuleName
	mod, ok := i.Modules[module]
	if !ok {
		return fmt.Errorf("module %s not initialized", module)
	}

	for _, update := range data.Updates {
		tm, ok := mod.Tables[update.TypeName]
		if !ok {
			return fmt.Errorf("object type %s not found in schema for module %s", update.TypeName, module)
		}

		var err error
		if update.Delete {
			err = tm.Delete(i.ctx, i.Tx, update.Key)
		} else {
			err = tm.InsertUpdate(i.ctx, i.Tx, update.Key, update.Value)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Indexer) commit(data appdata.CommitData) error {
	err := i.Tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	i.Tx, err = i.Db.BeginTx(i.ctx, nil)
	return err
}
