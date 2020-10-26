package errors

import "fmt"

// ErrCallback is a type for a function which returns an error and is used as a callback
type ErrCallback = func() error

// CallbackLog wraps an ErrCallback function which will log an error if it is returned
func CallbackLog(f ErrCallback) func() {
	return func() {
		if err := f(); err != nil {
			fmt.Println(err)
		}
	}
}
