package collections

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type company struct {
	City string
	Vat  uint64
}

type companyIndexes struct {
	City *GenericMultiIndex[string, string, string, company]
	Vat  *UniqueIndex[uint64, string, company]
}

func (c companyIndexes) IndexesList() []Index[string, company] {
	return []Index[string, company]{c.City, c.Vat}
}

func TestIndexedMap(t *testing.T) {
	sk, ctx := deps()
	schema := NewSchemaBuilder(sk)

	im := NewIndexedMap(schema, NewPrefix(0), "companies", StringKey, newTestValueCodec[company](),
		companyIndexes{
			City: NewGenericMultiIndex(schema, NewPrefix(1), "companies_by_city", StringKey, StringKey, func(pk string, value company) ([]IndexReference[string, string], error) {
				return []IndexReference[string, string]{{
					Referring: value.City,
					Referred:  pk,
				}}, nil
			}),
			Vat: NewUniqueIndex(schema, NewPrefix(2), "companies_by_vat", Uint64Key, StringKey, func(_ string, v company) (uint64, error) {
				return v.Vat, nil
			}),
		},
	)

	// test insertion
	err := im.Set(ctx, "1", company{
		City: "milan",
		Vat:  0,
	})
	require.NoError(t, err)

	err = im.Set(ctx, "2", company{
		City: "milan",
		Vat:  1,
	})
	require.NoError(t, err)

	err = im.Set(ctx, "3", company{
		City: "milan",
		Vat:  4,
	})
	require.NoError(t, err)

	pk, err := im.Indexes.Vat.ExactMatch(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "2", pk)

	// test a set which updates the indexes
	err = im.Set(ctx, "2", company{
		City: "milan",
		Vat:  2,
	})
	require.NoError(t, err)

	pk, err = im.Indexes.Vat.ExactMatch(ctx, 2)
	require.NoError(t, err)
	require.Equal(t, "2", pk)

	pk, err = im.Indexes.Vat.ExactMatch(ctx, 1)
	require.ErrorIs(t, err, ErrNotFound)

	// test removal
	err = im.Remove(ctx, "2")
	require.NoError(t, err)
	_, err = im.Indexes.Vat.ExactMatch(ctx, 2)
	require.ErrorIs(t, err, ErrNotFound)

	// test iteration
	iter, err := im.Iterate(ctx, nil)
	require.NoError(t, err)
	keys, err := iter.Keys()
	require.NoError(t, err)
	require.Equal(t, []string{"1", "3"}, keys)

	// test get
	v, err := im.Get(ctx, "3")
	require.NoError(t, err)
	require.Equal(t, company{"milan", 4}, v)
}
