package indexerrapid

import (
	"fmt"

	"pgregory.net/rapid"

	indexerbase "cosmossdk.io/indexer/base"
)

var (
	nameGen = rapid.String()
	kindGen = rapid.Map(rapid.IntRange(int(indexerbase.InvalidKind+1), int(indexerbase.MAX_VALID_KIND-1)),
		func(i int) indexerbase.Kind {
			return indexerbase.Kind(i)
		})
	boolGen = rapid.Bool()
)

var Field = rapid.Custom(func(t *rapid.T) indexerbase.Field {
	kind := kindGen.Draw(t, "kind")
	field := indexerbase.Field{
		Name:     nameGen.Draw(t, "name"),
		Kind:     kind,
		Nullable: boolGen.Draw(t, "nullable"),
	}

	switch kind {
	case indexerbase.EnumKind:
		field.EnumDefinition = EnumDefinition.Draw(t, "enumDefinition")
	case indexerbase.Bech32AddressKind:
		field.AddressPrefix = nameGen.Draw(t, "addressPrefix")
	default:
	}

	return field
})

func FieldValue(field indexerbase.Field) *rapid.Generator[any] {
	gen := baseFieldValue(field)

	if field.Nullable {
		return rapid.OneOf(gen, rapid.Just[any](nil)).AsAny()
	}

	return gen
}

func baseFieldValue(field indexerbase.Field) *rapid.Generator[any] {
	switch field.Kind {
	case indexerbase.StringKind:
		return rapid.String().AsAny()
	case indexerbase.BytesKind:
		return rapid.SliceOf(rapid.Byte()).AsAny()
	case indexerbase.Int8Kind:
		return rapid.Int8().AsAny()
	case indexerbase.Int16Kind:
		return rapid.Int16().AsAny()
	case indexerbase.Uint8Kind:
		return rapid.Uint8().AsAny()
	case indexerbase.Uint16Kind:
		return rapid.Uint16().AsAny()
	case indexerbase.Int32Kind:
		return rapid.Int32().AsAny()
	case indexerbase.Uint32Kind:
		return rapid.Uint32().AsAny()
	case indexerbase.Int64Kind:
		return rapid.Int64().AsAny()
	case indexerbase.Uint64Kind:
		return rapid.Uint64().AsAny()
	case indexerbase.Float32Kind:
		return rapid.Float32().AsAny()
	case indexerbase.Float64Kind:
		return rapid.Float64().AsAny()
	case indexerbase.IntegerKind:
		panic("TODO")
	case indexerbase.DecimalKind:
		panic("TODO")
	case indexerbase.BoolKind:
		return rapid.Bool().AsAny()
	case indexerbase.TimeKind:
		panic("TODO")
	case indexerbase.DurationKind:
		panic("TODO")
	case indexerbase.Bech32AddressKind:
		return rapid.SliceOfN(rapid.Byte(), 20, 64).AsAny()
	case indexerbase.EnumKind:
		gen := rapid.IntRange(0, len(field.EnumDefinition.Values)-1)
		return rapid.Map(gen, func(i int) any {
			return field.EnumDefinition.Values[i]
		})
	default:
		panic(fmt.Errorf("unexpected kind: %v", field.Kind))
	}
}

func KeyFieldsValue(keyFields []indexerbase.Field) *rapid.Generator[any] {
	gens := make([]*rapid.Generator[any], len(keyFields))
	for i, field := range keyFields {
		gens[i] = FieldValue(field)
	}
	return rapid.Custom(func(t *rapid.T) any {
		values := make([]any, len(keyFields))
		for i, gen := range gens {
			values[i] = gen.Draw(t, keyFields[i].Name)
		}
		return values
	})
}

func ValueFieldsValue(valueFields []indexerbase.Field) *rapid.Generator[any] {
	gens := make([]*rapid.Generator[any], len(valueFields))
	for i, field := range valueFields {
		gens[i] = FieldValue(field)
	}
	return rapid.Custom(func(t *rapid.T) any {
		// return ValueUpdates 50% of the time
		if boolGen.Draw(t, "valueUpdates") {
			updates := map[string]any{}

			for i, gen := range gens {
				// skip 50% of the time
				if boolGen.Draw(t, fmt.Sprintf("skip_%s", valueFields[i].Name)) {
					continue
				}
				updates[valueFields[i].Name] = gen.Draw(t, valueFields[i].Name)
			}

			return indexerbase.MapValueUpdates(updates)
		} else {

			values := make([]any, len(valueFields))
			for i, gen := range gens {
				values[i] = gen.Draw(t, valueFields[i].Name)
			}

			return values
		}
	})
}
