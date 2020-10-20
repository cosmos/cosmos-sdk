package types

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountResponse struct {
	Height int64    `json:"height,omitempty"`
	Result Response `json:"result,omitempty"`
}

type Response struct {
	Type  string      `json:"type,omitempty"`
	Value BaseAccount `json:"value,omitempty"`
}

type BaseAccount struct {
	Address       string    `json:"address" yaml:"address"`
	Coins         sdk.Coins `json:"coins" yaml:"coins"`
	PubKey        PublicKey `json:"public_key" yaml:"public_key"`
	AccountNumber string    `json:"account_number" yaml:"account_number"`
	Sequence      string    `json:"sequence" yaml:"sequence"`
}

type PublicKey struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

var Codec = simapp.MakeCodec()
