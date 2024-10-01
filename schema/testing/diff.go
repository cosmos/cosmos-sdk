package schematesting

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/cockroachdb/apd/v3"

	"cosmossdk.io/schema"
)

// DiffObjectKeys compares the values as object keys for the provided field and returns a diff if they
// differ or an empty string if they are equal.
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

// DiffObjectValues compares the values as object values for the provided field and returns a diff if they
// differ or an empty string if they are equal. Object values cannot be ValueUpdates for this comparison.
func DiffObjectValues(fields []schema.Field, expected, actual any) string {
	if len(fields) == 0 {
		return ""
	}

	_, ok := expected.(schema.ValueUpdates)
	_, ok2 := expected.(schema.ValueUpdates)

	if ok || ok2 {
		return "ValueUpdates is not expected when comparing state"
	}

	return DiffObjectKeys(fields, expected, actual)
}

// DiffFieldValues compares the values for the provided field and returns a diff if they differ or an empty
// string if they are equal.
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

// CompareKindValues compares the expected and actual values for the provided kind and returns true if they are equal,
// false if they are not, and an error if the types are not valid for the kind.
// For IntegerKind and DecimalKind values, comparisons are made based on equality of the underlying numeric
// values rather than their string encoding.
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
	case schema.IntegerKind:
		expectedInt := big.NewInt(0)
		expectedInt, ok := expectedInt.SetString(expected.(string), 10)
		if !ok {
			return false, fmt.Errorf("could not convert %v to big.Int", expected)
		}

		actualInt := big.NewInt(0)
		actualInt, ok = actualInt.SetString(actual.(string), 10)
		if !ok {
			return false, fmt.Errorf("could not convert %v to big.Int", actual)
		}

		if expectedInt.Cmp(actualInt) != 0 {
			return false, nil
		}
	case schema.DecimalKind:
		expectedDec, _, err := apd.NewFromString(expected.(string))
		if err != nil {
			return false, fmt.Errorf("could not decode %v as a decimal: %w", expected, err)
		}

		actualDec, _, err := apd.NewFromString(actual.(string))
		if err != nil {
			return false, fmt.Errorf("could not decode %v as a decimal: %w", actual, err)
		}

		if expectedDec.Cmp(actualDec) != 0 {
			return false, nil
		}
	default:
		if expected != actual {
			return false, nil
		}
	}
	return true, nil
}
