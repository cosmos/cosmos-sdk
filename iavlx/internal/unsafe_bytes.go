package internal

// UnsafeBytes wraps a byte slice that may or not be a direct reference to
// a memory-mapped file.
// Generally, an unsafe byte slice cannot be expected to live longer than the
// Pin on the object it was obtained from.
// As long as it is pinned, it is safe to use the UnsafeBytes() method to get
// the underlying byte slice without copying.
// If the byte slice needs to be retained beyond the Pin's lifetime, the
// SafeCopy() method must be used to get a safe copy of the byte slice.
type UnsafeBytes struct {
	bz   []byte
	safe bool
}

// WrapUnsafeBytes wraps an unsafe byte slice as UnsafeBytes, indicating that
// it is unsafe to use without copying.
// Use this method when you are wrapping a byte slice obtained from a memory-mapped file.
func WrapUnsafeBytes(bz []byte) UnsafeBytes {
	return UnsafeBytes{bz: bz, safe: false}
}

// WrapSafeBytes wraps a safe byte slice as UnsafeBytes, indicating that
// it is safe to use without copying.
// Use this method when you are wrapping a byte slice that is known to be safe,
// e.g., a byte slice allocated in regular garbage-collected memory.
func WrapSafeBytes(bz []byte) UnsafeBytes {
	return UnsafeBytes{bz: bz, safe: true}
}

// IsNil returns true if the underlying byte slice is nil.
func (ub UnsafeBytes) IsNil() bool {
	return ub.bz == nil
}

// UnsafeBytes returns the underlying byte slice without copying.
// The caller must ensure that the byte slice is not used beyond the lifetime
// of the Pin on the object it was obtained from.
func (ub UnsafeBytes) UnsafeBytes() []byte {
	return ub.bz
}

// SafeCopy returns a safe copy of the underlying byte slice.
// If the underlying byte slice is already safe or nil, it is returned as is.
// If the underlying byte slice is unsafe, a copy is made and returned.
func (ub UnsafeBytes) SafeCopy() []byte {
	if ub.safe {
		return ub.bz
	}
	if ub.bz == nil {
		return nil
	}
	copied := make([]byte, len(ub.bz))
	copy(copied, ub.bz)
	return copied
}
