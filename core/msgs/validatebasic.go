package msgs

import (
	"cosmossdk.io/core/address"
)

type ValidateBasic interface {
	ValidateBasic(address.Codec) error
}
