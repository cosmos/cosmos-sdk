package types

func NewVoteArrayFromBytes(value []byte) *VoteArray {
	bitarray := BitArray(value)
	return &VoteArray{
		bitarray: bitarray,
	}
}

func NewVoteArray(votersSize int) *VoteArray {
	bitarray := NewBitArray(votersSize)
	return &VoteArray{
		bitarray: bitarray,
	}
}

type VoteArray struct {
	bitarray BitArray
}

func (v VoteArray) Get(index int) Vote {
	return Vote{
		index:    index,
		bitarray: v.bitarray,
	}
}

func (v VoteArray) Bytes() []byte {
	return v.bitarray
}

type Vote struct {
	bitarray BitArray
	index    int
}

func (v Vote) Voted() bool {
	return !v.Missed()
}

func (v Vote) Missed() bool {
	return v.bitarray.Get(v.index)
}

func (v Vote) Miss() {
	v.bitarray.Fill(v.index)
}

func (v Vote) Vote() {
	v.bitarray.Clear(v.index)
}
