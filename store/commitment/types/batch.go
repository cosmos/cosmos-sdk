package types

// KVPair is a key-value pair. It is used to represent a change in a Batch.
// NOTE: The Value can be nil, which means the key is deleted.
type KVPair struct {
	Key   []byte
	Value []byte
}

// Batch is a change set that can be committed to the database atomically.
type Batch struct {
	Pairs []KVPair
}

// NewBatch creates a new Batch instance.
func NewBatch() *Batch {
	return &Batch{
		Pairs: []KVPair{},
	}
}

// Size returns the number of key-value pairs in the batch.
func (b *Batch) Size() int {
	return len(b.Pairs)
}

// Add adds a key-value pair to the batch.
func (b *Batch) Add(key, value []byte) {
	b.Pairs = append(b.Pairs, KVPair{
		Key:   key,
		Value: value,
	})
}
