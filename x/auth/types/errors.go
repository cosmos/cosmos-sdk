package types

import "cosmossdk.io/errors"

// x/auth module sentinel errors
var (
	ErrPubKeyChangeDisabled = errors.Register(ModuleName, 2, "pubkey change is disabled")
	ErrInvalidNewPubKey     = errors.Register(ModuleName, 3, "invalid new pubkey")
	ErrInvalidPubKeyProof   = errors.Register(ModuleName, 4, "invalid pubkey proof of possession")
	ErrAccountNotRekeyable  = errors.Register(ModuleName, 5, "account cannot change its pubkey")
)
