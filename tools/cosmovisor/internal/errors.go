package internal

import (
	"fmt"
)

type ErrUpgradeNeeded struct{}

func (e ErrUpgradeNeeded) Error() string {
	return fmt.Sprintf("upgrade needed")
}

var _ error = ErrUpgradeNeeded{}
