package postgres

import (
	"fmt"

	"cosmossdk.io/schema/appdata"
)

func (i *indexerImpl) listener() appdata.Listener {
	return appdata.Listener{
		InitializeModuleData: func(data appdata.ModuleInitializationData) error {
			moduleName := data.ModuleName
			modSchema := data.Schema
			_, ok := i.modules[moduleName]
			if ok {
				return fmt.Errorf("module %s already initialized", moduleName)
			}

			mm := newModuleIndexer(moduleName, modSchema, i.opts)
			i.modules[moduleName] = mm

			return mm.initializeSchema(i.ctx, i.tx)
		},
		StartBlock: func(data appdata.StartBlockData) error {
			_, err := i.tx.Exec("INSERT INTO block (number) VALUES ($1)", data.Height)
			return err
		},
		Commit: func(data appdata.CommitData) (func() error, error) {
			err := i.tx.Commit()
			if err != nil {
				return nil, err
			}

			i.tx, err = i.db.BeginTx(i.ctx, nil)
			return nil, err
		},
	}
}
