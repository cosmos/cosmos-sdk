package secp256r1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"io"
	"math/big"

	"github.com/tendermint/tendermint/crypto"
)

var pubKeyCurve = elliptic.P256()

const PrivKeyNistp256Size = 256

// PrivKeyNistp256 implements crypto.PrivKey.
type PrivKeyNistp256 [PrivKeyNistp256Size]byte

func (privKey PrivKeyNistp256) PubKey() PubKeyNistp256 {
	_, publicKey := PrivKeyFromBytes(pubKeyCurve, privKey[:])
	//TODO implement pubkey
	return publicKey
}

func (privKey PrivKeyNistp256) Sign(msg []byte) (sign []byte, err error) {
	privateKey, _ := PrivKeyFromBytes(pubKeyCurve, privKey[:])
	//TODO add sign params(decide on random generation algo to be used)
	return privateKey.Sign()
}

func (privKey PrivKeyNistp256) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(privKey)
}

func genPrivKey(rand io.Reader) (PrivKeyNistp256, error) {
	privatekey, err := ecdsa.GenerateKey(pubKeyCurve, rand) // this generates a public & private key pair
	if err != nil {
		fmt.Println(err)
		return [256]byte{}, err
	}
	x := new(big.Int)
	x = privatekey.D
	y := x.Bytes()

	var z PrivKeyNistp256
	copy(z[:], y)
	return z, nil
}
