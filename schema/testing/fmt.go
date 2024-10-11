package schematesting

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"

	"cosmossdk.io/schema"
)

// ObjectKeyString formats the object key as a string deterministically for storage in a map.
// The key must be valid for the object type and the object type must be valid.
// No validation is performed here.
func ObjectKeyString(objectType schema.StateObjectType, key any) string {
	keyFields := objectType.KeyFields
	n := len(keyFields)
	switch n {
	case 0:
		return ""
	case 1:
		valStr := fmtValue(keyFields[0].Kind, key)
		return fmt.Sprintf("%s=%v", keyFields[0].Name, valStr)
	default:
		ks := key.([]interface{})
		res := ""
		for i := 0; i < n; i++ {
			if i != 0 {
				res += ", "
			}
			valStr := fmtValue(keyFields[i].Kind, ks[i])
			res += fmt.Sprintf("%s=%v", keyFields[i].Name, valStr)
		}
		return res
	}
}

func fmtValue(kind schema.Kind, value any) string {
	switch kind {
	case schema.BytesKind, schema.AddressKind:
		return fmt.Sprintf("0x%x", value)
	case schema.DecimalKind, schema.IntegerKind:
		// we need to normalize decimal & integer strings to remove leading & trailing zeros
		d, _, err := apd.NewFromString(value.(string))
		if err != nil {
			panic(err)
		}
		r := &apd.Decimal{}
		r, _ = r.Reduce(d)
		return r.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}
