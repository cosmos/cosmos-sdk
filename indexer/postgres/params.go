package postgres

import (
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/schema"
)

// bindKeyParams binds the key to the key columns.
func (tm *objectIndexer) bindKeyParams(key interface{}) ([]interface{}, []string, error) {
	n := len(tm.typ.KeyFields)
	if n == 0 {
		// singleton, set _id = 1
		return []interface{}{1}, []string{"_id"}, nil
	} else if n == 1 {
		return tm.bindParams(tm.typ.KeyFields, []interface{}{key})
	} else {
		key, ok := key.([]interface{})
		if !ok {
			return nil, nil, errors.New("expected key to be a slice")
		}

		return tm.bindParams(tm.typ.KeyFields, key)
	}
}

func (tm *objectIndexer) bindValueParams(value interface{}) (params []interface{}, valueCols []string, err error) {
	n := len(tm.typ.ValueFields)
	if n == 0 {
		return nil, nil, nil
	} else if valueUpdates, ok := value.(schema.ValueUpdates); ok {
		var e error
		var fields []schema.Field
		var params []interface{}
		if err := valueUpdates.Iterate(func(name string, value interface{}) bool {
			field, ok := tm.valueFields[name]
			if !ok {
				e = fmt.Errorf("unknown column %q", name)
				return false
			}
			fields = append(fields, field)
			params = append(params, value)
			return true
		}); err != nil {
			return nil, nil, err
		}
		if e != nil {
			return nil, nil, e
		}

		return tm.bindParams(fields, params)
	} else if n == 1 {
		return tm.bindParams(tm.typ.ValueFields, []interface{}{value})
	} else {
		values, ok := value.([]interface{})
		if !ok {
			return nil, nil, errors.New("expected values to be a slice")
		}

		return tm.bindParams(tm.typ.ValueFields, values)
	}
}

func (tm *objectIndexer) bindParams(fields []schema.Field, values []interface{}) ([]interface{}, []string, error) {
	names := make([]string, 0, len(fields))
	params := make([]interface{}, 0, len(fields))
	for i, field := range fields {
		if i >= len(values) {
			return nil, nil, fmt.Errorf("missing value for field %q", field.Name)
		}

		param, err := tm.bindParam(field, values[i])
		if err != nil {
			return nil, nil, err
		}

		name, err := tm.updatableColumnName(field)
		if err != nil {
			return nil, nil, err
		}

		names = append(names, name)
		params = append(params, param)
	}
	return params, names, nil
}

func (tm *objectIndexer) bindParam(field schema.Field, value interface{}) (param interface{}, err error) {
	param = value
	if value == nil {
		if !field.Nullable {
			return nil, fmt.Errorf("expected non-null value for field %q", field.Name)
		}
	} else if field.Kind == schema.TimeKind {
		t, ok := value.(time.Time)
		if !ok {
			return nil, fmt.Errorf("expected time.Time value for field %q, got %T", field.Name, value)
		}

		param = t.UnixNano()
	} else if field.Kind == schema.DurationKind {
		t, ok := value.(time.Duration)
		if !ok {
			return nil, fmt.Errorf("expected time.Duration value for field %q, got %T", field.Name, value)
		}

		param = int64(t)
	} else if field.Kind == schema.AddressKind {
		param, err = tm.options.addressCodec.BytesToString(value.([]byte))
		if err != nil {
			return nil, fmt.Errorf("address encoding failed for field %q: %w", field.Name, err)
		}
	}
	return
}
