package postgres

import (
	"context"
	"database/sql"
	"strings"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/view"
)

var _ view.AppData = &indexerImpl{}

func (i *indexerImpl) AppState() view.AppState {
	return i
}

func (i *indexerImpl) BlockNum() (uint64, error) {
	var blockNum int64
	err := i.tx.QueryRow("SELECT coalesce(max(number), 0) FROM block").Scan(&blockNum)
	if err != nil {
		return 0, err
	}
	return uint64(blockNum), nil
}

type moduleView struct {
	moduleIndexer
	ctx  context.Context
	conn dbConn
}

func (i *indexerImpl) GetModule(moduleName string) (view.ModuleState, error) {
	mod, ok := i.modules[moduleName]
	if !ok {
		return nil, nil
	}
	return &moduleView{
		moduleIndexer: *mod,
		ctx:           i.ctx,
		conn:          i.tx,
	}, nil
}

func (i *indexerImpl) Modules(f func(modState view.ModuleState, err error) bool) {
	for _, mod := range i.modules {
		if !f(&moduleView{
			moduleIndexer: *mod,
			ctx:           i.ctx,
			conn:          i.tx,
		}, nil) {
			return
		}
	}
}

func (i *indexerImpl) NumModules() (int, error) {
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
		objectIndexer: *obj,
		ctx:           m.ctx,
		conn:          m.conn,
	}, nil
}

func (m *moduleView) ObjectCollections(f func(value view.ObjectCollection, err error) bool) {
	for _, obj := range m.tables {
		if !f(&objectView{
			objectIndexer: *obj,
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
	objectIndexer
	ctx  context.Context
	conn dbConn
}

func (tm *objectView) ObjectType() schema.StateObjectType {
	return tm.typ
}

func (tm *objectView) GetObject(key interface{}) (update schema.StateObjectUpdate, found bool, err error) {
	return tm.get(tm.ctx, tm.conn, key)
}

func (tm *objectView) AllState(f func(schema.StateObjectUpdate, error) bool) {
	buf := new(strings.Builder)
	err := tm.selectAllSql(buf)
	if err != nil {
		panic(err)
	}

	sqlStr := buf.String()
	if tm.options.logger != nil {
		tm.options.logger.Debug("Select", "sql", sqlStr)
	}

	rows, err := tm.conn.QueryContext(tm.ctx, sqlStr)
	if err != nil {
		panic(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			panic(err)
		}
	}(rows)

	for rows.Next() {
		update, found, err := tm.readRow(rows)
		if err == nil && !found {
			err = sql.ErrNoRows
		}
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
