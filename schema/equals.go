package schema

import (
	"bytes"
	"fmt"
)

// ObjectUpdatesEqual checks if the two object updates are equal.
// An error is only returned when there when an unexpected error.
func (o ObjectType) ObjectUpdatesEqual(x, y ObjectUpdate) (bool, error) {
	if x.TypeName != o.Name {
		return false, fmt.Errorf("unexpected type: %s", x.TypeName)
	}

	if y.TypeName != o.Name {
		return false, fmt.Errorf("unexpected type: %s", x.TypeName)
	}

	if x.Delete != y.Delete {
		return false, nil
	}

	if !ObjectKeysEqual(o.KeyFields, x.Key, y.Key) {
		return false, nil
	}

	if x.Delete {
		return true, nil
	}

	return ObjectValuesEqual(o.ValueFields, x.Value, y.Value)
}

// ObjectKeysEqual checks if the two object keys are equal for the provided fields.
// An error is only returned when there when an unexpected error.
func ObjectKeysEqual(fields []Field, x, y interface{}) bool {
	n := len(fields)
	switch n {
	case 0:
		return true
	case 1:
		return fields[0].Kind.ValuesEqual(x, y)
	default:
		xs := x.([]interface{})
		ys := y.([]interface{})
		for i := 0; i < n; i++ {
			if !fields[i].Kind.ValuesEqual(xs[i], ys[i]) {
				return false
			}
		}
		return true
	}
}

// ObjectValuesEqual checks if the two object values are equal for the provided fields.
func ObjectValuesEqual(fields []Field, x, y interface{}) (bool, error) {
	_, ok1 := x.(ValueUpdates)
	_, ok2 := y.(ValueUpdates)
	if ok1 || ok2 {
		vmap1, err := collectValueUpdates(fields, x)
		if err != nil {
			return false, err
		}
		vmap2, err := collectValueUpdates(fields, x)
		if err != nil {
			return false, err
		}

		for _, f := range fields {
			v1, ok1 := vmap1[f.Name]
			v2, ok2 := vmap2[f.Name]
			if ok1 != ok2 {
				return false, nil
			}

			// both empty
			if !ok1 {
				return true, nil
			}

			if !f.Kind.ValuesEqual(v1, v2) {
				return false, nil
			}
		}
	}
	return ObjectKeysEqual(fields, x, y), nil
}

func collectValueUpdates(fields []Field, x interface{}) (map[string]interface{}, error) {
	vmap := map[string]interface{}{}
	if vu, ok := x.(ValueUpdates); ok {
		err := vu.Iterate(func(k string, v interface{}) bool {
			vmap[k] = v
			return true
		})
		return vmap, err
	} else if xs, ok := x.([]interface{}); ok {
		for i, f := range fields {
			vmap[f.Name] = xs[i]
		}
		return vmap, nil
	} else {
		return nil, fmt.Errorf("unexpected type: %T", x)
	}
}

// ValuesEqual checks if the two values are equal for the provided kind. It checks Kind.ValidateValueType first.
func (t Kind) ValuesEqual(x, y interface{}) bool {
	if t.ValidateValueType(x) != nil {
		return false
	}

	if t.ValidateValueType(y) != nil {
		return false
	}

	switch t {
	case BytesKind, JSONKind:
		return bytes.Equal(x.([]byte), y.([]byte))
	default:
		return x == y
	}
}
