package schematesting

import (
	"bytes"
	"fmt"

	"cosmossdk.io/schema"
)

func DiffObjectKeys(fields []schema.Field, expected, actual any) string {
	n := len(fields)
	switch n {
	case 0:
		return ""
	case 1:
		return DiffFieldValues(fields[0], expected, actual)
	default:
		actualValues, ok := actual.([]interface{})
		if !ok {
			return fmt.Sprintf("ERROR: expected array of values for actual, got %v\n", actual)
		}
		expectedValues, ok := expected.([]interface{})
		if !ok {
			return fmt.Sprintf("ERROR: expected array of values for expected, got %v\n", actual)
		}
		res := ""
		for i := 0; i < n; i++ {
			res += DiffFieldValues(fields[i], expectedValues[i], actualValues[i])
		}
		return res
	}
}

func DiffObjectValues(fields []schema.Field, expected, actual any) string {
	if len(fields) == 0 {
		return ""
	}

	_, ok := expected.(schema.ValueUpdates)
	_, ok2 := expected.(schema.ValueUpdates)

	if ok || ok2 {
		return fmt.Sprintf("ValueUpdates is not expected when comparing state")
	}

	return DiffObjectKeys(fields, expected, actual)
}

func DiffFieldValues(field schema.Field, expected, actual any) string {
	if field.Nullable {
		if expected == nil {
			if actual == nil {
				return ""
			} else {
				return fmt.Sprintf("%s: expected nil, got %v\n", field.Name, actual)
			}
		} else if actual == nil {
			return fmt.Sprintf("%s: expected %v, got nil\n", field.Name, expected)
		}
	}

	eq, err := CompareKindValues(field.Kind, actual, expected)
	if err != nil {
		return fmt.Sprintf("%s: ERROR: %v\n", field.Name, err)
	}
	if !eq {
		return fmt.Sprintf("%s: expected %v, got %v\n", field.Name, expected, actual)
	}
	return ""
}

func CompareKindValues(kind schema.Kind, expected, actual any) (bool, error) {
	if kind.ValidateValueType(expected) != nil {
		return false, fmt.Errorf("unexpected type %T for kind %s", expected, kind)
	}

	if kind.ValidateValueType(actual) != nil {
		return false, fmt.Errorf("unexpected type %T for kind %s", actual, kind)
	}

	switch kind {
	case schema.BytesKind, schema.JSONKind, schema.AddressKind:
		if !bytes.Equal(expected.([]byte), actual.([]byte)) {
			return false, nil
		}
	default:
		if expected != actual {
			return false, nil
		}
	}
	return true, nil
}

//
//// ObjectUpdatesEqual checks if the two object updates are equal.
//// An error is only returned when there when an unexpected error.
//func ObjectUpdatesEqual(o schema.ObjectType, x, y schema.ObjectUpdate) (bool, error) {
//	if x.TypeName != o.Name {
//		return false, fmt.Errorf("unexpected type: %s", x.TypeName)
//	}
//
//	if y.TypeName != o.Name {
//		return false, fmt.Errorf("unexpected type: %s", x.TypeName)
//	}
//
//	if x.Delete != y.Delete {
//		return false, nil
//	}
//
//	if !ObjectKeysEqual(o.KeyFields, x.Key, y.Key) {
//		return false, nil
//	}
//
//	if x.Delete {
//		return true, nil
//	}
//
//	return ObjectValuesEqual(o.ValueFields, x.Value, y.Value)
//}
//
//// ObjectKeysEqual checks if the two object keys are equal for the provided fields.
//// An error is only returned when there when an unexpected error.
//func ObjectKeysEqual(fields []schema.Field, x, y interface{}) bool {
//	n := len(fields)
//	switch n {
//	case 0:
//		return true
//	case 1:
//		return ValuesEqual(fields[0].Kind, x, y)
//	default:
//		xs := x.([]interface{})
//		ys := y.([]interface{})
//		for i := 0; i < n; i++ {
//			if !ValuesEqual(fields[i].Kind, xs[i], ys[i]) {
//				return false
//			}
//		}
//		return true
//	}
//}
//
//// ObjectValuesEqual checks if the two object values are equal for the provided fields.
//func ObjectValuesEqual(fields []schema.Field, x, y interface{}) (bool, error) {
//	_, ok1 := x.(schema.ValueUpdates)
//	_, ok2 := y.(schema.ValueUpdates)
//	if ok1 || ok2 {
//		vmap1, err := collectValueUpdates(fields, x)
//		if err != nil {
//			return false, err
//		}
//		vmap2, err := collectValueUpdates(fields, x)
//		if err != nil {
//			return false, err
//		}
//
//		for _, f := range fields {
//			v1, ok1 := vmap1[f.Name]
//			v2, ok2 := vmap2[f.Name]
//			if ok1 != ok2 {
//				return false, nil
//			}
//
//			// both empty
//			if !ok1 {
//				return true, nil
//			}
//
//			if !ValuesEqual(f.Kind, v1, v2) {
//				return false, nil
//			}
//		}
//	}
//	return ObjectKeysEqual(fields, x, y), nil
//}
//
//func collectValueUpdates(fields []schema.Field, x interface{}) (map[string]interface{}, error) {
//	vmap := map[string]interface{}{}
//	if vu, ok := x.(schema.ValueUpdates); ok {
//		err := vu.Iterate(func(k string, v interface{}) bool {
//			vmap[k] = v
//			return true
//		})
//		return vmap, err
//	} else if xs, ok := x.([]interface{}); ok {
//		for i, f := range fields {
//			vmap[f.Name] = xs[i]
//		}
//		return vmap, nil
//	} else {
//		return nil, fmt.Errorf("unexpected type: %T", x)
//	}
//}
//
//// ValuesEqual checks if the two values are equal for the provided kind. It checks Kind.ValidateValueType first.
//func ValuesEqual(t schema.Kind, x, y interface{}) bool {
//	if t.ValidateValueType(x) != nil {
//		return false
//	}
//
//	if t.ValidateValueType(y) != nil {
//		return false
//	}
//
//	switch t {
//	case schema.BytesKind, schema.JSONKind, schema.AddressKind:
//		return bytes.Equal(x.([]byte), y.([]byte))
//	default:
//		return x == y
//	}
//}
