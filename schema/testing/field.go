package schematesting

import (
	"fmt"
	"time"

	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

var (
	kindGen = rapid.Map(rapid.IntRange(int(schema.InvalidKind+1), int(schema.MAX_VALID_KIND-1)),
		func(i int) schema.Kind {
			return schema.Kind(i)
		})
	boolGen = rapid.Bool()
)

// FieldGen generates random Field's based on the validity criteria of fields.
var FieldGen = rapid.Custom(func(t *rapid.T) schema.Field {
	kind := kindGen.Draw(t, "kind")
	field := schema.Field{
		Name:     NameGen.Draw(t, "name"),
		Kind:     kind,
		Nullable: boolGen.Draw(t, "nullable"),
	}

	switch kind {
	case schema.EnumKind:
		field.EnumType = EnumType.Draw(t, "enumDefinition")
	default:
	}

	return field
})

// FieldValueGen generates random valid values for the field, aiming to exercise the full range of possible
// values for the field.
func FieldValueGen(field schema.Field) *rapid.Generator[any] {
	gen := baseFieldValue(field)

	if field.Nullable {
		return rapid.OneOf(gen, rapid.Just[any](nil)).AsAny()
	}

	return gen
}

func baseFieldValue(field schema.Field) *rapid.Generator[any] {
	switch field.Kind {
	case schema.StringKind:
		return rapid.StringOf(rapid.Rune().Filter(func(r rune) bool {
			return r != 0 // filter out NULL characters
		})).AsAny()
	case schema.BytesKind:
		return rapid.SliceOf(rapid.Byte()).AsAny()
	case schema.Int8Kind:
		return rapid.Int8().AsAny()
	case schema.Int16Kind:
		return rapid.Int16().AsAny()
	case schema.Uint8Kind:
		return rapid.Uint8().AsAny()
	case schema.Uint16Kind:
		return rapid.Uint16().AsAny()
	case schema.Int32Kind:
		return rapid.Int32().AsAny()
	case schema.Uint32Kind:
		return rapid.Uint32().AsAny()
	case schema.Int64Kind:
		return rapid.Int64().AsAny()
	case schema.Uint64Kind:
		return rapid.Uint64().AsAny()
	case schema.Float32Kind:
		return rapid.Float32().AsAny()
	case schema.Float64Kind:
		return rapid.Float64().AsAny()
	case schema.IntegerStringKind:
		return rapid.StringMatching(schema.IntegerFormat).AsAny()
	case schema.DecimalStringKind:
		return rapid.StringMatching(schema.DecimalFormat).AsAny()
	case schema.BoolKind:
		return rapid.Bool().AsAny()
	case schema.TimeKind:
		return rapid.Map(rapid.Int64(), func(i int64) time.Time {
			return time.Unix(0, i)
		}).AsAny()
	case schema.DurationKind:
		return rapid.Map(rapid.Int64(), func(i int64) time.Duration {
			return time.Duration(i)
		}).AsAny()
	case schema.AddressKind:
		return rapid.SliceOfN(rapid.Byte(), 20, 64).AsAny()
	case schema.EnumKind:
		return rapid.SampledFrom(field.EnumType.Values).AsAny()
	default:
		panic(fmt.Errorf("unexpected kind: %v", field.Kind))
	}
}

// ObjectKeyGen generates a value that is valid for the provided object key fields.
func ObjectKeyGen(keyFields []schema.Field) *rapid.Generator[any] {
	if len(keyFields) == 0 {
		return rapid.Just[any](nil)
	}

	if len(keyFields) == 1 {
		return FieldValueGen(keyFields[0])
	}

	gens := make([]*rapid.Generator[any], len(keyFields))
	for i, field := range keyFields {
		gens[i] = FieldValueGen(field)
	}

	return rapid.Custom(func(t *rapid.T) any {
		values := make([]any, len(keyFields))
		for i, gen := range gens {
			values[i] = gen.Draw(t, keyFields[i].Name)
		}
		return values
	})
}

// ObjectValueGen generates a value that is valid for the provided object value fields. The
// forUpdate parameter indicates whether the generator should generate value that
// are valid for insertion (in the case forUpdate is false) or for update (in the case forUpdate is true).
// Values that are for update may skip some fields in a ValueUpdates instance whereas values for insertion
// will always contain all values.
func ObjectValueGen(valueFields []schema.Field, forUpdate bool) *rapid.Generator[any] {
	// special case where there are no value fields
	// we shouldn't end up here, but just in case
	if len(valueFields) == 0 {
		return rapid.Just[any](nil)
	}

	gens := make([]*rapid.Generator[any], len(valueFields))
	for i, field := range valueFields {
		gens[i] = FieldValueGen(field)
	}
	return rapid.Custom(func(t *rapid.T) any {
		// return ValueUpdates 50% of the time
		if boolGen.Draw(t, "valueUpdates") {
			updates := map[string]any{}

			n := len(valueFields)
			for i, gen := range gens {
				lastField := i == n-1
				haveUpdates := len(updates) > 0
				// skip 50% of the time if this is an update
				// but check if we have updates by the time we reach the last field
				// so we don't have an empty update
				if forUpdate &&
					(!lastField || haveUpdates) &&
					boolGen.Draw(t, fmt.Sprintf("skip_%s", valueFields[i].Name)) {
					continue
				}
				updates[valueFields[i].Name] = gen.Draw(t, valueFields[i].Name)
			}

			return schema.MapValueUpdates(updates)
		} else {
			if len(valueFields) == 1 {
				return gens[0].Draw(t, valueFields[0].Name)
			}

			values := make([]any, len(valueFields))
			for i, gen := range gens {
				values[i] = gen.Draw(t, valueFields[i].Name)
			}

			return values
		}
	})
}
