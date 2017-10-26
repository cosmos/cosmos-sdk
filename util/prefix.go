package util

// PrefixedKey returns the absolute path to a given key in a particular
// app's state-space
//
// This is useful for to set up queries for this particular app data
func PrefixedKey(app string, key []byte) []byte {
	prefix := append([]byte(app), byte(0))
	return append(prefix, key...)
}
