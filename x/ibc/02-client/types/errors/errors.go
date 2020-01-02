package errors

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SubModuleName is the error codespace
const SubModuleName string = "ibc/client"

// IBC client sentinel errors
var (
	ErrClientExists           = sdkerrors.Register(SubModuleName, 1, "light client already exists")
	ErrClientNotFound         = sdkerrors.Register(SubModuleName, 2, "light client not found")
	ErrClientFrozen           = sdkerrors.Register(SubModuleName, 3, "light client is frozen due to misbehaviour")
	ErrConsensusStateNotFound = sdkerrors.Register(SubModuleName, 4, "consensus state not found")
	ErrInvalidConsensus       = sdkerrors.Register(SubModuleName, 5, "invalid consensus state")
	ErrClientTypeNotFound     = sdkerrors.Register(SubModuleName, 6, "client type not found")
	ErrInvalidClientType      = sdkerrors.Register(SubModuleName, 7, "invalid client type")
	ErrRootNotFound           = sdkerrors.Register(SubModuleName, 8, "commitment root not found")
	ErrInvalidHeader          = sdkerrors.Register(SubModuleName, 9, "invalid block header")
	ErrInvalidEvidence        = sdkerrors.Register(SubModuleName, 10, "invalid light client misbehaviour evidence")
	ErrCommitterNotFound      = sdkerrors.Register(SubModuleName, 11, "commiter not found")
	ErrInvalidCommitter       = sdkerrors.Register(SubModuleName, 12, "invalid commiter")
)
