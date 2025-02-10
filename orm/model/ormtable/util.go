package ormtable

// prefixEndBytes returns the []byte that would end a
// range query for all []byte with a certain prefix
// Deals with last byte of prefix being FF without overflowing
func prefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for {
		if end[len(end)-1] != byte(255) {
			end[len(end)-1]++
			break
		}

		end = end[:len(end)-1]

		if len(end) == 0 {
			end = nil
			break
		}
	}

	return end
}

// inclusiveEndBytes returns the []byte that would end a
// range query such that the input would be included
func inclusiveEndBytes(inclusiveBytes []byte) []byte {
	return append(inclusiveBytes, byte(0x00))
}
