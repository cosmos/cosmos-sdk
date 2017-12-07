package rest

import (
	"github.com/tendermint/go-crypto/keys"

	sdk "github.com/cosmos/cosmos-sdk"
)

// TODO: consistency between Passphrase and Password !!!

type RequestCreate struct {
	Name       string `json:"name,omitempty" validate:"required,min=3,printascii"`
	Passphrase string `json:"password,omitempty" validate:"required,min=10"`

	// Algo is the requested algorithm to create the key
	Algo string `json:"algo,omitempty"`
}

type ResponseCreate struct {
	Key  keys.Info `json:"key,omitempty"`
	Seed string    `json:"seed_phrase,omitempty"`
}

//-----------------------------------------------------------------------

type RequestRecover struct {
	Name       string `json:"name,omitempty" validate:"required,min=3,printascii"`
	Passphrase string `json:"password,omitempty" validate:"required,min=10"`
	Seed       string `json:"seed_phrase,omitempty" validate:"required,min=23"`

	// Algo is the requested algorithm to create the key
	Algo string `json:"algo,omitempty"`
}

type ResponseRecover struct {
	Key keys.Info `json:"key,omitempty"`
}

//-----------------------------------------------------------------------

type RequestDelete struct {
	Name       string `json:"name,omitempty" validate:"required,min=3,printascii"`
	Passphrase string `json:"password,omitempty" validate:"required,min=10"`
}

//-----------------------------------------------------------------------

type RequestUpdate struct {
	Name    string `json:"name,omitempty" validate:"required,min=3,printascii"`
	OldPass string `json:"password,omitempty" validate:"required,min=10"`
	NewPass string `json:"new_passphrase,omitempty" validate:"required,min=10"`
}

//-----------------------------------------------------------------------

type RequestSign struct {
	Name     string `json:"name,omitempty" validate:"required,min=3,printascii"`
	Password string `json:"password,omitempty" validate:"required,min=10"`

	Tx sdk.Tx `json:"tx" validate:"required"`
}

//-----------------------------------------------------------------------
