package types

import (
	"strings"
)

// ValidatePortID validates that the provided port identifier is not empty
// and that it's
func ValidatePortID(portID string) bool {
	if strings.TrimSpace(portID) == "" {
		return false
	}
	if len(portID) < 3 || len(portID) > 10 {
		return false
	}
	return true
}
