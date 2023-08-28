package util

// CopyBytes copies a byte slice. It returns <nil> if the argument is <nil>.
func CopyBytes(bz []byte) []byte {
	if bz == nil {
		return nil
	}

	ret := make([]byte, len(bz))
	_ = copy(ret, bz)

	return ret
}
