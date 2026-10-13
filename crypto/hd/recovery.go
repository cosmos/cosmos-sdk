package hd

// AccountKeyAlgo derives and generates an account key from a mnemonic.
// It matches keyring.SignatureAlgo.
type AccountKeyAlgo interface {
	Name() PubKeyType
	Derive() DeriveFn
	Generate() GenerateFn
}

// AccountKeyDerivation is one way a mnemonic may have produced the account key
// at a given coin type, account, and index.
type AccountKeyDerivation struct {
	// Scheme is "pqhd-<scheme id>" for PQHD paths and "bip44" for BIP44 paths.
	Scheme     string
	PubKeyType PubKeyType
	Path       string
	Algo       AccountKeyAlgo
}

// AccountKeyDerivations returns every account key derivation this SDK supports
// for the given coin type, account, and index. A wallet restoring from a
// mnemonic must try each of them, because neither a stored key nor an address
// records which derivation was used.
//
// PQHD derivations come first: each scheme has its own path, so a PQHD
// candidate is unambiguous. The BIP44 path is shared by secp256k1 and legacy
// ML-DSA-65 keys, so it appears once for each algorithm.
func AccountKeyDerivations(coinType, account, index uint32) ([]AccountKeyDerivation, error) {
	pqhdPath, err := CreateMlDsa65BIP85Path(coinType, account, index)
	if err != nil {
		return nil, err
	}
	bip44Path := CreateHDPath(coinType, account, index).String()
	return []AccountKeyDerivation{
		{Scheme: "pqhd-201", PubKeyType: MlDsa65Type, Path: pqhdPath, Algo: MlDsa65BIP85},
		{Scheme: "bip44", PubKeyType: Secp256k1Type, Path: bip44Path, Algo: Secp256k1},
		{Scheme: "bip44", PubKeyType: MlDsa65Type, Path: bip44Path, Algo: MlDsa65},
	}, nil
}
