package schematesting

import (
	"testing"

	"cosmossdk.io/schema"
)

func TestObjectKeyString(t *testing.T) {
	tt := []struct {
		objectType schema.StateObjectType
		key        any
		expected   string
	}{
		{
			objectType: schema.StateObjectType{
				Name: "Singleton",
				ValueFields: []schema.Field{
					{Name: "Value", Kind: schema.StringKind},
				},
			},
			key:      nil,
			expected: "",
		},
		{
			objectType: schema.StateObjectType{
				Name:      "Simple",
				KeyFields: []schema.Field{{Name: "Key", Kind: schema.StringKind}},
			},
			key:      "key",
			expected: "Key=key",
		},
		{
			objectType: schema.StateObjectType{
				Name: "BytesAddressDecInt",
				KeyFields: []schema.Field{
					{Name: "Bz", Kind: schema.BytesKind},
					{Name: "Addr", Kind: schema.AddressKind},
					{Name: "Dec", Kind: schema.DecimalKind},
					{Name: "Int", Kind: schema.IntegerKind},
				},
			},
			key: []interface{}{
				[]byte{0x01, 0x02},
				[]byte{0x03, 0x04},
				"123.4560000",               // trailing zeros should get removed
				"0000012345678900000000000", // leading zeros should get removed and this should be in exponential form
			},
			expected: "Bz=0x0102, Addr=0x0304, Dec=123.456, Int=1.23456789E+19",
		},
	}

	for _, tc := range tt {
		t.Run(tc.objectType.Name, func(t *testing.T) {
			actual := ObjectKeyString(tc.objectType, tc.key)
			if actual != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}
