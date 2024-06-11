package errors

import (
	"testing"
)

func TestABCIError(t *testing.T) {
	if err := ABCIError(testCodespace, 2, "custom"); err.Error() != "custom: tx parse error" {
		t.Errorf("expected error message: custom: tx parse error, got: %v", err.Error())
	}
	if err := ABCIError("unknown", 1, "custom"); err.Error() != "custom: unknown" {
		t.Errorf("expected error message: custom: unknown, got: %v", err.Error())
	}
}

const testCodespace = "testtesttest"

var (
	ErrTxDecode                = Register(testCodespace, 2, "tx parse error")
	ErrInvalidSequence         = Register(testCodespace, 3, "invalid sequence")
	ErrUnauthorized            = Register(testCodespace, 4, "unauthorized")
	ErrInsufficientFunds       = Register(testCodespace, 5, "insufficient funds")
	ErrUnknownRequest          = Register(testCodespace, 6, "unknown request")
	ErrInvalidAddress          = Register(testCodespace, 7, "invalid address")
	ErrInvalidPubKey           = Register(testCodespace, 8, "invalid pubkey")
	ErrUnknownAddress          = Register(testCodespace, 9, "unknown address")
	ErrInvalidCoins            = Register(testCodespace, 10, "invalid coins")
	ErrOutOfGas                = Register(testCodespace, 11, "out of gas")
	ErrInsufficientFee         = Register(testCodespace, 13, "insufficient fee")
	ErrTooManySignatures       = Register(testCodespace, 14, "maximum number of signatures exceeded")
	ErrNoSignatures            = Register(testCodespace, 15, "no signatures supplied")
	ErrJSONMarshal             = Register(testCodespace, 16, "failed to marshal JSON bytes")
	ErrJSONUnmarshal           = Register(testCodespace, 17, "failed to unmarshal JSON bytes")
	ErrInvalidRequest          = Register(testCodespace, 18, "invalid request")
	ErrMempoolIsFull           = Register(testCodespace, 20, "mempool is full")
	ErrTxTooLarge              = Register(testCodespace, 21, "tx too large")
	ErrKeyNotFound             = Register(testCodespace, 22, "key not found")
	ErrorInvalidSigner         = Register(testCodespace, 24, "tx intended signer does not match the given signer")
	ErrInvalidChainID          = Register(testCodespace, 28, "invalid chain-id")
	ErrInvalidType             = Register(testCodespace, 29, "invalid type")
	ErrUnknownExtensionOptions = Register(testCodespace, 31, "unknown extension options")
	ErrPackAny                 = Register(testCodespace, 33, "failed packing protobuf message to Any")
	ErrLogic                   = Register(testCodespace, 35, "internal logic error")
	ErrIO                      = Register(testCodespace, 39, "Internal IO error")
)
