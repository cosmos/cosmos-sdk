package store

// KVPair defines a key-value pair with additional metadata that is used to
// track writes. Deletion can be denoted by a nil value or explicitly by the
// Delete field.
type KVPair struct {
	Key      []byte
	Value    []byte
	StoreKey string // optional
}

// Changeset defines a set of KVPair entries.
type Changeset struct {
	Pairs []KVPair
}

func NewChangeSet(pairs ...KVPair) *Changeset {
	return &Changeset{
		Pairs: pairs,
	}
}

// Size returns the number of key-value pairs in the batch.
func (cs *Changeset) Size() int {
	return len(cs.Pairs)
}

// Add adds a key-value pair to the ChangeSet.
func (cs *Changeset) Add(key, value []byte) {
	cs.Pairs = append(cs.Pairs, KVPair{
		Key:   key,
		Value: value,
	})
}

// AddKVPair adds a KVPair to the ChangeSet.
func (cs *Changeset) AddKVPair(pair KVPair) {
	cs.Pairs = append(cs.Pairs, pair)
}
