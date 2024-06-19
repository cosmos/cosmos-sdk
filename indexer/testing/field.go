package indexertesting

import (
	"encoding/json"
	"fmt"
	"time"

	indexerbase "cosmossdk.io/indexer/base"
)

func (f *BaseFixture) ValueForField(field indexerbase.Field) interface{} {
	// if it's nullable, return nil 50% of the time
	if field.Nullable && f.faker.Bool() {
		return nil
	}

	switch field.Kind {
	case indexerbase.StringKind:
		return f.faker.LoremIpsumSentence(f.faker.IntN(100))
	case indexerbase.BytesKind:
		return f.randBytes()
	case indexerbase.Int8Kind:
		return f.faker.Int8()
	case indexerbase.Int16Kind:
		return f.faker.Int16()
	case indexerbase.Uint8Kind:
		return f.faker.Uint16()
	case indexerbase.Uint16Kind:
		return f.faker.Uint16()
	case indexerbase.Int32Kind:
		return f.faker.Int32()
	case indexerbase.Uint32Kind:
		return f.faker.Uint32()
	case indexerbase.Int64Kind:
		return f.faker.Int64()
	case indexerbase.Uint64Kind:
		return f.faker.Uint64()
	case indexerbase.IntegerKind:
		x := f.faker.Int64()
		return fmt.Sprintf("%d", x)
	case indexerbase.DecimalKind:
		x := f.faker.Int64()
		y := f.faker.UintN(1000000)
		return fmt.Sprintf("%d.%d", x, y)
	case indexerbase.BoolKind:
		return f.faker.Bool()
	case indexerbase.TimeKind:
		return time.Unix(f.faker.Int64(), int64(f.faker.UintN(1000000000)))
	case indexerbase.DurationKind:
		return time.Duration(f.faker.Int64())
	case indexerbase.Float32Kind:
		return f.faker.Float32()
	case indexerbase.Float64Kind:
		return f.faker.Float64()
	case indexerbase.Bech32AddressKind:
		// TODO: select from some actually valid known bech32 address strings and bytes"
		return "cosmos1abcdefgh1234567890"
	case indexerbase.EnumKind:
		return f.faker.RandomString(testEnum.Values)
	case indexerbase.JSONKind:
		// TODO: other types
		bz, err := f.faker.JSON(nil)
		if err != nil {
			panic(err)
		}
		return json.RawMessage(bz)
	default:
	}
	panic(fmt.Errorf("unexpected kind: %v", field.Kind))
}

func (f *BaseFixture) randBytes() []byte {
	n := f.rnd.IntN(1024)
	bz := make([]byte, n)
	for i := 0; i < n; i++ {
		bz[i] = byte(f.rnd.Uint32N(256))
	}
	return bz
}

func (f *BaseFixture) ValueForKeyFields(keyFields []indexerbase.Field) interface{} {
	if len(keyFields) == 0 {
		return nil
	}

	if len(keyFields) == 1 {
		return f.ValueForField(keyFields[0])
	}

	values := make([]interface{}, len(keyFields))
	for i := range keyFields {
		values[i] = f.ValueForField(keyFields[i])
	}

	return values
}

func (f *BaseFixture) ValueForValueField(valueFields []indexerbase.Field) interface{} {
	if len(valueFields) == 0 {
		return nil
	}

	if len(valueFields) == 1 {
		return f.ValueForField(valueFields[0])
	}

	// return ValueUpdates 50% of the time
	if f.faker.Bool() {
		valueUpdates := map[string]interface{}{}
		for _, field := range valueFields {
			// exclude a field 50% of the time
			if f.faker.Bool() {
				continue
			}
			valueUpdates[field.Name] = f.ValueForField(field)
		}
		return indexerbase.MapValueUpdates(valueUpdates)
	} else {
		values := make([]interface{}, len(valueFields))
		for i := range valueFields {
			values[i] = f.ValueForField(valueFields[i])
		}

		return values
	}
}
