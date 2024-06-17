package indexertesting

import (
	"encoding/json"
	"fmt"
	rand "math/rand/v2"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	indexerbase "cosmossdk.io/indexer/base"
)

// ListenerTestFixture is a test fixture for testing listener implementations with a pre-defined data set
// that attempts to cover all known types of objects and fields. The test data currently includes data for
// two fake modules over three blocks of data. The data set should remain relatively stable between releases
// and generally only be changed when new features are added, so it should be suitable for regression or golden tests.
type ListenerTestFixture struct {
	rndSource    rand.Source
	block        uint64
	listener     indexerbase.Listener
	allKeyModule *testModule
}

type ListenerTestFixtureOptions struct {
	EventAlignedWrites bool
}

func NewListenerTestFixture(listener indexerbase.Listener, options ListenerTestFixtureOptions) *ListenerTestFixture {
	src := rand.NewPCG(1, 2)
	return &ListenerTestFixture{
		rndSource:    src,
		listener:     listener,
		allKeyModule: mkAllKeysModule(src),
	}
}

func (f *ListenerTestFixture) Initialize() error {
	if f.listener.InitializeModuleSchema != nil {
		err := f.listener.InitializeModuleSchema(f.allKeyModule.name, f.allKeyModule.schema)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *ListenerTestFixture) NextBlock() error {
	// TODO:
	//f.block++
	//
	//if f.listener.StartBlock != nil {
	//	err := f.listener.StartBlock(f.block)
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	//err := f.allKeyModule.updater(f.rndSource, &f.listener)
	//if err != nil {
	//	return err
	//}
	//
	//if f.listener.Commit != nil {
	//	err := f.listener.Commit()
	//	if err != nil {
	//		return err
	//	}
	//}

	return nil
}

func (f *ListenerTestFixture) block3() error {
	return nil
}

var moduleSchemaA = indexerbase.ModuleSchema{
	ObjectTypes: []indexerbase.ObjectType{
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

type testModule struct {
	name   string
	schema indexerbase.ModuleSchema
	state  map[string]*testObjectStore
}

type testObjectStore struct {
	updater func(rand.Source, *indexerbase.Listener) error
	state   map[string]kvPair
}

type kvPair struct {
	key   any
	value any
	state valueState
}

type valueState int

const (
	valueStateNotInitialized valueState = iota
	valueStateSet
	valueStateDeleted
)

func mkAllKeysModule(src rand.Source) *testModule {
	mod := &testModule{
		name:  "all_keys",
		state: map[string]*testObjectStore{},
	}
	for i := 1; i < int(maxKind); i++ {
		kind := indexerbase.Kind(i)
		typ := mkTestObjectType(kind)
		mod.schema.ObjectTypes = append(mod.schema.ObjectTypes, typ)
		state := map[string]kvPair{}
		// generate 5 keys
		for j := 0; j < 5; j++ {
			key1 := mkTestValue(src, kind, false)
			key2 := mkTestValue(src, kind, true)
			key := []any{key1, key2}
			state[fmt.Sprintf("%v", key)] = kvPair{
				key: key,
			}
		}

		objStore := &testObjectStore{
			state: state,
		}
		mod.state[typ.Name] = objStore
	}

	return mod
}

func mkTestObjectType(kind indexerbase.Kind) indexerbase.ObjectType {
	field := indexerbase.Field{
		Name: fmt.Sprintf("test_%v", kind),
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

	return indexerbase.ObjectType{
		Name:        fmt.Sprintf("test_%v", kind),
		KeyFields:   []indexerbase.Field{key1Field, key2Field},
		ValueFields: []indexerbase.Field{val1Field, val2Field},
	}
}

func mkTestUpdate(rnd rand.Source, kind indexerbase.Kind) indexerbase.ObjectUpdate {
	update := indexerbase.ObjectUpdate{}

	k1 := mkTestValue(rnd, kind, false)
	k2 := mkTestValue(rnd, kind, true)
	update.Key = []any{k1, k2}

	// delete 50% of the time
	if rnd.Uint64()%2 == 0 {
		update.Delete = true
		return update
	}

	v1 := mkTestValue(rnd, kind, false)
	v2 := mkTestValue(rnd, kind, true)
	update.Value = []any{v1, v2}

	return update
}

func mkTestValue(src rand.Source, kind indexerbase.Kind, nullable bool) any {
	faker := gofakeit.NewFaker(src, false)
	// if it's nullable, return nil 50% of the time
	if nullable && faker.Bool() {
		return nil
	}

	switch kind {
	case indexerbase.StringKind:
		// TODO fmt.Stringer
		return faker.LoremIpsumSentence(faker.IntN(100))
	case indexerbase.BytesKind:
		return randBytes(src)
	case indexerbase.Int8Kind:
		return faker.Int8()
	case indexerbase.Int16Kind:
		return faker.Int16()
	case indexerbase.Uint8Kind:
		return faker.Uint16()
	case indexerbase.Uint16Kind:
		return faker.Uint16()
	case indexerbase.Int32Kind:
		return faker.Int32()
	case indexerbase.Uint32Kind:
		return faker.Uint32()
	case indexerbase.Int64Kind:
		return faker.Int64()
	case indexerbase.Uint64Kind:
		return faker.Uint64()
	case indexerbase.IntegerKind:
		x := faker.Int64()
		return fmt.Sprintf("%d", x)
	case indexerbase.DecimalKind:
		x := faker.Int64()
		y := faker.UintN(1000000)
		return fmt.Sprintf("%d.%d", x, y)
	case indexerbase.BoolKind:
		return faker.Bool()
	case indexerbase.TimeKind:
		return time.Unix(faker.Int64(), int64(faker.UintN(1000000000)))
	case indexerbase.DurationKind:
		return time.Duration(faker.Int64())
	case indexerbase.Float32Kind:
		return faker.Float32()
	case indexerbase.Float64Kind:
		return faker.Float64()
	case indexerbase.Bech32AddressKind:
		// TODO: select from some actually valid known bech32 address strings and bytes"
		return "cosmos1abcdefgh1234567890"
	case indexerbase.EnumKind:
		return faker.RandomString(testEnum.Values)
	case indexerbase.JSONKind:
		// TODO: other types
		bz, err := faker.JSON(nil)
		if err != nil {
			panic(err)
		}
		return json.RawMessage(bz)
	default:
	}
	panic(fmt.Errorf("unexpected kind: %v", kind))
}

func randBytes(src rand.Source) []byte {
	rnd := rand.New(src)
	n := rnd.IntN(1024)
	bz := make([]byte, n)
	for i := 0; i < n; i++ {
		bz[i] = byte(rnd.Uint32N(256))
	}
	return bz
}

var testEnum = indexerbase.EnumDefinition{
	Name:   "TestEnum",
	Values: []string{"A", "B", "C"},
}
