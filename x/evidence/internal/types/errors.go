// DONTCOVER
package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/evidence module errors that reserve codes 400-499
var (
	ErrNoEvidenceHandlerExists = sdkerrors.Register(ModuleName, 400, "unregistered handler for evidence type")
	ErrInvalidEvidence         = sdkerrors.Register(ModuleName, 401, "invalid evidence")
	ErrNoEvidenceExists        = sdkerrors.Register(ModuleName, 402, "evidence does not exist")
	ErrEvidenceExists          = sdkerrors.Register(ModuleName, 403, "evidence already exists")
)
