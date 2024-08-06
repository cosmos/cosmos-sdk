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

func (i *Indexer) BlockNum() uint64 {
	return 0
}

type moduleView struct {
	ModuleIndexer
	ctx  context.Context
	conn DBConn
}

func (i *Indexer) GetModule(moduleName string) (view.ModuleState, bool) {
	mod, ok := i.modules[moduleName]
	if !ok {
		return nil, false
	}
	return &moduleView{
		ModuleIndexer: *mod,
		ctx:           i.ctx,
		conn:          i.tx,
	}, true
}

func (i *Indexer) Modules(f func(moduleName string, modState view.ModuleState) bool) {
	for name, mod := range i.modules {
		if !f(name, &moduleView{
			ModuleIndexer: *mod,
			ctx:           i.ctx,
			conn:          i.tx,
		}) {
			return
		}
	}
}

func (i *Indexer) NumModules() int {
	return len(i.modules)
}

func (m *moduleView) ModuleSchema() schema.ModuleSchema {
	return m.schema
}

func (m *moduleView) GetObjectCollection(objectType string) (view.ObjectCollection, bool) {
	obj, ok := m.tables[objectType]
	if !ok {
		return nil, false
	}
	return &objectView{
		ObjectIndexer: *obj,
		ctx:           m.ctx,
		conn:          m.conn,
	}, true
}

func (m *moduleView) ObjectCollections(f func(value view.ObjectCollection) bool) {
	for _, obj := range m.tables {
		if !f(&objectView{
			ObjectIndexer: *obj,
			ctx:           m.ctx,
			conn:          m.conn,
		}) {
			return
		}
	}
}

func (m *moduleView) NumObjectCollections() int {
	return len(m.tables)
}

type objectView struct {
	ObjectIndexer
	ctx  context.Context
	conn DBConn
}

func (tm *objectView) ObjectType() schema.ObjectType {
	return tm.typ
}

func (tm *objectView) GetObject(key interface{}) (update schema.ObjectUpdate, found bool) {
	update, err := tm.Get(tm.ctx, tm.conn, key)
	if err != nil {
		return schema.ObjectUpdate{}, false
	}
	return update, true
}

func (tm *objectView) AllState(f func(schema.ObjectUpdate) bool) {
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
		if err != nil {
			panic(err)
		}
		if !f(update) {
			return
		}
	}
}

func (tm *objectView) Len() int {
	n, err := tm.Count(tm.ctx, tm.conn)
	if err != nil {
		panic(err)
	}
	return n
}
