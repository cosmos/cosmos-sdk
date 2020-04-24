package keyring

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
)

type SignatureAlgo interface {
	Name() hd.PubKeyType
	Derive() hd.DeriveFn
	Generate() hd.GenerateFn
}

func NewSigningAlgoFromString(str string) (SignatureAlgo, error) {
	if str != string(hd.Secp256k1.Name()) {
		return nil, fmt.Errorf("provided algorithm `%s` is not supported", str)
	}

	return hd.Secp256k1, nil
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
