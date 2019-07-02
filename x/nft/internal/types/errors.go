package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CodeType definition
type CodeType = sdk.CodeType

// NFT error code
const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeInvalidCollection CodeType = 650
	CodeUnknownCollection CodeType = 651
	CodeInvalidNFT        CodeType = 652
	CodeUnknownNFT        CodeType = 653
	CodeNFTAlreadyExists  CodeType = 654
	CodeEmptyMetadata     CodeType = 655
)

// ErrInvalidCollection is an error
func ErrInvalidCollection(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidCollection, "invalid NFT collection")
}

// ErrUnknownCollection is an error
func ErrUnknownCollection(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeUnknownCollection, msg)
	}
	return sdk.NewError(codespace, CodeUnknownCollection, "unknown NFT collection")
}

// ErrInvalidNFT is an error
func ErrInvalidNFT(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidNFT, "invalid NFT")
}

// ErrNFTAlreadyExists is an error when an invalid NFT is minted
func ErrNFTAlreadyExists(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeNFTAlreadyExists, msg)
	}
	return sdk.NewError(codespace, CodeNFTAlreadyExists, "NFT already exists")
}

// ErrUnknownNFT is an error
func ErrUnknownNFT(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeUnknownNFT, msg)
	}
	return sdk.NewError(codespace, CodeUnknownNFT, "unknown NFT")
}

// ErrEmptyMetadata is an error when metadata is empty
func ErrEmptyMetadata(codespace sdk.CodespaceType, msg string) sdk.Error {
	if msg != "" {
		return sdk.NewError(codespace, CodeEmptyMetadata, msg)
	}
	return sdk.NewError(codespace, CodeEmptyMetadata, "NFT metadata can't be empty")
}
