package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// client error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodeClientExists           sdk.CodeType = 101
	CodeClientNotFound         sdk.CodeType = 102
	CodeClientFrozen           sdk.CodeType = 103
	CodeConsensusStateNotFound sdk.CodeType = 104
	CodeInvalidConsensusState  sdk.CodeType = 105
	CodeClientTypeNotFound     sdk.CodeType = 106
	CodeInvalidClientType      sdk.CodeType = 107
	CodeRootNotFound           sdk.CodeType = 108
)

// ErrClientExists implements sdk.Error
func ErrClientExists(codespace sdk.CodespaceType, clientID string) sdk.Error {
	return sdk.NewError(codespace, CodeClientExists, fmt.Sprintf("client with ID %s already exists", clientID))
}

// ErrClientNotFound implements sdk.Error
func ErrClientNotFound(codespace sdk.CodespaceType, clientID string) sdk.Error {
	return sdk.NewError(codespace, CodeClientNotFound, fmt.Sprintf("client with ID %s not found", clientID))
}

// ErrClientFrozen implements sdk.Error
func ErrClientFrozen(codespace sdk.CodespaceType, clientID string) sdk.Error {
	return sdk.NewError(codespace, CodeClientFrozen, fmt.Sprintf("client with ID %s is frozen due to misbehaviour", clientID))
}

// ErrConsensusStateNotFound implements sdk.Error
func ErrConsensusStateNotFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeConsensusStateNotFound, "consensus state not found")
}

// ErrInvalidConsensus implements sdk.Error
func ErrInvalidConsensus(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidConsensusState, "invalid consensus state")
}

// ErrClientTypeNotFound implements sdk.Error
func ErrClientTypeNotFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeClientTypeNotFound, "client type not found")
}

// ErrInvalidClientType implements sdk.Error
func ErrInvalidClientType(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidClientType, "client type mismatch")
}

// ErrRootNotFound implements sdk.Error
func ErrRootNotFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeRootNotFound, "commitment root not found")
}
