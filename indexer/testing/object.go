package indexertesting

import indexerbase "cosmossdk.io/indexer/base"

func (f *BaseFixture) RandObjectUpdate(objectType indexerbase.ObjectType) indexerbase.ObjectUpdate {
	update := indexerbase.ObjectUpdate{
		TypeName: objectType.Name,
	}

	update.Key = f.ValueForKeyFields(objectType.KeyFields)

	// delete 50% of the time
	if f.faker.Bool() {
		update.Delete = true
	} else {
		update.Value = f.ValueForValueField(objectType.ValueFields)
	}

	return update
}

// TODO: object update for key set
