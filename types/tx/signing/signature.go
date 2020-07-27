package signing

import "github.com/tendermint/tendermint/crypto"

// SignatureV2 is a convenience type that is easier to use in application logic
// than the protobuf SignerInfo's and raw signature bytes. It goes beyond the
// first sdk.Signature types by supporting sign modes and explicitly nested
// multi-signatures. It is intended to be used for both building and verifying
// signatures.
type SignatureV2 struct {
	// PubKey is the public key to use for verifying the signature
	PubKey crypto.PubKey

	// Data is the actual data of the signature which includes SignMode's and
	// the signatures themselves for either single or multi-signatures.
	Data SignatureData
}
