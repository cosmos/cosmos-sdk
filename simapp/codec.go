package simapp

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// AppCodec defines the application-level codec. This codec contains all the
// required module-specific codecs that are to be provided upon initialization.
type AppCodec struct {
	amino *codec.Codec

	Staking      *staking.Codec
	Distribution *distr.Codec
}

func NewAppCodec() *AppCodec {
	amino := MakeCodec()

	return &AppCodec{
		amino:        amino,
		Staking:      staking.NewCodec(amino),
		Distribution: distr.NewCodec(amino),
	}
}

// MakeCodec creates and returns a reference to an Amino codec that has all the
// necessary types and interfaces registered. This codec is provided to all the
// modules the application depends on.
//
// NOTE: This codec will be deprecated in favor of AppCodec once all modules are
// migrated.
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc)
	vesting.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}
