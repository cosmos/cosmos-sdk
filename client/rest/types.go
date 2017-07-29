package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/go-crypto/keys"
	data "github.com/tendermint/go-wire/data"
)

type CreateKeyRequest struct {
	Name       string `json:"name,omitempty" validate:"required,min=4,printascii"`
	Passphrase string `json:"passphrase,omitempty" validate:"required,min=10"`

	// Algo is the requested algorithm to create the key
	Algo string `json:"algo,omitempty"`
}

type DeleteKeyRequest struct {
	Name       string `json:"name,omitempty" validate:"required,min=4,printascii"`
	Passphrase string `json:"passphrase,omitempty" validate:"required,min=10"`
}

type UpdateKeyRequest struct {
	Name    string `json:"name,omitempty" validate:"required,min=4,printascii"`
	OldPass string `json:"passphrase,omitempty" validate:"required,min=10"`
	NewPass string `json:"new_passphrase,omitempty" validate:"required,min=10"`
}

type SignRequest struct {
	Name     string `json:"name,omitempty" validate:"required,min=4,printascii"`
	Password string `json:"password,omitempty" validate:"required,min=10"`

	Tx basecoin.Tx `json:"tx" validate:"required"`
}

type ErrorResponse struct {
	Success bool `json:"success,omitempty"`

	// Error is the error message if Success is false
	Error string `json:"error,omitempty"`

	// Code is set if Success is false
	Code int `json:"code,omitempty"`
}

type CreateKeyResponse struct {
	Key  keys.Info `json:"key,omitempty"`
	Seed string    `json:"seed_phrase,omitempty"`
}

// SendInput is the request to send an amount from one actor to another.
// Note: Not using the `validator:""` tags here because SendInput has
// many fields so it would be nice to figure out all the invalid
// inputs and report them back to the caller, in one shot.
type SendInput struct {
	Fees     *coin.Coin `json:"fees"`
	Multi    bool       `json:"multi,omitempty"`
	Sequence uint32     `json:"sequence"`

	To     *basecoin.Actor `json:"to"`
	From   *basecoin.Actor `json:"from"`
	Amount coin.Coins      `json:"amount"`
}

// Validators

var theValidator = validator.New()

func validate(req interface{}) error {
	return errors.Wrap(theValidator.Struct(req), "Validate")
}

// Helpers
func parseRequestJSON(r *http.Request, save interface{}) error {
	defer r.Body.Close()

	slurp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return errors.Wrap(err, "Read Request")
	}
	if err := json.Unmarshal(slurp, save); err != nil {
		return errors.Wrap(err, "Parse")
	}
	return validate(save)
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeCode(w, data, 200)
}

func writeCode(w http.ResponseWriter, out interface{}, code int) {
	blob, err := data.ToJSON(out)
	if err != nil {
		writeError(w, err)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write(blob)
	}
}

func writeError(w http.ResponseWriter, err error) {
	resp := &ErrorResponse{
		Code:  406,
		Error: err.Error(),
	}
	writeCode(w, resp, 406)
}
