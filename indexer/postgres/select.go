package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/schema"
)

// Count returns the number of rows in the table.
func (tm *objectIndexer) count(ctx context.Context, conn dbConn) (int, error) {
	sqlStr := fmt.Sprintf("SELECT COUNT(*) FROM %q;", tm.tableName())
	if tm.options.logger != nil {
		tm.options.logger.Debug("Count", "sql", sqlStr)
	}
	row := conn.QueryRowContext(ctx, sqlStr)
	var count int
	err := row.Scan(&count)
	return count, err
}

// exists checks if a row with the provided key exists in the table.
func (tm *objectIndexer) exists(ctx context.Context, conn dbConn, key interface{}) (bool, error) {
	buf := new(strings.Builder)
	params, err := tm.existsSqlAndParams(buf, key)
	if err != nil {
		return false, err
	}

	return tm.checkExists(ctx, conn, buf.String(), params)
}

// checkExists checks if a row exists in the table.
func (tm *objectIndexer) checkExists(ctx context.Context, conn dbConn, sqlStr string, params []interface{}) (bool, error) {
	if tm.options.logger != nil {
		tm.options.logger.Debug("Check exists", "sql", sqlStr, "params", params)
	}
	var res interface{}
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

// existsSqlAndParams generates a SELECT statement to check if a row with the provided key exists in the table.
func (tm *objectIndexer) existsSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "SELECT 1 FROM %q", tm.tableName())
	if err != nil {
		return nil, err
	}

	_, keyParams, err := tm.whereSqlAndParams(w, key, 1)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, ";")
	return keyParams, err
}

func (tm *objectIndexer) get(ctx context.Context, conn dbConn, key interface{}) (schema.StateObjectUpdate, bool, error) {
	buf := new(strings.Builder)
	params, err := tm.getSqlAndParams(buf, key)
	if err != nil {
		return schema.StateObjectUpdate{}, false, err
	}

	sqlStr := buf.String()
	if tm.options.logger != nil {
		tm.options.logger.Debug("Get", "sql", sqlStr, "params", params)
	}

	row := conn.QueryRowContext(ctx, sqlStr, params...)
	return tm.readRow(row)
}

func (tm *objectIndexer) selectAllSql(w io.Writer) error {
	err := tm.selectAllClause(w)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, ";")
	return err
}

func (tm *objectIndexer) getSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
	err := tm.selectAllClause(w)
	if err != nil {
		return nil, err
	}

	keyParams, keyCols, err := tm.bindKeyParams(key)
	if err != nil {
		return nil, err
	}

	_, keyParams, err = tm.whereSql(w, keyParams, keyCols, 1)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, ";")
	return keyParams, err
}

func (tm *objectIndexer) selectAllClause(w io.Writer) error {
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

	if !tm.options.disableRetainDeletions && tm.typ.RetainDeletions {
		allFields = append(allFields, "_deleted")
	}

	_, err := fmt.Fprintf(w, "SELECT %s FROM %q", strings.Join(allFields, ", "), tm.tableName())
	if err != nil {
		return err
	}

	return nil
}

func (tm *objectIndexer) readRow(row interface{ Scan(...interface{}) error }) (schema.StateObjectUpdate, bool, error) {
	var res []interface{}
	for _, f := range tm.typ.KeyFields {
		res = append(res, tm.colBindValue(f))
	}

	for _, f := range tm.typ.ValueFields {
		res = append(res, tm.colBindValue(f))
	}

	if !tm.options.disableRetainDeletions && tm.typ.RetainDeletions {
		res = append(res, new(bool))
	}

	err := row.Scan(res...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return schema.StateObjectUpdate{}, false, err
		}
		return schema.StateObjectUpdate{}, false, err
	}

	var keys []interface{}
	for _, field := range tm.typ.KeyFields {
		x, err := tm.readCol(field, res[0])
		if err != nil {
			return schema.StateObjectUpdate{}, false, err
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
			return schema.StateObjectUpdate{}, false, err
		}
		values = append(values, x)
		res = res[1:]
	}

	var value interface{} = values
	if len(values) == 1 {
		value = values[0]
	}

	update := schema.StateObjectUpdate{
		TypeName: tm.typ.Name,
		Key:      key,
		Value:    value,
	}

	if !tm.options.disableRetainDeletions && tm.typ.RetainDeletions {
		deleted := res[0].(*bool)
		if *deleted {
			update.Delete = true
		}
	}

	return update, true, nil
}

func (tm *objectIndexer) colBindValue(field schema.Field) interface{} {
	switch field.Kind {
	case schema.BytesKind:
		return new(interface{})
	default:
		return new(sql.NullString)
	}
}

func (tm *objectIndexer) readCol(field schema.Field, value interface{}) (interface{}, error) {
	switch field.Kind {
	case schema.BytesKind:
		// for bytes types we either get []byte or nil
		value = *value.(*interface{})
		return value, nil
	default:
	}

	nullStr := *value.(*sql.NullString)
	if field.Nullable {
		if !nullStr.Valid {
			return nil, nil
		}
	}
	str := nullStr.String

	switch field.Kind {
	case schema.StringKind, schema.EnumKind, schema.IntegerKind, schema.DecimalKind:
		return str, nil
	case schema.Uint8Kind:
		value, err := strconv.ParseUint(str, 10, 8)
		return uint8(value), err
	case schema.Uint16Kind:
		value, err := strconv.ParseUint(str, 10, 16)
		return uint16(value), err
	case schema.Uint32Kind:
		value, err := strconv.ParseUint(str, 10, 32)
		return uint32(value), err
	case schema.Uint64Kind:
		value, err := strconv.ParseUint(str, 10, 64)
		return value, err
	case schema.Int8Kind:
		value, err := strconv.ParseInt(str, 10, 8)
		return int8(value), err
	case schema.Int16Kind:
		value, err := strconv.ParseInt(str, 10, 16)
		return int16(value), err
	case schema.Int32Kind:
		value, err := strconv.ParseInt(str, 10, 32)
		return int32(value), err
	case schema.Int64Kind:
		value, err := strconv.ParseInt(str, 10, 64)
		return value, err
	case schema.Float32Kind:
		value, err := strconv.ParseFloat(str, 32)
		return float32(value), err
	case schema.Float64Kind:
		value, err := strconv.ParseFloat(str, 64)
		return value, err
	case schema.BoolKind:
		value, err := strconv.ParseBool(str)
		return value, err
	case schema.JSONKind:
		return json.RawMessage(str), nil
	case schema.TimeKind:
		value, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, err
		}
		return time.Unix(0, value), nil
	case schema.DurationKind:
		value, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, err
		}
		return time.Duration(value), nil
	case schema.AddressKind:
		return tm.options.addressCodec.StringToBytes(str)
	default:
		return value, nil
	}
}
