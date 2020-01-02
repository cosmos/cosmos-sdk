package types

// NewBitArray create new BitArray instance for elements not lower then hinted.
func NewBitArray(hint int64) BitArray {
	quo := hint / 8
	rem := hint % 8
	if rem != 0 {
		quo++
	}
	return make(BitArray, quo)
}

// BitArray can be used to check for existence of an element in an efficient way.
type BitArray []byte

func (b BitArray) position(index int64) int64 {
	return index / 8
}

func (b BitArray) bit(index int64) int64 {
	return index % 8
}

// Clear sets bit at a given position to zero. Panics if bit is out of bounds.
func (b BitArray) Clear(index int64) {
	b.Set(index, 0)
}

// Fill sets bit at a given position to one. Panics if bit is out of bounds.
func (b BitArray) Fill(index int64) {
	b.Set(index, 1)
}

// Size returns amount of bits that can fit in the bit array instance.
func (b BitArray) Size() int64 {
	return int64(len(b) * 8)
}

// Set sets bit to either one or zero. Panics if bit is out of bounds.
func (b BitArray) Set(index, val int64) {
	pos := b.position(index)
	if pos >= int64(len(b)) {
		panic("index overflow")
	}
	bit := b.bit(index)
	if val != 0 {
		b[pos] |= 1 << bit
	} else {
		b[pos] &= ^(1 << bit)
	}
}

// Get returns true if bit at a given position is set. Panics if bit is out of bounds.
func (b BitArray) Get(index int64) bool {
	pos := b.position(index)
	if pos >= int64(len(b)) {
		panic("index overlow")
	}
	return (b[pos] & (1 << b.bit(index))) > 0
}
