package solomachine

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
)

// Name returns the solo machine client name.
func Name() string {
	return types.SubModuleName
}
