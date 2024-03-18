package store

// StoreUpgrades defines a series of transformations to apply the multistore db upon load
type StoreUpgrades struct {
	Added   []string      `json:"added"`
	Renamed []StoreRename `json:"renamed"`
	Deleted []string      `json:"deleted"`
}

// StoreRename defines a name change of a sub-store.
// All data previously under a PrefixStore with OldKey will be copied
// to a PrefixStore with NewKey, then deleted from OldKey store.
type StoreRename struct {
	OldKey string `json:"old_key"`
	NewKey string `json:"new_key"`
}

// IsAdded returns true if the given key should be added
func (s *StoreUpgrades) IsAdded(key string) bool {
	if s == nil {
		return false
	}
	for _, added := range s.Added {
		if key == added {
			return true
		}
	}
	return false
}

// IsDeleted returns true if the given key should be deleted
func (s *StoreUpgrades) IsDeleted(key string) bool {
	if s == nil {
		return false
	}
	for _, d := range s.Deleted {
		if d == key {
			return true
		}
	}
	return false
}

// RenamedFrom returns the oldKey if it was renamed
// Returns "" if it was not renamed
func (s *StoreUpgrades) RenamedFrom(key string) string {
	if s == nil {
		return ""
	}
	for _, re := range s.Renamed {
		if re.NewKey == key {
			return re.OldKey
		}
	}
	return ""
}
