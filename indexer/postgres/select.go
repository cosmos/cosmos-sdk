package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"cosmossdk.io/schema"
)

// Count returns the number of rows in the table.
func (tm *ObjectIndexer) Count(ctx context.Context, conn DBConn) (int, error) {
	row := conn.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %q;", tm.TableName()))
	var count int
	err := row.Scan(&count)
	return count, err
}

// Exists checks if a row with the provided key exists in the table.
func (tm *ObjectIndexer) Exists(ctx context.Context, conn DBConn, key interface{}) (bool, error) {
	buf := new(strings.Builder)
	params, err := tm.ExistsSqlAndParams(buf, key)
	if err != nil {
		return false, err
	}

	return tm.checkExists(ctx, conn, buf.String(), params)
}

// ExistsSqlAndParams generates a SELECT statement to check if a row with the provided key exists in the table.
func (tm *ObjectIndexer) ExistsSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "SELECT 1 FROM %q", tm.TableName())
	if err != nil {
		return nil, err
	}

	_, keyParams, err := tm.WhereSqlAndParams(w, key, 1)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, ";")
	return keyParams, err
}

func (tm *ObjectIndexer) Get(ctx context.Context, conn DBConn, key interface{}) (schema.ObjectUpdate, error) {
	buf := new(strings.Builder)
	params, err := tm.GetSqlAndParams(buf, key)
	if err != nil {
		return schema.ObjectUpdate{}, err
	}

	sqlStr := buf.String()
	tm.options.Logger("Select", "sql", sqlStr, "params", params)

	row := conn.QueryRowContext(ctx, sqlStr, params...)
	return tm.readRow(row)
}

func (tm *ObjectIndexer) SelectAllSql(w io.Writer) error {
	err := tm.selectAllClause(w)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, ";")
	return err
}

func (tm *ObjectIndexer) GetSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
	err := tm.selectAllClause(w)
	if err != nil {
		return nil, err
	}

	keyParams, keyCols, err := tm.bindKeyParams(key)
	if err != nil {
		return nil, err
	}

	_, keyParams, err = tm.WhereSql(w, keyParams, keyCols, 1)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, ";")
	return keyParams, err
}

func (tm *ObjectIndexer) selectAllClause(w io.Writer) error {
	allFields := make([]string, 0, len(tm.typ.KeyFields)+len(tm.typ.ValueFields))

	for _, field := range tm.typ.KeyFields {
		colName, err := tm.updatableColumnName(field)
		if err != nil {
			return err
		}
		allFields = append(allFields, colName)
	}

	for _, field := range tm.typ.ValueFields {
		colName, err := tm.updatableColumnName(field)
		if err != nil {
			return err
		}
		allFields = append(allFields, colName)
	}

	if !tm.options.DisableRetainDeletions && tm.typ.RetainDeletions {
		allFields = append(allFields, "_deleted")
	}

	_, err := fmt.Fprintf(w, "SELECT %s FROM %q", strings.Join(allFields, ", "), tm.TableName())
	if err != nil {
		return err
	}

	return nil
}

func (tm *ObjectIndexer) readRow(row interface{ Scan(...interface{}) error }) (schema.ObjectUpdate, error) {
	var res []interface{}
	for _, f := range tm.typ.KeyFields {
		res = append(res, tm.colBindValue(f))
	}

	for _, f := range tm.typ.ValueFields {
		res = append(res, tm.colBindValue(f))
	}

	if !tm.options.DisableRetainDeletions && tm.typ.RetainDeletions {
		res = append(res, new(bool))
	}

	err := row.Scan(res...)
	if err != nil {
		return schema.ObjectUpdate{}, err
	}

	var keys []interface{}
	for _, field := range tm.typ.KeyFields {
		x, err := tm.readCol(field, res[0])
		if err != nil {
			return schema.ObjectUpdate{}, err
		}
		keys = append(keys, x)
		res = res[1:]
	}

	var key interface{} = keys
	if len(keys) == 1 {
		key = keys[0]
	}

	var values []interface{}
	for _, field := range tm.typ.ValueFields {
		x, err := tm.readCol(field, res[0])
		if err != nil {
			return schema.ObjectUpdate{}, err
		}
		values = append(values, x)
		res = res[1:]
	}

	var value interface{} = values
	if len(values) == 1 {
		value = values[0]
	}

	update := schema.ObjectUpdate{
		TypeName: tm.typ.Name,
		Key:      key,
		Value:    value,
	}

	if !tm.options.DisableRetainDeletions && tm.typ.RetainDeletions {
		deleted := res[0].(*bool)
		if *deleted {
			update.Delete = true
		}
	}

	return update, nil
}

