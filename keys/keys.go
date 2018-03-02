package keys

import "fmt"

type CryptoAlgo string

const (
	AlgoEd25519   = CryptoAlgo("ed25519")
	AlgoSecp256k1 = CryptoAlgo("secp256k1")
)

func cryptoAlgoToByte(key CryptoAlgo) byte {
	switch key {
	case AlgoEd25519:
		return 0x01
	case AlgoSecp256k1:
		return 0x02
	default:
		panic(fmt.Sprintf("Unexpected type key %v", key))
	}
}

func byteToCryptoAlgo(b byte) CryptoAlgo {
	switch b {
	case 0x01:
		return AlgoEd25519
	case 0x02:
		return AlgoSecp256k1
	default:
		panic(fmt.Sprintf("Unexpected type byte %X", b))
	}
}
