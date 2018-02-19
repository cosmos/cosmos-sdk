package sdk

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
)

type ErrorResponse struct {
	Success bool `json:"success,omitempty"`

	// Err is the error message if Success is false
	Err string `json:"error,omitempty"`

	// Code is set if Success is false
	Code int `json:"code,omitempty"`
}

// ErrorWithCode makes an ErrorResponse with the
// provided err's Error() content, and status code.
// It panics if err is nil.
func ErrorWithCode(err error, code int) *ErrorResponse {
	return &ErrorResponse{
		Err:  err.Error(),
		Code: code,
	}
}

// Ensure that ErrorResponse implements error
var _ error = (*ErrorResponse)(nil)

func (er *ErrorResponse) Error() string {
	return er.Err
}

// Ensure that ErrorResponse implements httpCoder
var _ httpCoder = (*ErrorResponse)(nil)

func (er *ErrorResponse) HTTPCode() int {
	return er.Code
}

var errNilBody = errors.Errorf("expecting a non-nil body")

// FparseJSON unmarshals into save, the body of the provided reader.
// Since it uses json.Unmarshal, save must be of a pointer type
// or compatible with json.Unmarshal.
func FparseJSON(r io.Reader, save interface{}) error {
	if r == nil {
		return errors.Wrap(errNilBody, "Reader")
	}

	dec := json.NewDecoder(r)
	if err := dec.Decode(save); err != nil {
		return errors.Wrap(err, "Decode/Unmarshal")
	}
	return nil
}

// ParseRequestJSON unmarshals into save, the body of the
// request. It closes the body of the request after parsing.
// Since it uses json.Unmarshal, save must be of a pointer type
// or compatible with json.Unmarshal.
func ParseRequestJSON(r *http.Request, save interface{}) error {
	if r == nil || r.Body == nil {
		return errNilBody
	}
	defer r.Body.Close()

	return FparseJSON(r.Body, save)
}

// ParseRequestAndValidateJSON unmarshals into save, the body of the
// request and invokes a validator on the saved content. To ensure
// validation, make sure to set tags "validate" on your struct as
// per https://godoc.org/gopkg.in/go-playground/validator.v9.
// It closes the body of the request after parsing.
// Since it uses json.Unmarshal, save must be of a pointer type
// or compatible with json.Unmarshal.
func ParseRequestAndValidateJSON(r *http.Request, save interface{}) error {
	if r == nil || r.Body == nil {
		return errNilBody
	}
	defer r.Body.Close()

	return FparseAndValidateJSON(r.Body, save)
}

// FparseAndValidateJSON like FparseJSON unmarshals into save,
// the body of the provided reader. However, it invokes the validator
// to check the set validators on your struct fields as per
// per https://godoc.org/gopkg.in/go-playground/validator.v9.
// Since it uses json.Unmarshal, save must be of a pointer type
// or compatible with json.Unmarshal.
func FparseAndValidateJSON(r io.Reader, save interface{}) error {
	if err := FparseJSON(r, save); err != nil {
		return err
	}
	return validate(save)
}

var theValidator = validator.New()

func validate(obj interface{}) error {
	return errors.Wrap(theValidator.Struct(obj), "Validate")
}

// WriteSuccess JSON marshals the content provided, to an HTTP
// response, setting the provided status code and setting header
// "Content-Type" to "application/json".
func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteCode(w, data, 200)
}

// WriteCode JSON marshals content, to an HTTP response,
// setting the provided status code, and setting header
// "Content-Type" to "application/json". If JSON marshalling fails
// with an error, WriteCode instead writes out the error invoking
// WriteError.
func WriteCode(w http.ResponseWriter, out interface{}, code int) {
	blob, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		WriteError(w, err)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write(blob)
	}
}

type httpCoder interface {
	HTTPCode() int
}

// WriteError is a convenience function to write out an
// error to an http.ResponseWriter, to send out an error
// that's structured as JSON i.e the form
//    {"error": sss, "code": ddd}
// If err implements the interface HTTPCode() int,
// it will use that status code otherwise, it will
// set code to be http.StatusBadRequest
func WriteError(w http.ResponseWriter, err error) {
	code := http.StatusBadRequest
	if httpC, ok := err.(httpCoder); ok {
		code = httpC.HTTPCode()
	}

	WriteCode(w, ErrorWithCode(err, code), code)
}
