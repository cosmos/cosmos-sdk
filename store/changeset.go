package store

// KVPair defines a key-value pair with additional metadata that is used to
// track writes. Deletion can be denoted by a nil value or explicitly by the
// Delete field.
type KVPair struct {
	Key      []byte
	Value    []byte
	StoreKey string // Optional for snapshot restore
}

type KVPairs []KVPair

// Changeset defines a set of KVPair entries by maintaining a map
// from store key to a slice of KVPair objects.
type Changeset struct {
	Pairs map[string]KVPairs
}

func NewChangeset(pairs map[string]KVPairs) *Changeset {
	return &Changeset{
		Pairs: pairs,
	}
}

// Size returns the number of key-value pairs in the batch.
func (cs *Changeset) Size() int {
	cnt := 0
	for _, pairs := range cs.Pairs {
		cnt += len(pairs)
	}

	return cnt
}

// Add adds a key-value pair to the ChangeSet.
func (cs *Changeset) Add(storeKey string, key, value []byte) {
	cs.Pairs[storeKey] = append(cs.Pairs[storeKey], KVPair{
		Key:   key,
		Value: value,
	})
}

// AddKVPair adds a KVPair to the ChangeSet.
func (cs *Changeset) AddKVPair(storeKey string, pair KVPair) {
	cs.Pairs[storeKey] = append(cs.Pairs[storeKey], pair)
}
