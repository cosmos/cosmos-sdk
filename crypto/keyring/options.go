package keyring

// KeyringOption overrides keyring configuratoin options.
type KeyringOption func(options *keyringOptions)

type keyringOptions struct {
	supportedAlgos       SigningAlgoList
	supportedAlgosLedger SigningAlgoList
}
