package errors

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ModuleName is the error codespace
const ModuleName string = "ibc/client"

// IBC client sentinel errors
var (
	ErrClientExists           = sdkerrors.Register(ModuleName, 1, "light client already exists")
	ErrClientNotFound         = sdkerrors.Register(ModuleName, 2, "light client not found")
	ErrClientFrozen           = sdkerrors.Register(ModuleName, 3, "light client is frozen due to misbehaviour")
	ErrConsensusStateNotFound = sdkerrors.Register(ModuleName, 4, "consensus state not found")
	ErrInvalidConsensus       = sdkerrors.Register(ModuleName, 5, "invalid consensus state")
	ErrClientTypeNotFound     = sdkerrors.Register(ModuleName, 6, "client type not found")
	ErrInvalidClientType      = sdkerrors.Register(ModuleName, 7, "invalid client type")
	ErrRootNotFound           = sdkerrors.Register(ModuleName, 8, "commitment root not found")
	ErrInvalidHeader          = sdkerrors.Register(ModuleName, 9, "invalid block header")
	ErrInvalidEvidence        = sdkerrors.Register(ModuleName, 10, "invalid light client misbehaviour evidence")
	ErrCommitterNotFound      = sdkerrors.Register(ModuleName, 11, "commiter not found")
	ErrInvalidCommitter       = sdkerrors.Register(ModuleName, 12, "invalid commiter")
)
