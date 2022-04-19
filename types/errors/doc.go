// Package errors provides a shared set of errors for use in the SDK,
// aliases functionality in the github.com/cosmos/cosmos-sdk/errors module
// that used to be in this package, and provides some helpers for converting
// errors to ABCI response code.
//
// New code should generally import github.com/cosmos/cosmos-sdk/errors directly
// and define a custom set of errors in custom codespace, rather than importing
// this package.
package errors
