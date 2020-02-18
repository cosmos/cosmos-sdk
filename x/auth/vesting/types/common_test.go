package types_test

import (
	"github.com/cosmos/cosmos-sdk/codec"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	_ authtypes.Codec = (*Codec)(nil)

	vestingCdc = NewCodec(codec.New())
)

type Codec struct {
	codec.Marshaler

	// Keep reference to the amino codec to allow backwards compatibility along
	// with type, and interface registration.
	amino *codec.Codec
}

func NewCodec(amino *codec.Codec) *Codec {
	return &Codec{Marshaler: codec.NewHybridCodec(amino), amino: amino}
}

// MarshalAccount marshals an Account interface. If the given type implements
// the Marshaler interface, it is treated as a Proto-defined message and
// serialized that way. Otherwise, it falls back on the internal Amino codec.
func (c *Codec) MarshalAccount(accI authexported.Account) ([]byte, error) {
	acc := &types.TestVestingAccount{}
	acc.SetAccount(accI)
	return c.Marshaler.MarshalBinaryLengthPrefixed(acc)
}

// UnmarshalAccount returns an Account interface from raw encoded account bytes
// of a Proto-based Account type. An error is returned upon decoding failure.
func (c *Codec) UnmarshalAccount(bz []byte) (authexported.Account, error) {
	acc := &types.TestVestingAccount{}
	if err := c.Marshaler.UnmarshalBinaryLengthPrefixed(bz, acc); err != nil {
		return nil, err
	}
	return acc.GetAccount(), nil
}

// MarshalAccountJSON JSON encodes an account object implementing the Account
// interface.
func (c *Codec) MarshalAccountJSON(acc authexported.Account) ([]byte, error) {
	return c.Marshaler.MarshalJSON(acc)
}

// UnmarshalAccountJSON returns an Account from JSON encoded bytes.
func (c *Codec) UnmarshalAccountJSON(bz []byte) (authexported.Account, error) {
	acc := &types.TestVestingAccount{}
	if err := c.Marshaler.UnmarshalJSON(bz, acc); err != nil {
		return nil, err
	}

	return acc.GetAccount(), nil
}
