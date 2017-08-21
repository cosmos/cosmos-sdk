package rest

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/tendermint/go-crypto/keys"
)

type CreateKeyRequest struct {
	Name       string `json:"name,omitempty" validate:"required,min=3,printascii"`
	Passphrase string `json:"password,omitempty" validate:"required,min=10"`

	// Algo is the requested algorithm to create the key
	Algo string `json:"algo,omitempty"`
}

type DeleteKeyRequest struct {
	Name       string `json:"name,omitempty" validate:"required,min=3,printascii"`
	Passphrase string `json:"password,omitempty" validate:"required,min=10"`
}

type UpdateKeyRequest struct {
	Name    string `json:"name,omitempty" validate:"required,min=3,printascii"`
	OldPass string `json:"password,omitempty" validate:"required,min=10"`
	NewPass string `json:"new_passphrase,omitempty" validate:"required,min=10"`
}

type SignRequest struct {
	Name     string `json:"name,omitempty" validate:"required,min=3,printascii"`
	Password string `json:"password,omitempty" validate:"required,min=10"`

	Tx sdk.Tx `json:"tx" validate:"required"`
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

	To     *sdk.Actor `json:"to"`
	From   *sdk.Actor `json:"from"`
	Amount coin.Coins      `json:"amount"`
}
