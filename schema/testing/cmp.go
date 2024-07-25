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
