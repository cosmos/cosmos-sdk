package keys

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"
)

func GenPrivKey(pkType string) (crypto.PrivKey, error) {
	switch pkType {
	// case "ed25519":
	// case "secp2":

	default:
		return nil, fmt.Errorf("invalid key type: %s")
	}
}
