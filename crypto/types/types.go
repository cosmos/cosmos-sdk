package types

import (
	"github.com/tendermint/tendermint/crypto"
)

// PublicKey specifies a public key
type PublicKey struct {
	// sum specifies which type of public key is wrapped
	//
	// Types that are valid to be assigned to Sum:
	//	*PublicKey_Secp256K1
	//	*PublicKey_Ed25519
	//	*PublicKey_Sr25519
	//	*PublicKey_Multisig
	//	*PublicKey_Secp256R1
	//	*PublicKey_AminoMultisig
	//	*PublicKey_AnyPubkey
	Sum isPublicKey_Sum `protobuf_oneof:"sum"`

	cachedValue crypto.PubKey
}

// GetCachedPubKey returns the cached PubKey instance wrapped in the PublicKey.
// This will only be set if the PublicKeyCodec is cache-wrapped using CacheWrapCodec
func (pk PublicKey) GetCachedPubKey() crypto.PubKey {
	return pk.cachedValue
}
