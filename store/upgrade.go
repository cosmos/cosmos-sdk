package store

import "golang.org/x/exp/slices"

// StoreUpgrades defines a series of transformations to apply the RootStore upon
// loading a version.
type StoreUpgrades struct {
	Add    []string      `json:"add"`
	Rename []StoreRename `json:"rename"`
	Delete []string      `json:"delete"`
}

// StoreRename defines a change in a store key. All data previously stored under
// the current store key should be migrated to the new store key, while also
// deleting the old store key.
type StoreRename struct {
	OldKey string `json:"old_key"`
	NewKey string `json:"new_key"`
}

// IsAdded returns true if the given key should be added.
func (s *StoreUpgrades) IsAdded(key string) bool {
	if s == nil {
		return false
	}

	return slices.Contains(s.Add, key)
}

// IsDeleted returns true if the given key should be deleted.
func (s *StoreUpgrades) IsDeleted(key string) bool {
	if s == nil {
		return false
	}

	return slices.Contains(s.Delete, key)
}

// RenamedFrom returns the oldKey if it was renamed. It returns an empty string
// if it was not renamed.
func (s *StoreUpgrades) RenamedFrom(key string) string {
	if s == nil {
		return ""
	}

	for _, re := range s.Rename {
		if re.NewKey == key {
			return re.OldKey
		}
	}

	return ""
}
