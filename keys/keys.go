package keys

import "fmt"

type SignAlgo string

const (
	AlgoEd25519   = SignAlgo("ed25519")
	AlgoSecp256k1 = SignAlgo("secp256k1")
)

func cryptoAlgoToByte(key SignAlgo) byte {
	switch key {
	case AlgoEd25519:
		return 0x01
	case AlgoSecp256k1:
		return 0x02
	default:
		panic(fmt.Sprintf("Unexpected type key %v", key))
	}
}

func byteToSignAlgo(b byte) SignAlgo {
	switch b {
	case 0x01:
		return AlgoEd25519
	case 0x02:
		return AlgoSecp256k1
	default:
		panic(fmt.Sprintf("Unexpected type byte %X", b))
	}
}
