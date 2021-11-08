package errors

import (
	"github.com/cosmos/cosmos-sdk/types/errors"
)

// mathCodespace is the codespace for all errors defined in orm package
const ormCodespace = "orm"

var (
	// ErrORMNotFound defines an error when requested entity doesn't exist in the state.
	ErrORMNotFound = errors.Register(ormCodespace, 11, "")
)
