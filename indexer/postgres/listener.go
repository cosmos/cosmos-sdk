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
		OnObjectUpdate: func(data appdata.ObjectUpdateData) error {
			module := data.ModuleName
			mod, ok := i.modules[module]
			if !ok {
				return fmt.Errorf("module %s not initialized", module)
			}

			for _, update := range data.Updates {
				if i.logger != nil {
					i.logger.Debug("OnObjectUpdate", "module", module, "type", update.TypeName, "key", update.Key, "delete", update.Delete, "value", update.Value)
				}
				tm, ok := mod.tables[update.TypeName]
				if !ok {
					return fmt.Errorf("object type %s not found in schema for module %s", update.TypeName, module)
				}

				var err error
				if update.Delete {
					err = tm.delete(i.ctx, i.tx, update.Key)
				} else {
					err = tm.insertUpdate(i.ctx, i.tx, update.Key, update.Value)
				}
				if err != nil {
					return err
				}
			}
			return nil
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
