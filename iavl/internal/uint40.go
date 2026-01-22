package internal

import "fmt"

// KVOffset is a 39-bit unsigned integer stored in 5 bytes with little-endian encoding,
// with a high-bit flag to indicate whether the data is in the WAL or KV data file.
// Bit 39 (high bit of byte 4) is the location flag: 0 = WAL file (default), 1 = KV data file.
// Bits 0-38 are the offset within that file.
type KVOffset [5]byte

const (
	// MaxKVOffset is the maximum offset value (39 bits = 512GB).
	MaxKVOffset = 1<<39 - 1
	// kvOffsetKVFlag is the bit flag indicating the offset points to KV data file (not WAL).
	kvOffsetKVFlag = 1 << 39
	// kvOffsetMask masks off the location flag to get the raw offset.
	kvOffsetMask = MaxKVOffset
)

// NewKVOffset creates a new KVOffset with the given offset and location.
// If inKVData is true, the offset points to the KV data file; otherwise it points to the WAL.
func NewKVOffset(v uint64, inKVData bool) KVOffset {
	if v > MaxKVOffset {
		panic(fmt.Sprintf("value %d overflows KVOffset (max %d)", v, MaxKVOffset))
	}
	if inKVData {
		return newKVOffsetRaw(v | kvOffsetKVFlag)
	}
	return newKVOffsetRaw(v)
}

// newKVOffsetRaw creates a KVOffset from a raw uint64 (including any flags).
func newKVOffsetRaw(v uint64) KVOffset {
	var u KVOffset
	u[0] = byte(v)
	u[1] = byte(v >> 8)
	u[2] = byte(v >> 16)
	u[3] = byte(v >> 24)
	u[4] = byte(v >> 32)
	return u
}

// IsZero returns true if the offset is zero (and not a WAL offset).
func (u KVOffset) IsZero() bool {
	return u[0] == 0 && u[1] == 0 && u[2] == 0 && u[3] == 0 && u[4] == 0
}

// IsWAL returns true if this offset points to data in the WAL file (flag=0).
func (u KVOffset) IsWAL() bool {
	return u[4]&0x80 == 0
}

// IsKVData returns true if this offset points to data in the KV data file (flag=1).
func (u KVOffset) IsKVData() bool {
	return u[4]&0x80 != 0
}

// Offset returns the raw offset value without the location flag.
func (u KVOffset) Offset() uint64 {
	return u.toUint64Raw() & kvOffsetMask
}

// toUint64Raw converts the KVOffset to a uint64 including the WAL flag.
func (u KVOffset) toUint64Raw() uint64 {
	return uint64(u[0]) | uint64(u[1])<<8 | uint64(u[2])<<16 | uint64(u[3])<<24 | uint64(u[4])<<32
}

// String implements fmt.Stringer.
func (u KVOffset) String() string {
	if u.IsWAL() {
		return fmt.Sprintf("%d(WAL)", u.Offset())
	}
	return fmt.Sprintf("%d", u.Offset())
}
