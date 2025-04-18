package keeper

import (
	"crypto/sha256"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type EvmAddressSpace struct{}

func (e EvmAddressSpace) Name() string {
	return "evm"
}

func (e EvmAddressSpace) DeriveAddress(id AccountID, pubKey cryptotypes.PubKey) Address {
	switch pubKey.(type) {
	case *secp256k1.PubKey:
		// do EVM address derivation
		panic("keccak-256(pubKey.Bytes())") // need go-ethereum dep
	default:
		// default to taking the SHA-256 of the account ID prefixed with the string "acc"
		hash := sha256.New()
		hash.Write([]byte("acc"))
		hash.Write(id[:])
		addr := hash.Sum(nil)
		return addr[:20]
	}
}

var _ AddressSpaceManager = EvmAddressSpace{}
