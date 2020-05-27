package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC connection sentinel errors
var (
	ErrConnectionExists              = sdkerrors.Register(SubModuleName, 2, "connection already exists")
	ErrConnectionNotFound            = sdkerrors.Register(SubModuleName, 3, "connection not found")
	ErrClientConnectionPathsNotFound = sdkerrors.Register(SubModuleName, 4, "light client connection paths not found")
	ErrConnectionPath                = sdkerrors.Register(SubModuleName, 5, "connection path is not associated to the given light client")
	ErrInvalidConnectionState        = sdkerrors.Register(SubModuleName, 6, "invalid connection state")
	ErrInvalidCounterparty           = sdkerrors.Register(SubModuleName, 7, "invalid counterparty connection")
	ErrInvalidConnection             = sdkerrors.Register(SubModuleName, 8, "invalid connection")
)
