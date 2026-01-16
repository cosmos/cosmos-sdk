package internal

import "bytes"

// cmpInlineKeyPrefix compares a key's inline prefix against another full key.
//
// Parameters:
//   - inlinePrefix: the 8-byte inline prefix buffer from BranchLayout
//   - inlineLen: the actual full key length (may exceed 8)
//   - other: the full key to compare against
//
// Returns:
//   - cmp: the comparison result (-1, 0, 1)
//   - needFullKey: true only if the full key must be read to determine ordering
func cmpInlineKeyPrefix(inlinePrefix []byte, inlineLen int, other []byte) (cmp int, needFullKey bool) {
	// How many bytes of our key are available inline
	inlineAvail := inlineLen
	if inlineAvail > MaxInlineKeyCopyLen {
		inlineAvail = MaxInlineKeyCopyLen
	}

	// Compare the overlapping portion
	cmpLen := inlineAvail
	if len(other) < cmpLen {
		cmpLen = len(other)
	}

	cmp = bytes.Compare(inlinePrefix[:cmpLen], other[:cmpLen])
	if cmp != 0 {
		// Content differs in overlapping portion - result is determined
		return cmp, false
	}

	// First cmpLen bytes are equal, now consider lengths
	if inlineLen <= MaxInlineKeyCopyLen {
		// Our complete key is inline
		if inlineLen < len(other) {
			return -1, false // our key is shorter
		} else if inlineLen > len(other) {
			return 1, false // our key is longer
		}
		return 0, false // identical
	}

	// inlineLen > 8: our key extends beyond the inline prefix
	if len(other) <= MaxInlineKeyCopyLen {
		// Other is fully covered, our key is longer with matching prefix
		return 1, false
	}

	// Both keys are > 8 bytes and first 8 bytes match - must read full key
	return 0, true
}
