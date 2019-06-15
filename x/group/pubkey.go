package group

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

type PubKey struct {
	addr  sdk.AccAddress
}

func (key PubKey) Address() crypto.Address {
	return crypto.Address(key.addr)
}

func (key PubKey) Bytes() []byte {
	panic("not to be serialized")
}

func (key PubKey) VerifyBytes(msg []byte, sig []byte) bool {
	panic("can't verify signatures with a group PubKey")
}

func (key PubKey) Equals(crypto.PubKey) bool {
	panic("not to be compared")
}
