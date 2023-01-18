package indexes

/*
func TestUniqueIndex(t *testing.T) {
	sk, ctx := collections.deps()
	schema := collections.NewSchemaBuilder(sk)
	ui := NeqUnique(schema, collections.NewPrefix("unique_index"), "unique_index", collections.Uint64Key, collections.Uint64Key, func(_ uint64, v collections.company) (uint64, error) {
		return v.Vat, nil
	})

	// map company with id 1 to vat 1_1
	err := ui.Reference(ctx, 1, collections.company{Vat: 1_1}, nil)
	require.NoError(t, err)

	// map company with id 2 to vat 2_2
	err = ui.Reference(ctx, 2, collections.company{Vat: 2_2}, nil)
	require.NoError(t, err)

	// mapping company 3 with vat 1_1 must yield to a ErrConflict
	err = ui.Reference(ctx, 1, collections.company{Vat: 1_1}, nil)
	require.ErrorIs(t, err, collections.ErrConflict)

	// assert references are correct
	id, err := ui.ExactMatch(ctx, 1_1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), id)

	id, err = ui.ExactMatch(ctx, 2_2)
	require.NoError(t, err)
	require.Equal(t, uint64(2), id)

	// on reference updates, the new referencing key is created and the old is removed
	err = ui.Reference(ctx, 1, collections.company{Vat: 1_2}, &collections.company{Vat: 1_1})
	require.NoError(t, err)
	id, err = ui.ExactMatch(ctx, 1_2) // assert a new reference is created
	require.NoError(t, err)
	require.Equal(t, uint64(1), id)
	_, err = ui.ExactMatch(ctx, 1_1) // assert old reference was removed
	require.ErrorIs(t, err, collections.ErrNotFound)
}
*/
