package util

import (
	"fmt"

	log "github.com/tendermint/tendermint/libs/log"
)

func CombineErrors(ret error, also error, desc string) error {
	if also != nil {
		if ret != nil {
			ret = fmt.Errorf("%w; %v: %v", ret, desc, also)
		} else {
			ret = also
		}
	}
	return ret
}

// LogDeferred logs an error in a deferred function call if the returned error is non-nil.
func LogDeferred(logger log.Logger, f func() error) {
	if err := f(); err != nil {
		logger.Error(err.Error())
	}
}
