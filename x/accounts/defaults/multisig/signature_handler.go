package multisig

// SignatureHandler allows custom signatures, it must be able to produce sign bytes
// and also verify the received signatures. It does NOT produce signatures as this
// is done off-chain.
// Note: implementers will probably want to have also a signing method in this handler
// so they can use it in the client.
type SignatureHandler interface {
	VerifySignature(signBytes []byte, pubkeys [][]byte) error

	// ValidatePubKey must error if the provided key does not comply with the
	// established format.
	ValidatePubKey([]byte) error

	// not sure if needed
	RecoverPubKey([]byte) ([]byte, error)

	GetSignBytes(msgs []byte, pubkeys [][]byte) error
}
