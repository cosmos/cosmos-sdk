package util

import (
	"errors"
	"fmt"
)

func CombineErrors(ret, also error, desc string) error {
	if also != nil {
		if ret != nil {
			ret = fmt.Errorf("%v; %s: %w", ret, desc, also)
		} else {
			ret = also
		}
	}
	return ret
}
