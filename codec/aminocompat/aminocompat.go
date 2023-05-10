package aminocompat

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

var jsonMarshaler = reflect.TypeOf(new(json.Marshaler))

func AllClear(v any) error {
	rv := reflect.ValueOf(v)
	return allClear(rv)
}

func allClear(v reflect.Value) error {
	if !v.IsValid() {
		return errors.New("not valid")
	}

	// Derefence the pointer.
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = reflect.Indirect(v)
	}

	// Now we can walk the value as we've dereferenced it.
	// 1. If it a json.Marshaler, then skip it.
	if vt := v.Type(); vt.Kind() == reflect.Interface && vt.Implements(jsonMarshaler) {
		return nil
	}

	switch vkind := v.Kind(); vkind {
	case reflect.Int, reflect.Uint, reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64:
		return nil

	case reflect.String:
		return nil

	case reflect.Map, reflect.Complex64, reflect.Complex128, reflect.Float32, reflect.Float64:
		return fmt.Errorf("not supported: %v", vkind)

	case reflect.Struct:
		// Walk the struct fields.
		typ := v.Type()
		for i, n := 0, v.NumField(); i < n; i++ {
			fi := v.Field(i)

			for fi.Kind() == reflect.Ptr {
				fi = reflect.Indirect(fi)
			}
			if !fi.IsValid() {
				return fmt.Errorf("field #%d is invalid", i)
			}

			ti := typ.Field(i)
			if unexportedFieldName(ti.Name) {
				continue
			}

			// Now check this field as well.
			if err := allClear(fi); err != nil {
				return fmt.Errorf("field #%d: %w", i, err)
			}
		}
		return nil

	case reflect.Slice, reflect.Array:
		// To ensure thoroughness, let's just traverse every single element.
                // If we've got a slice of a slice or that composition we need to iterate on that.
		for i, n := 0, v.Len(); i < n; i++ {
			evi := v.Index(i)
			if err := allClear(evi); err != nil {
				return fmt.Errorf("field #%d: %w", i, err)
			}
		}
		return nil

	case reflect.Interface:
		// Walk through the concrete type.
		return allClear(v.Elem())
	}

	return fmt.Errorf("not supported: %v", v.Kind())
}

func unexportedFieldName(name string) bool {
	return len(name) > 0 && name[0] >= 'a' && name[0] <= 'z'
}
