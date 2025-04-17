package keeper

type PublicKeyAlgorithm interface {
	// Name returns the name of the public key algorithm.
	Name() string
	// NativeAddresses derives the native addresses in each native address space for the provided public key bytes.
	NativeAddresses(keyBytes []byte) []AddressEntry
	// VerifySignature verifies a signature against a message using the public key bytes.
	VerifySignature(keyBytes, msg, sig []byte) bool
}

type AddressEntry struct {
	Type    AddressType
	Address Address
}
