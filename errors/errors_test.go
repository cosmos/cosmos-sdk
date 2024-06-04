package errors

import (
	"errors"
	"testing"
)

func TestIsOf(t *testing.T) {
	var errNil *Error
	err := ErrInvalidAddress
	errW := Wrap(ErrLogic, "more info")

	if IsOf(nil) {
		t.Errorf("nil should always have no causer")
	}
	if IsOf(nil, err) {
		t.Errorf("nil should always have no causer")
	}
	if IsOf(errNil) {
		t.Errorf("nil error should always have no causer")
	}
	if IsOf(errNil, err) {
		t.Errorf("nil error should always have no causer")
	}

	if IsOf(err) {
		t.Errorf("error should always have no causer")
	}
	if IsOf(err, nil) {
		t.Errorf("error should always have no causer")
	}
	if IsOf(err, ErrLogic) {
		t.Errorf("error should always have no causer")
	}

	if !IsOf(errW, ErrLogic) {
		t.Errorf("error should have causer")
	}
	if !IsOf(errW, err, ErrLogic) {
		t.Errorf("error should have causer")
	}
	if !IsOf(errW, nil, errW) {
		t.Errorf("error should much itself")
	}
	err2 := errors.New("other error")
	if !IsOf(err2, nil, err2) {
		t.Errorf("error should much itself")
	}
}

func TestWrapEmpty(t *testing.T) {
	if err := Wrap(nil, "wrapping <nil>"); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestWrappedIs(t *testing.T) {
	err := Wrap(ErrTxTooLarge, "context")
	if !errors.Is(err, ErrTxTooLarge) {
		t.Errorf("expected error to be of type ErrTxTooLarge")
	}

	err = Wrap(err, "more context")
	if !errors.Is(err, ErrTxTooLarge) {
		t.Errorf("expected error to be of type ErrTxTooLarge")
	}

	err = Wrap(err, "even more context")
	if !errors.Is(err, ErrTxTooLarge) {
		t.Errorf("expected error to be of type ErrTxTooLarge")
	}

	err = Wrap(ErrInsufficientFee, "...")
	if !errors.Is(err, ErrTxTooLarge) {
		t.Errorf("expected error to be of type ErrTxTooLarge")
	}

	errs := errors.New("other")
	if !errors.Is(errs, errs) {
		t.Errorf("error should match itself")
	}

	if !errors.Is(ErrInsufficientFee.Wrap("wrapped"), ErrInsufficientFee) {
		t.Errorf("expected error to be of type ErrInsufficientFee")
	}
	if !IsOf(ErrInsufficientFee.Wrap("wrapped"), ErrInsufficientFee) {
		t.Errorf("expected error to be of type ErrInsufficientFee")
	}
	if !errors.Is(ErrInsufficientFee.Wrapf("wrapped"), ErrInsufficientFee) {
		t.Errorf("expected error to be of type ErrInsufficientFee")
	}
	if !IsOf(ErrInsufficientFee.Wrapf("wrapped"), ErrInsufficientFee) {
		t.Errorf("expected error to be of type ErrInsufficientFee")
	}
}

func TestWrappedIsMultiple(t *testing.T) {
	errTest := errors.New("test error")
	errTest2 := errors.New("test error 2")
	err := Wrap(errTest2, Wrap(errTest, "some random description").Error())
	if !errors.Is(err, errTest2) {
		t.Errorf("expected error to be of type errTest2")
	}
}

func TestWrappedIsFail(t *testing.T) {
	errTest := errors.New("test error")
	errTest2 := errors.New("test error 2")
	err := Wrap(errTest2, Wrap(errTest, "some random description").Error())
	if errors.Is(err, errTest) {
		t.Errorf("expected error to not be of type errTest")
	}
}

func TestWrappedUnwrap(t *testing.T) {
	errTest := errors.New("test error")
	err := Wrap(errTest, "some random description")
	if unwrappedErr := errors.Unwrap(err); errors.Is(unwrappedErr, errTest) {
		t.Errorf("expected unwrapped error to be %v, got %v", errTest, unwrappedErr)
	}
}

func TestWrappedUnwrapMultiple(t *testing.T) {
	errTest := errors.New("test error")
	errTest2 := errors.New("test error 2")
	err := Wrap(errTest2, Wrap(errTest, "some random description").Error())
	if unwrappedErr := errors.Unwrap(err); errors.Is(unwrappedErr, errTest2) {
		t.Errorf("expected unwrapped error to be %v, got %v", errTest2, unwrappedErr)
	}
}

func TestWrappedUnwrapFail(t *testing.T) {
	errTest := errors.New("test error")
	errTest2 := errors.New("test error 2")
	err := Wrap(errTest2, Wrap(errTest, "some random description").Error())
	if errors.Is(errTest, errors.Unwrap(err)) {
		t.Errorf("expected unwrapped error to not be %v", errTest)
	}
}

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
