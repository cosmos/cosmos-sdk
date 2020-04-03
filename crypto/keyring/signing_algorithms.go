package keyring

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/algo"
)

type SignatureAlgo interface {
	Name() algo.PubKeyType
	DeriveKey() algo.DeriveKeyFn
	PrivKeyGen() algo.PrivKeyGenFn
}

func NewSigningAlgoFromString(str string) (SignatureAlgo, error) {
	if str != string(algo.Secp256k1.Name()) {
		return nil, fmt.Errorf("provided algorithm `%s` is not supported", str)
	}

	return algo.Secp256k1, nil
}

type SigningAlgoList []SignatureAlgo

func (l SigningAlgoList) Contains(algo SignatureAlgo) bool {
	for _, cAlgo := range l {
		if cAlgo.Name() == algo.Name() {
			return true
		}
	}

	return false
}
