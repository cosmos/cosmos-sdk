package internal

import "fmt"

// Uint40 is a 40-bit unsigned integer stored in 5 bytes with little-endian encoding.
type Uint40 [5]byte

const MaxUint40 = 1<<40 - 1

// NewUint40 creates a new Uint40 from a uint64.
func NewUint40(v uint64) Uint40 {
	if v>>40 != 0 {
		panic(fmt.Sprintf("value %d overflows Uint40", v))
	}
	var u Uint40
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	u[2] = byte(v >> 16)
	u[3] = byte(v >> 24)
	u[4] = byte(v >> 32)
	return u
}

func (u Uint40) IsZero() bool {
	return u[0] == 0 && u[1] == 0 && u[2] == 0 && u[3] == 0 && u[4] == 0
}

// ToUint64 converts the Uint40 to a uint64.
func (u Uint40) ToUint64() uint64 {
	return uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 | uint64(u[4])<<32
}

// String implements fmt.Stringer.
func (u Uint40) String() string {
	return fmt.Sprintf("%d", u.ToUint64())
}
