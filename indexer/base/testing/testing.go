package indexertesting

import (
	"fmt"
	"math/rand"
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

func mkTestModule() (indexerbase.ModuleSchema, func(*rand.Rand) []indexerbase.EntityUpdate) {
	schema := indexerbase.ModuleSchema{}
	for i := 1; i < int(maxKind); i++ {
		schema.Tables = append(schema.Tables, mkTestTable(indexerbase.Kind(i)))
	}

	return schema, func(rnd *rand.Rand) []indexerbase.EntityUpdate {
		var updates []indexerbase.EntityUpdate
		for i := 1; i < int(maxKind); i++ {
			// 0-10 updates per kind
			n := int(rnd.Int31n(11))
			for j := 0; j < n; j++ {
				updates = append(updates, mkTestUpdate(rnd, indexerbase.Kind(i)))
			}
		}
		return updates
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

func mkTestUpdate(rnd *rand.Rand, kind indexerbase.Kind) indexerbase.EntityUpdate {
	update := indexerbase.EntityUpdate{}

	k1 := mkTestValue(rnd, kind, false)
	k2 := mkTestValue(rnd, kind, true)
	update.Key = []any{k1, k2}

	// delete 10% of the time
	if rnd.Int31n(10) == 1 {
		update.Delete = true
		return update
	}

	v1 := mkTestValue(rnd, kind, false)
	v2 := mkTestValue(rnd, kind, true)
	update.Value = []any{v1, v2}

	return update
}

func mkTestValue(rnd *rand.Rand, kind indexerbase.Kind, nullable bool) any {
	// if it's nullable, return nil 10% of the time
	if nullable && rnd.Int31n(10) == 1 {
		return nil
	}

	switch kind {
	case indexerbase.StringKind:
		// TODO fmt.Stringer
		return string(randBz(rnd))
	case indexerbase.BytesKind:
		return randBz(rnd)
	case indexerbase.Int8Kind:
		return int8(rnd.Int31n(256) - 128)
	case indexerbase.Int16Kind:
		return int16(rnd.Int31n(65536) - 32768)
	case indexerbase.Uint8Kind:
		return uint8(rnd.Int31n(256))
	case indexerbase.Uint16Kind:
		return uint16(rnd.Int31n(65536))
	case indexerbase.Int32Kind:
		return int32(rnd.Int63n(4294967296) - 2147483648)
	case indexerbase.Uint32Kind:
		return uint32(rnd.Int63n(4294967296))
	case indexerbase.Int64Kind:
		return rnd.Int63()
	case indexerbase.Uint64Kind:
		return rnd.Uint64()
	case indexerbase.IntegerKind:
		x := rnd.Int63()
		return fmt.Sprintf("%d", x)
	case indexerbase.DecimalKind:
		x := rnd.Int63()
		y := rnd.Int63n(1000000000)
		return fmt.Sprintf("%d.%d", x, y)
	case indexerbase.BoolKind:
		return rnd.Int31n(2) == 1
	case indexerbase.TimeKind:
		return time.Unix(rnd.Int63(), rnd.Int63n(1000000000))
	case indexerbase.DurationKind:
		return time.Duration(rnd.Int63())
	case indexerbase.Float32Kind:
		return float32(rnd.Float64())
	case indexerbase.Float64Kind:
		return rnd.Float64()
	case indexerbase.Bech32AddressKind:
		panic("TODO: select from some actually valid known bech32 address strings and bytes")
	case indexerbase.EnumKind:
		return testEnum.Values[rnd.Int31n(int32(len(testEnum.Values)))]
	case indexerbase.JSONKind:
		//// TODO other types
		//return json.RawMessage(`{"seed": ` + strconv.FormatUint(seed, 10) + `}`)
		panic("TODO")
	default:
	}
	panic(fmt.Errorf("unexpected kind: %v", kind))
}

func randBz(rnd *rand.Rand) []byte {
	n := rnd.Int31n(1024)
	bz := make([]byte, n)
	_, err := rnd.Read(bz)
	if err != nil {
		panic(err)
	}
	return bz
}

var testEnum = indexerbase.EnumDefinition{
	Name:   "TestEnum",
	Values: []string{"A", "B", "C"},
}
