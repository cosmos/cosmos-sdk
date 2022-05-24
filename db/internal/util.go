package util

import (
	"fmt"

	dbm "github.com/cosmos/cosmos-sdk/db"
)

func ValidateKv(key, value []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	if value == nil {
		return dbm.ErrValueNil
	}
	return nil
}

func CombineErrors(ret error, also error, desc string) error {
	if also != nil {
		if ret != nil {
			ret = fmt.Errorf("%w; %s: %v", ret, desc, also)
		} else {
			ret = also
		}
	}
	return ret
}
