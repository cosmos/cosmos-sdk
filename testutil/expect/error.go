package expect

import "github.com/stretchr/testify/require"

// Error checks if the received error is not nil and it's string contains
// the `expected` message. If `expected` is empty then received should be nil.
func Error(r *require.Assertions, expected string, received error) {
	if expected == "" {
		r.NoError(received)
	} else {
		r.Error(received)
		if received != nil { // if Assertions is not "require" then Error won't fail
			r.Contains(received.Error(), expected)
		}
	}
}
