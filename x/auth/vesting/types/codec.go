package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
)

var _ authtypes.AuthCodec = (*Codec)(nil)

type Codec struct {
	codec.Marshaler

	// Keep reference to the amino codec to allow backwards compatibility along
	// with type, and interface registration.
	amino *codec.Codec
}

func NewCodec(amino *codec.Codec) *Codec {
	return &Codec{Marshaler: codec.NewHybridCodec(amino), amino: amino}
}

// MarshalAccount marshals an AccountI interface. If the given type implements
// the Marshaler interface, it is treated as a Proto-defined message and
// serialized that way. Otherwise, it falls back on the internal Amino codec.
func (c *Codec) MarshalAccount(accI authexported.AccountI) ([]byte, error) {
	acc := &VestingAccount{}
	acc.SetAccountI(accI)
	return c.Marshaler.MarshalBinaryLengthPrefixed(acc)
}

// UnmarshalAccount returns an AccountI interface from raw encoded account bytes
// of a Proto-based Account type. An error is returned upon decoding failure.
func (c *Codec) UnmarshalAccount(bz []byte) (authexported.AccountI, error) {
	acc := &VestingAccount{}
	if err := c.Marshaler.UnmarshalBinaryLengthPrefixed(bz, acc); err != nil {
		return nil, err
	}
	return acc.GetAccountI(), nil
}

// MarshalAccountJSON JSON encodes an account object implementing the AccountI
// interface.
func (c *Codec) MarshalAccountJSON(acc authexported.AccountI) ([]byte, error) {
	return c.Marshaler.MarshalJSON(acc)
}

// UnmarshalAccountJSON returns an AccountI from JSON encoded bytes.
func (c *Codec) UnmarshalAccountJSON(bz []byte) (authexported.AccountI, error) {
	acc := &VestingAccount{}
	if err := c.Marshaler.UnmarshalJSON(bz, acc); err != nil {
		return nil, err
	}

	return acc.GetAccountI(), nil
}

// ----------------------------------------------------------------------------

// VestingCdc is the global vesting-specific codec for x/auth module.
//
// NOTE: This codec is deprecated, where a codec via NewCodec without an Amino
// codec should be used.
var VestingCdc = NewCodec(codec.New())

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	cdc.RegisterConcrete(&PeriodicVestingAccount{}, "cosmos-sdk/PeriodicVestingAccount", nil)
}

func init() {
	RegisterCodec(VestingCdc.amino)
	codec.RegisterCrypto(VestingCdc.amino)
	VestingCdc.amino.Seal()
}
