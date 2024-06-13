package indexertesting

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	indexerbase "cosmossdk.io/indexer/base"
)

// ListenerTestFixture is a test fixture for testing listener implementations with a pre-defined data set
// that attempts to cover all known types of tables and fields. The test data currently includes data for
// two fake modules over three blocks of data. The data set should remain relatively stable between releases
// and generally only be changed when new features are added, so it should be suitable for regression or golden tests.
type ListenerTestFixture struct {
	listener indexerbase.Listener
}

type ListenerTestFixtureOptions struct {
	EventAlignedWrites bool
}

func NewListenerTestFixture(listener indexerbase.Listener, options ListenerTestFixtureOptions) *ListenerTestFixture {
	return &ListenerTestFixture{
		listener: listener,
	}
}

func (f *ListenerTestFixture) Initialize() error {
	return nil
}

func (f *ListenerTestFixture) NextBlock() (bool, error) {
	return false, nil
}

func (f *ListenerTestFixture) block1() error {
	return nil
}

func (f *ListenerTestFixture) block2() error {
	return nil
}

func (f *ListenerTestFixture) block3() error {
	return nil
}

var moduleSchemaA = indexerbase.ModuleSchema{
	Tables: []indexerbase.Table{
		{
			"Singleton",
			[]indexerbase.Field{},
			[]indexerbase.Field{
				{
					Name: "Value",
					Kind: indexerbase.StringKind,
				},
			},
			false,
		},
		{
			Name: "Simple",
			KeyFields: []indexerbase.Field{
				{
					Name: "Key",
					Kind: indexerbase.StringKind,
				},
			},
			ValueFields: []indexerbase.Field{
				{
					Name: "Value1",
					Kind: indexerbase.Int32Kind,
				},
				{
					Name: "Value2",
					Kind: indexerbase.BytesKind,
				},
			},
		},
		{
			Name: "Two Keys",
			KeyFields: []indexerbase.Field{
				{
					Name: "Key1",
					Kind: indexerbase.StringKind,
				},
				{
					Name: "Key2",
					Kind: indexerbase.Int32Kind,
				},
			},
		},
		{
			Name: "Main Values",
		},
		{
			Name: "No Values",
		},
	},
}

var maxKind = indexerbase.JSONKind

func mkTestModule() (indexerbase.ModuleSchema, func(seed int) []indexerbase.EntityUpdate) {
	schema := indexerbase.ModuleSchema{}
	for i := 1; i < int(maxKind); i++ {
		schema.Tables = append(schema.Tables, mkTestTable(indexerbase.Kind(i)))
	}

	return schema, func(seed int) []indexerbase.EntityUpdate {
		panic("TODO")
	}
}

func mkTestTable(kind indexerbase.Kind) indexerbase.Table {
	field := indexerbase.Field{
		Name: fmt.Sprintf("test_%s", kind),
		Kind: kind,
	}

	if kind == indexerbase.EnumKind {
		field.EnumDefinition = testEnum
	}

	if kind == indexerbase.Bech32AddressKind {
		field.AddressPrefix = "cosmos"
	}

	key1Field := field
	key1Field.Name = "keyNotNull"
	key2Field := field
	key2Field.Name = "keyNullable"
	key2Field.Nullable = true
	val1Field := field
	val1Field.Name = "valNotNull"
	val2Field := field
	val2Field.Name = "valNullable"
	val2Field.Nullable = true

	return indexerbase.Table{
		Name:        "KindTable",
		KeyFields:   []indexerbase.Field{key1Field, key2Field},
		ValueFields: []indexerbase.Field{val1Field, val2Field},
	}
}

func mkTestUpdate(seed uint64, kind indexerbase.Kind) indexerbase.EntityUpdate {
	update := indexerbase.EntityUpdate{}

	k1 := mkTestValue(seed, kind, false)
	k2 := mkTestValue(seed+1, kind, true)
	update.Key = []any{k1, k2}

	// delete 10% of the time
	if seed%10 == 0 {
		update.Delete = true
		return update
	}

	v1 := mkTestValue(seed+2, kind, false)
	v2 := mkTestValue(seed+3, kind, true)
	update.Value = []any{v1, v2}

	return update
}

func mkTestValue(seed uint64, kind indexerbase.Kind, nullable bool) any {
	// if it's nullable, return nil 10% of the time
	if nullable && seed%10 == 1 {
		return nil
	}

	switch kind {
	case indexerbase.StringKind:
		// TODO fmt.Stringer
		return "seed" + strconv.FormatUint(seed, 10)
	case indexerbase.BytesKind:
		return []byte("seed" + strconv.FormatUint(seed, 10))
	case indexerbase.Int8Kind:
		return int8(seed)
	case indexerbase.Int16Kind:
		return int16(seed)
	case indexerbase.Uint8Kind:
		return uint8(seed)
	case indexerbase.Uint16Kind:
		return uint16(seed)
	case indexerbase.Int32Kind:
		return int32(seed)
	case indexerbase.Uint32Kind:
		return uint32(seed)
	case indexerbase.Int64Kind:
		return int64(seed)
	case indexerbase.Uint64Kind:
		return uint64(seed)
	case indexerbase.IntegerKind:
		// TODO fmt.Stringer, int64
		return fmt.Sprintf("%d", seed)
	case indexerbase.DecimalKind:
		// TODO fmt.Stringer
		return fmt.Sprintf("%d.%d", seed, seed)
	case indexerbase.BoolKind:
		return seed%2 == 0
	case indexerbase.TimeKind:
		return time.Unix(int64(seed), 0)
	case indexerbase.DurationKind:
		return time.Duration(seed) * time.Second
	case indexerbase.Float32Kind:
		return float32(seed)
	case indexerbase.Float64Kind:
		return float64(seed)
	case indexerbase.Bech32AddressKind:
		// TODO bytes
		return "cosmos1address" + strconv.FormatUint(seed, 10)
	case indexerbase.EnumKind:
		return testEnum.Values[int(seed)%len(testEnum.Values)]
	case indexerbase.JSONKind:
		// TODO other types
		return json.RawMessage(`{"seed": ` + strconv.FormatUint(seed, 10) + `}`)
	default:
	}
	panic(fmt.Errorf("unexpected kind: %v", kind))
}

var testEnum = indexerbase.EnumDefinition{
	Name:   "TestEnum",
	Values: []string{"A", "B", "C"},
}
