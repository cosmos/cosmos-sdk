package keys

import (
	"encoding/hex"
	"fmt"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)


func encode(derivedPriv [32]byte) string {
	src := make([]byte, len(derivedPriv))
	for idx, m := range derivedPriv {
		src[idx] = m
	}
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return string(dst)
}

func decode(key string) [32]byte {
	src := []byte(key)
	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		panic(err)
	}

	if n != 32 {
		panic("invalid input!")
	}

	var res [32]byte
	for idx, m := range dst {
		res[idx] = m
	}

	return res
}

func deriveKeyByPrivKey(privKey string, algo SigningAlgo) ([]byte, error) {
	switch algo {
	case Secp256k1:
		decodePriv := decode(privKey)
		keyStr := encode(decodePriv)
		if privKey != keyStr {
			return nil, fmt.Errorf("invalid private key '%s', algo '%s'", privKey, algo)
		}
		return decodePriv[:], nil
	case SigningAlgo("eth_secp256k1"):
		privKeyECDSA, err := ethcrypto.HexToECDSA(privKey)
		if err != nil {
			return nil, fmt.Errorf("invalid private key '%s', algo '%s', error: %s", privKey, algo, err)
		}

		return ethcrypto.FromECDSA(privKeyECDSA), nil
	default:
		return nil, errors.Wrap(ErrUnsupportedSigningAlgo, string(algo))
	}
}
