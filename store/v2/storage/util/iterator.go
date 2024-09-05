package util

// IterateWithPrefix returns the begin and end keys for an iterator over a domain
// and prefix.
func IterateWithPrefix(prefix, begin, end []byte) ([]byte, []byte) {
	if len(prefix) == 0 {
		return begin, end
	}

	begin = cloneAppend(prefix, begin)

	if end == nil {
		end = CopyIncr(prefix)
	} else {
		end = cloneAppend(prefix, end)
	}

	return begin, end
}

func cloneAppend(front, tail []byte) (res []byte) {
	res = make([]byte, len(front)+len(tail))

	n := copy(res, front)
	copy(res[n:], tail)

	return res
}

func CopyIncr(bz []byte) []byte {
	if len(bz) == 0 {
		panic("copyIncr expects non-zero bz length")
	}

	ret := make([]byte, len(bz))
	copy(ret, bz)

	for i := len(bz) - 1; i >= 0; i-- {
		if ret[i] < byte(0xFF) {
			ret[i]++
			return ret
		}

		ret[i] = byte(0x00)

		if i == 0 {
			// overflow
			return nil
		}
	}

	return nil
}
