package internal

import (
	"fmt"
)

type ErrRestartNeeded struct{}

func (e ErrRestartNeeded) Error() string {
	return fmt.Sprintf("upgrade needed")
}

var _ error = ErrRestartNeeded{}
