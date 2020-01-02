package types

// NewVoteArrayFromBytes creates VoteArray from stored bytes.
func NewVoteArrayFromBytes(value []byte) *VoteArray {
	bitarray := BitArray(value)
	return &VoteArray{
		bitarray: bitarray,
	}
}

// NewVoteArray creates VoteArray instance for number of elements not lower then size.
func NewVoteArray(size int64) *VoteArray {
	bitarray := NewBitArray(size)
	return &VoteArray{
		bitarray: bitarray,
	}
}

// VoteArray can be used to check if vote was missed in an efficient way.
type VoteArray struct {
	bitarray BitArray
}

// Get returns vote stored at index.
func (v VoteArray) Get(index int64) Vote {
	return Vote{
		index:    index,
		bitarray: v.bitarray,
	}
}

// Bytes returns byte slice that can be saved on disk.
func (v VoteArray) Bytes() []byte {
	return v.bitarray
}

// Vote can be used to check current state of the vote at a given index or update it.
type Vote struct {
	bitarray BitArray
	index    int64
}

// Voted is true if validator voted on an expected height.
func (v Vote) Voted() bool {
	return !v.Missed()
}

// Missed is true if validator didn't vote on an expected height.
func (v Vote) Missed() bool {
	return v.bitarray.Get(v.index)
}

// Miss updates vote.
func (v Vote) Miss() {
	v.bitarray.Fill(v.index)
}

// Vote updates vote.
func (v Vote) Vote() {
	v.bitarray.Clear(v.index)
}
