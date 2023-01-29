package collections_test

import (
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"
	"github.com/stretchr/testify/require"
)

type company struct {
	City string
	Vat  uint64
}

type companyIndexes struct {
	// City is an index of the company indexed map. It indexes a company
	// given its city. The index is multi, meaning that there can be multiple
	// companies from the same city.
	City *collections.GenericMultiIndex[string, string, string, company]
	// Vat is an index of the company indexed map. It indexes a company
	// given its VAT number. The index is unique, meaning that there can be
	// only one VAT number for a company.
	Vat *collections.GenericUniqueIndex[uint64, string, string, company]
}

func (c companyIndexes) IndexesList() []collections.Index[string, company] {
	return []collections.Index[string, company]{c.City, c.Vat}
}

func newTestIndexedMap(schema *collections.SchemaBuilder) *collections.IndexedMap[string, company, companyIndexes] {
	return collections.NewIndexedMap(schema, collections.NewPrefix(0), "companies", collections.StringKey, colltest.MockValueCodec[company](),
		companyIndexes{
			City: collections.NewGenericMultiIndex(schema, collections.NewPrefix(1), "companies_by_city", collections.StringKey, collections.StringKey, func(pk string, value company) ([]collections.IndexReference[string, string], error) {
				return []collections.IndexReference[string, string]{collections.NewIndexReference(value.City, pk)}, nil
			}),
			Vat: collections.NewGenericUniqueIndex(schema, collections.NewPrefix(2), "companies_by_vat", collections.Uint64Key, collections.StringKey, func(pk string, v company) ([]collections.IndexReference[uint64, string], error) {
				return []collections.IndexReference[uint64, string]{collections.NewIndexReference(v.Vat, pk)}, nil
			}),
		},
	)
}

func TestIndexedMap(t *testing.T) {
	sk, ctx := colltest.MockStore()
	schema := collections.NewSchemaBuilder(sk)

	im := newTestIndexedMap(schema)

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

	pk, err := im.Indexes.Vat.Get(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "2", pk)

	// test a set which updates the indexes
	err = im.Set(ctx, "2", company{
		City: "milan",
		Vat:  2,
	})
	require.NoError(t, err)

	pk, err = im.Indexes.Vat.Get(ctx, 2)
	require.NoError(t, err)
	require.Equal(t, "2", pk)

	_, err = im.Indexes.Vat.Get(ctx, 1)
	require.ErrorIs(t, err, collections.ErrNotFound)

	// test removal
	err = im.Remove(ctx, "2")
	require.NoError(t, err)
	_, err = im.Indexes.Vat.Get(ctx, 2)
	require.ErrorIs(t, err, collections.ErrNotFound)

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
