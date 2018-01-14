package crypto

import (
	"github.com/tendermint/go-wire"
)

var cdc = wire.NewCodec()

func init() {
	// NOTE: It's important that there be no conflicts here,
	// as that would change the canonical representations,
	// and therefore change the address.
	// TODO: Add feature to go-wire to ensure that there
	// are no conflicts.
	RegisterWire(cdc)
}

func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterInterface((*PubKey)(nil), nil)
	cdc.RegisterConcrete(PubKeyEd25519{},
		"com.tendermint.wire.PubKeyEd25519", nil)
	cdc.RegisterConcrete(PubKeySecp256k1{},
		"com.tendermint.wire.PubKeySecp256k1", nil)

	cdc.RegisterInterface((*PrivKey)(nil), nil)
	cdc.RegisterConcrete(PrivKeyEd25519{},
		"com.tendermint.wire.PrivKeyEd25519", nil)
	cdc.RegisterConcrete(PrivKeySecp256k1{},
		"com.tendermint.wire.PrivKeySecp256k1", nil)

	cdc.RegisterInterface((*Signature)(nil), nil)
	cdc.RegisterConcrete(SignatureEd25519{},
		"com.tendermint.wire.SignatureKeyEd25519", nil)
	cdc.RegisterConcrete(SignatureSecp256k1{},
		"com.tendermint.wire.SignatureKeySecp256k1", nil)
}
