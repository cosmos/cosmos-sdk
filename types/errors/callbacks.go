package errors

import (
	"log"
	"os"
)

// TODO: use a proper logger here and make it customizable
var logger = log.New(os.Stderr, "defer_callback: ", log.LstdFlags|log.LUTC|log.Llongfile)

// Callback is a type for a function which returns an error and is used as a callback
type Callback = func() error

// CallbackLog wraps an ErrCallback function which will log an error if it is returned
func CallbackLog(f Callback) func() {
	return func() {
		if err := f(); err != nil {
			logger.Println(err)
		}
	}
}

// LogError logs an error if it's not nil
func LogError(err error) {
	if err != nil {
		logger.Println(err)
	}
}
