package util

import (
	"fmt"
)

func CombineErrors(ret, also error, desc string) error {
	if also != nil {
		if ret != nil {
			ret = fmt.Errorf("%w; %v: %w", ret, desc, also)
		} else {
			ret = also
		}
	}
	return ret
}
