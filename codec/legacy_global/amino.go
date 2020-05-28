package legacy_global

import (
	"github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/multisig"
)

// Cdc defines a global generic sealed Amino codec to be used throughout sdk. It
// has all Tendermint crypto and evidence types registered.
//
// TODO: Consider removing this global.
var Cdc *codec.Codec

func init() {
	Cdc = codec.New()
	RegisterCrypto(Cdc)
	RegisterEvidences(Cdc)
	Cdc.Seal()
}

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(cdc *codec.Codec) {
	multisig.RegisterAmino(cdc.Amino)
}

// RegisterEvidences registers Tendermint evidence types with the provided Amino
// codec.
func RegisterEvidences(cdc *codec.Codec) {
	types.RegisterEvidences(cdc.Amino)
}
