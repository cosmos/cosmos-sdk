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
	return true
}
