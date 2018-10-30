package utils

import (
	"fmt"
)

func bz(s string) []byte { return []byte(s) }

// Used for tests - formats integer to bytes
// nolint
func KeyFmt(i int) []byte { return bz(fmt.Sprintf("key%0.8d", i)) }
func ValFmt(i int) []byte { return bz(fmt.Sprintf("value%0.8d", i)) }
