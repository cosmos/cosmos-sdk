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
			var (
				headerBz []byte
				err      error
			)

			if data.HeaderJSON != nil {
				headerBz, err = data.HeaderJSON()
				if err != nil {
					return err
				}
			} else if data.HeaderBytes != nil {
				headerBz, err = data.HeaderBytes()
				if err != nil {
					return err
				}
			}

			// TODO: verify the format of headerBz, otherwise we'll get `ERROR: invalid input syntax for type json (SQLSTATE 22P02)`
			_, err = i.tx.Exec("INSERT INTO block (number, header) VALUES ($1, $2)", data.Height, headerBz)

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
			fmt.Println("Commit ERROR 000")
			err := i.tx.Commit()
			fmt.Println("Commit ERROR 111", err)
			if err != nil {
				return nil, err
			}

			i.tx, err = i.db.BeginTx(i.ctx, nil)
			fmt.Println("Commit ERROR 2222", err)
			return nil, err
		},
	}
}
