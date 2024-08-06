package postgres

import (
	"context"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/view"
)

var _ view.AppData = &Indexer{}

func (i *Indexer) AppState() view.AppState {
	return i
}

func (i *Indexer) BlockNum() uint64 {
	//TODO implement me
	panic("implement me")
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
	//TODO implement me
	panic("implement me")
}

func (m *moduleView) NumObjectCollections() int {
	//TODO implement me
	panic("implement me")
}

type objectView struct {
	ObjectIndexer
	ctx  context.Context
	conn DBConn
}

func (tm *objectView) ObjectType() schema.ObjectType {
	return tm.typ
}

func (tm *objectView) GetObject(key any) (update schema.ObjectUpdate, found bool) {
	//TODO implement me
	panic("implement me")
}

func (tm *objectView) AllState(f func(schema.ObjectUpdate) bool) {
	//TODO implement me
	panic("implement me")
}

func (tm *objectView) Len() int {
	n, err := tm.Count(tm.ctx, tm.conn)
	if err != nil {
		panic(err)
	}
	return n
}
