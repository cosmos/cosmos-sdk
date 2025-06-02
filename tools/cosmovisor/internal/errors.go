package internal

import (
	"fmt"
)

type ErrUpgradeNeeded struct {
	KnownHeight uint64
}

func (e ErrUpgradeNeeded) Error() string {
	return fmt.Sprintf("upgrade needed")
}

var _ error = ErrUpgradeNeeded{}
