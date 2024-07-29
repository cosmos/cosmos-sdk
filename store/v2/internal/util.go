package internal

import "strings"

func IsMemoryStoreKey(key string) bool {
	return strings.HasPrefix(key, "memory:")
}
