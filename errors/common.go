package errors

/**
*    Copyright (C) 2017 Ethan Frey
**/

import (
	rawerr "errors"
	"fmt"

	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin"
)

var (
	errDecoding          = rawerr.New("Error decoding input")
	errUnauthorized      = rawerr.New("Unauthorized")
	errInvalidAddress    = rawerr.New("Invalid Address")
	errInvalidCoins      = rawerr.New("Invalid Coins")
	errInvalidFormat     = rawerr.New("Invalid Format")
	errInvalidSequence   = rawerr.New("Invalid Sequence")
	errInvalidSignature  = rawerr.New("Invalid Signature")
	errInsufficientFees  = rawerr.New("Insufficient Fees")
	errInsufficientFunds = rawerr.New("Insufficient Funds")
	errNoInputs          = rawerr.New("No Input Coins")
	errNoOutputs         = rawerr.New("No Output Coins")
	errTooLarge          = rawerr.New("Input size too large")
	errMissingSignature  = rawerr.New("Signature missing")
	errTooManySignatures = rawerr.New("Too many signatures")
	errNoChain           = rawerr.New("No chain id provided")
	errWrongChain        = rawerr.New("Wrong chain for tx")
	errUnknownTxType     = rawerr.New("Tx type unknown")
)

func ErrUnknownTxType(tx basecoin.Tx) TMError {
	msg := fmt.Sprintf("%T", tx.Unwrap())
	w := errors.Wrap(errUnknownTxType, msg)
	return WithCode(w, abci.CodeType_UnknownRequest)
}

func IsUnknownTxTypeErr(err error) bool {
	return IsSameError(errUnknownTxType, err)
}

func ErrInternal(msg string) TMError {
	return New(msg, abci.CodeType_InternalError)
}

// IsInternalErr matches any error that is not classified
func IsInternalErr(err error) bool {
	return HasErrorCode(err, abci.CodeType_InternalError)
}

func ErrDecoding() TMError {
	return WithCode(errDecoding, abci.CodeType_EncodingError)
}

func IsDecodingErr(err error) bool {
	return IsSameError(errDecoding, err)
}

func ErrUnauthorized() TMError {
	return WithCode(errUnauthorized, abci.CodeType_Unauthorized)
}

// IsUnauthorizedErr is generic helper for any unauthorized errors,
// also specific sub-types
func IsUnauthorizedErr(err error) bool {
	return HasErrorCode(err, abci.CodeType_Unauthorized)
}

func ErrMissingSignature() TMError {
	return WithCode(errMissingSignature, abci.CodeType_Unauthorized)
}

func IsMissingSignatureErr(err error) bool {
	return IsSameError(errMissingSignature, err)
}

func ErrTooManySignatures() TMError {
	return WithCode(errTooManySignatures, abci.CodeType_Unauthorized)
}

func IsTooManySignaturesErr(err error) bool {
	return IsSameError(errTooManySignatures, err)
}

func ErrInvalidSignature() TMError {
	return WithCode(errInvalidSignature, abci.CodeType_Unauthorized)
}

func IsInvalidSignatureErr(err error) bool {
	return IsSameError(errInvalidSignature, err)
}

func ErrNoChain() TMError {
	return WithCode(errNoChain, abci.CodeType_Unauthorized)
}

func IsNoChainErr(err error) bool {
	return IsSameError(errNoChain, err)
}

func ErrWrongChain(chain string) TMError {
	msg := errors.Wrap(errWrongChain, chain)
	return WithCode(msg, abci.CodeType_Unauthorized)
}

func IsWrongChainErr(err error) bool {
	return IsSameError(errWrongChain, err)
}

func InvalidAddress() TMError {
	return WithCode(errInvalidAddress, abci.CodeType_BaseInvalidInput)
}

func InvalidCoins() TMError {
	return WithCode(errInvalidCoins, abci.CodeType_BaseInvalidInput)
}

func InvalidFormat() TMError {
	return WithCode(errInvalidFormat, abci.CodeType_BaseInvalidInput)
}

func InvalidSequence() TMError {
	return WithCode(errInvalidSequence, abci.CodeType_BaseInvalidInput)
}

func InsufficientFees() TMError {
	return WithCode(errInsufficientFees, abci.CodeType_BaseInvalidInput)
}

func InsufficientFunds() TMError {
	return WithCode(errInsufficientFunds, abci.CodeType_BaseInvalidInput)
}

func NoInputs() TMError {
	return WithCode(errNoInputs, abci.CodeType_BaseInvalidInput)
}

func NoOutputs() TMError {
	return WithCode(errNoOutputs, abci.CodeType_BaseInvalidOutput)
}

func ErrTooLarge() TMError {
	return WithCode(errTooLarge, abci.CodeType_EncodingError)
}

func IsTooLargeErr(err error) bool {
	return IsSameError(errTooLarge, err)
}