func (tm *ObjectIndexer) colBindValue(field schema.Field) interface{} {
	switch field.Kind {
	case schema.BoolKind:
		return new(bool)
	case schema.Uint8Kind:
		return new(uint8)
	case schema.Uint16Kind:
		return new(uint16)
	case schema.Uint32Kind:
		return new(uint32)
	case schema.Uint64Kind:
		return new(uint64)
	case schema.Int8Kind:
		return new(int8)
	case schema.Int16Kind:
		return new(int16)
	case schema.Int32Kind:
		return new(int32)
	case schema.Int64Kind:
		return new(int64)
	case schema.Float32Kind:
		return new(float32)
	case schema.Float64Kind:
		return new(float64)
	case schema.BytesKind, schema.AddressKind:
		return new([]byte)
	case schema.TimeKind, schema.DurationKind:
		return new(int64)
	case schema.JSONKind:
		return new(json.RawMessage)
	default:
		return new(string)
	}
}

func (tm *ObjectIndexer) readCol(field schema.Field, value interface{}) (interface{}, error) {
	value = *value.(*interface{})
	switch field.Kind {
	case schema.TimeKind:
		return time.Unix(0, value.(int64)), nil
	case schema.DurationKind:
		return time.Duration(value.(int64)), nil
	default:
		return value, nil
	}
}

// Equals checks if a row with the provided key and value exists.
func (tm *ObjectIndexer) Equals(ctx context.Context, conn DBConn, key, val interface{}) (bool, error) {
	buf := new(strings.Builder)
	params, err := tm.EqualsSqlAndParams(buf, key, val)
	if err != nil {
		return false, err
	}

	return tm.checkExists(ctx, conn, buf.String(), params)
}

// EqualsSqlAndParams generates a SELECT statement to check if a row with the provided key and value exists in the table.
func (tm *ObjectIndexer) EqualsSqlAndParams(w io.Writer, key, val interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "SELECT 1 FROM %q", tm.TableName())
	if err != nil {
		return nil, err
	}

	keyParams, keyCols, err := tm.bindKeyParams(key)
	if err != nil {
		return nil, err
	}

	valueParams, valueCols, err := tm.bindValueParams(val)
	if err != nil {
		return nil, err
	}

	allParams := make([]interface{}, 0, len(keyParams)+len(valueParams))
	allParams = append(allParams, keyParams...)
	allParams = append(allParams, valueParams...)

	allCols := make([]string, 0, len(keyCols)+len(valueCols))
	allCols = append(allCols, keyCols...)
	allCols = append(allCols, valueCols...)

	_, allParams, err = tm.WhereSql(w, allParams, allCols, 1)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, ";")
	return allParams, err
}

// checkExists checks if a row exists in the table.
func (tm *ObjectIndexer) checkExists(ctx context.Context, conn DBConn, sqlStr string, params []interface{}) (bool, error) {
	tm.options.Logger("Select", "sql", sqlStr, "params", params)
	var res interface{}
	// TODO check for multiple rows which would be a logic error
	err := conn.QueryRowContext(ctx, sqlStr, params...).Scan(&res)
	switch err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}
