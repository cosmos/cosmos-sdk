package postgres

import (
	"encoding/json"
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
			err := i.tx.Commit()
			if err != nil {
				return nil, err
			}

			i.tx, err = i.db.BeginTx(i.ctx, nil)
			return nil, err
		},
		OnTx:    txListener(i),
		OnEvent: eventListener(i),
	}
}

func txListener(i *indexerImpl) func(data appdata.TxData) error {
	return func(td appdata.TxData) error {
		var bz []byte
		if td.Bytes != nil {
			var err error
			bz, err = td.Bytes()
			if err != nil {
				return err
			}
		}

		var jsonData json.RawMessage
		if td.JSON != nil {
			var err error
			jsonData, err = td.JSON()
			if err != nil {
				return err
			}
		}

		_, err := i.tx.Exec("INSERT INTO tx (block_number, index_in_block, data, bytes) VALUES ($1, $2, $3, $4)",
			td.BlockNumber, td.TxIndex, jsonData, bz)

		return err
	}
}

func eventListener(i *indexerImpl) func(data appdata.EventData) error {
	return func(data appdata.EventData) error {
		for _, e := range data.Events {
			var jsonData json.RawMessage

			if e.Data != nil {
				var err error
				jsonData, err = e.Data()
				if err != nil {
					return fmt.Errorf("failed to get event data: %w", err)
				}
			} else if e.Attributes != nil {
				attrs, err := e.Attributes()
				if err != nil {
					return fmt.Errorf("failed to get event attributes: %w", err)
				}

				attrsMap := map[string]interface{}{}
				for _, attr := range attrs {
					attrsMap[attr.Key] = attr.Value
				}

				jsonData, err = json.Marshal(attrsMap)
				if err != nil {
					return fmt.Errorf("failed to marshal event attributes: %w", err)
				}
			}

			_, err := i.tx.Exec("INSERT INTO event (block_number, block_stage, tx_index, msg_index, event_index, type, data) VALUES ($1, $2, $3, $4, $5, $6, $7)",
				e.BlockNumber, e.BlockStage, e.TxIndex, e.MsgIndex, e.EventIndex, e.Type, jsonData)
			if err != nil {
				return fmt.Errorf("failed to index event: %w", err)
			}
		}
		return nil
	}
}
