package errors

import "fmt"

// Panic panics on error
// Should be only used with interface methods, which require return error, but the
// error is always nil
func Panic(err error) {
	if err != nil {
		panic(fmt.Errorf("logic error - this should never happen. %w", err))
	}
}
