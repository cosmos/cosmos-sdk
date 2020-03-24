package keyring

// KeybaseOption overrides options for the db
type KeybaseOption func(*kbOptions)

type kbOptions struct {
	keygenFunc           PrivKeyGenFunc
	deriveFunc           DeriveKeyFunc
	supportedAlgos       []SigningAlgo
	supportedAlgosLedger []SigningAlgo
}

// WithKeygenFunc applies an overridden key generation function to generate the private key.
func WithKeygenFunc(f PrivKeyGenFunc) KeybaseOption {
	return func(o *kbOptions) {
		o.keygenFunc = f
	}
}

// WithDeriveFunc applies an overridden key derivation function to generate the private key.
func WithDeriveFunc(f DeriveKeyFunc) KeybaseOption {
	return func(o *kbOptions) {
		o.deriveFunc = f
	}
}

// WithSupportedAlgos defines the list of accepted SigningAlgos.
func WithSupportedAlgos(algos []SigningAlgo) KeybaseOption {
	return func(o *kbOptions) {
		o.supportedAlgos = algos
	}
}

// WithSupportedAlgosLedger defines the list of accepted SigningAlgos compatible with Ledger.
func WithSupportedAlgosLedger(algos []SigningAlgo) KeybaseOption {
	return func(o *kbOptions) {
		o.supportedAlgosLedger = algos
	}
}
