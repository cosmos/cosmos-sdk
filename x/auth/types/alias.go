package types

import (
	proto "github.com/cosmos/gogoproto/proto"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountAliasI interface {
	proto.Message

	GetAddress() sdk.AccAddress
	SetAddress(sdk.AccAddress) error // errors if already set.

	GetPubKey() cryptotypes.PubKey // can return nil.
	SetPubKey(cryptotypes.PubKey) error

	GetAccountNumber() uint64
	SetAccountNumber(uint64) error

	GetSequence() uint64
	SetSequence(uint64) error

	// Ensure that account implements stringer
	String() string
}
