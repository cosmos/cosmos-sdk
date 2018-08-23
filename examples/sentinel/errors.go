package sentinel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	DefaultCodeSpace sdk.CodespaceType = 19

	CodeInvalidPubKey          sdk.CodeType = 1
	CodeTimeInterval           sdk.CodeType = 2
	CodeInvalidIpAdress        sdk.CodeType = 3
	CodeUnknownIpAddress       sdk.CodeType = 4
	CodeUnknownSessionid       sdk.CodeType = 5
	CodeInvalidSessionid       sdk.CodeType = 6
	CodeInvalidNetspeed        sdk.CodeType = 7
	CodeInvalidPricePerGb      sdk.CodeType = 8
	CodeMarshal                sdk.CodeType = 9
	CodeUnMarshal              sdk.CodeType = 10
	CodeKeyBase                sdk.CodeType = 11
	CodeSignMsg                sdk.CodeType = 12
	CodeAccountAddressExist    sdk.CodeType = 13
	CodeAccountAddressNotExist sdk.CodeType = 14
	CodeInvalidLocation        sdk.CodeType = 15
	CodeBech32Decode           sdk.CodeType = 16
)

func ErrInvalidPubKey(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeInvalidPubKey, msg)
}
func ErrTimeInterval(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeTimeInterval, msg)
}
func ErrInvalidIpAdress(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeInvalidIpAdress, msg)
}
func ErrUnknownIpAddress(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeUnknownIpAddress, msg)
}
func ErrUnknownSessionid(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeUnknownSessionid, msg)
}
func ErrInvalidSessionid(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeInvalidSessionid, msg)
}
func ErrInvalidNetspeed(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeInvalidNetspeed, msg)
}
func ErrMarshal(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeMarshal, msg)
}
func ErrUnMarshal(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeUnMarshal, msg)
}
func ErrKeyBase(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeKeyBase, msg)
}
func ErrSignMsg(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeSignMsg, msg)
}
func ErrAccountAddressExist(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeAccountAddressExist, msg)
}
func ErrAccountAddressNotExist(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeAccountAddressNotExist, msg)
}
func ErrInvalidPricePerGb(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeInvalidPricePerGb, msg)
}
func ErrInvalidLocation(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeInvalidLocation, msg)
}
func ErrBech32Decode(msg string) sdk.Error {

	return sdk.NewError(DefaultCodeSpace, CodeBech32Decode, msg)
}
