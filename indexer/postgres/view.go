package postgres

import (
	"context"
	"strings"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/view"
)

var _ view.AppData = &Indexer{}

func (i *Indexer) AppState() view.AppState {
	return i
}

func (i *Indexer) BlockNum() (uint64, error) {
	var blockNum int64
	err := i.tx.QueryRow("SELECT max(number) FROM block").Scan(blockNum)
	if err != nil {
		return 0, err
	}
	return uint64(blockNum), nil
}

type moduleView struct {
	ModuleIndexer
	ctx  context.Context
	conn DBConn
}

func (i *Indexer) GetModule(moduleName string) (view.ModuleState, error) {
	mod, ok := i.modules[moduleName]
	if !ok {
		return nil, nil
	}
	return &moduleView{
		ModuleIndexer: *mod,
		ctx:           i.ctx,
		conn:          i.tx,
	}, nil
}

func (i *Indexer) Modules(f func(modState view.ModuleState, err error) bool) {
	for _, mod := range i.modules {
		if !f(&moduleView{
			ModuleIndexer: *mod,
			ctx:           i.ctx,
			conn:          i.tx,
		}, nil) {
			return
		}
	}
}

func (i *Indexer) NumModules() (int, error) {
	return len(i.modules), nil
}

func (m *moduleView) ModuleName() string {
	return m.moduleName
}

func (m *moduleView) ModuleSchema() schema.ModuleSchema {
	return m.schema
}

func (m *moduleView) GetObjectCollection(objectType string) (view.ObjectCollection, error) {
	obj, ok := m.tables[objectType]
	if !ok {
		return nil, nil
	}
	return &objectView{
		ObjectIndexer: *obj,
		ctx:           m.ctx,
		conn:          m.conn,
	}, nil
}

func (m *moduleView) ObjectCollections(f func(value view.ObjectCollection, err error) bool) {
	for _, obj := range m.tables {
		if !f(&objectView{
			ObjectIndexer: *obj,
			ctx:           m.ctx,
			conn:          m.conn,
		}, nil) {
			return
		}
	}
}

func (m *moduleView) NumObjectCollections() (int, error) {
	return len(m.tables), nil
}

type objectView struct {
	ObjectIndexer
	ctx  context.Context
	conn DBConn
}

func (tm *objectView) ObjectType() schema.ObjectType {
	return tm.typ
}

func (tm *objectView) GetObject(key interface{}) (update schema.ObjectUpdate, found bool, err error) {
	update, err = tm.Get(tm.ctx, tm.conn, key)
	if err != nil {
		return schema.ObjectUpdate{}, false, err
	}
	return update, true, err
}

func (tm *objectView) AllState(f func(schema.ObjectUpdate, error) bool) {
	buf := new(strings.Builder)
	err := tm.SelectAllSql(buf)
	if err != nil {
		panic(err)
	}
	sqlStr := buf.String()
	if tm.options.Logger != nil {
		tm.options.Logger("Select", "sql", sqlStr)
	}
	rows, err := tm.conn.QueryContext(tm.ctx, sqlStr)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		update, err := tm.readRow(rows)
		if !f(update, err) {
			return
		}
	}
}

func (tm *objectView) Len() (int, error) {
	n, err := tm.count(tm.ctx, tm.conn)
	if err != nil {
		return 0, err
	}
	return n, nil
}
