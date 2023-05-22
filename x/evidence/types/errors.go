// DONTCOVER
package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/evidence module sentinel errors
var (
	ErrNoEvidenceHandlerExists = errorsmod.Register(ModuleName, 2, "unregistered handler for evidence type")
	ErrInvalidEvidence         = errorsmod.Register(ModuleName, 3, "invalid evidence")
	ErrNoEvidenceExists        = errorsmod.Register(ModuleName, 4, "evidence does not exist")
	ErrEvidenceExists          = errorsmod.Register(ModuleName, 5, "evidence already exists")
)
