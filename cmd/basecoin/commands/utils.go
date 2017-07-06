package commands

import "encoding/hex"

// Returns true for non-empty hex-string prefixed with "0x"
func isHex(s string) bool {
	if len(s) > 2 && s[:2] == "0x" {
		_, err := hex.DecodeString(s[2:])
		if err != nil {
			return false
		}
		return true
	}
	return false
}

// StripHex remove the first two hex bytes
func StripHex(s string) string {
	if isHex(s) {
		return s[2:]
	}
	return s
}
