package crypto

import (
	"fmt"

	"crypto/subtle"

	. "github.com/tendermint/tmlibs/common"
)

func SignatureFromBytes(pubKeyBytes []byte) (pubKey Signature, err error) {
	err = cdc.UnmarshalBinaryBare(pubKeyBytes, &pubKey)
	return
}

//----------------------------------------

type Signature interface {
	Bytes() []byte
	IsZero() bool
	Equals(Signature) bool
}

//-------------------------------------

var _ Signature = SignatureEd25519{}

// Implements Signature
type SignatureEd25519 [64]byte

func (sig SignatureEd25519) Bytes() []byte {
	bz, err := cdc.MarshalBinaryBare(sig)
	if err != nil {
		panic(err)
	}
	return bz
}

func (sig SignatureEd25519) IsZero() bool { return len(sig) == 0 }

func (sig SignatureEd25519) String() string { return fmt.Sprintf("/%X.../", Fingerprint(sig[:])) }

func (sig SignatureEd25519) Equals(other Signature) bool {
	if otherEd, ok := other.(SignatureEd25519); ok {
		return subtle.ConstantTimeCompare(sig[:], otherEd[:]) == 1
	} else {
		return false
	}
}

func SignatureEd25519FromBytes(data []byte) Signature {
	var sig SignatureEd25519
	copy(sig[:], data)
	return sig
}

//-------------------------------------

var _ Signature = SignatureSecp256k1{}

// Implements Signature
type SignatureSecp256k1 []byte

func (sig SignatureSecp256k1) Bytes() []byte {
	bz, err := cdc.MarshalBinaryBare(sig)
	if err != nil {
		panic(err)
	}
	return bz
}

func (sig SignatureSecp256k1) IsZero() bool { return len(sig) == 0 }

func (sig SignatureSecp256k1) String() string { return fmt.Sprintf("/%X.../", Fingerprint(sig[:])) }

func (sig SignatureSecp256k1) Equals(other Signature) bool {
	if otherSecp, ok := other.(SignatureSecp256k1); ok {
		return subtle.ConstantTimeCompare(sig[:], otherSecp[:]) == 1
	} else {
		return false
	}
}

func SignatureSecp256k1FromBytes(data []byte) Signature {
	sig := make(SignatureSecp256k1, len(data))
	copy(sig[:], data)
	return sig
}
