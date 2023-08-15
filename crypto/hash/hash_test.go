package hash

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/assert"
	"testing"
)

var defaultBasicAddress = []byte{112, 20, 2, 255, 164, 54, 124, 4, 155, 99, 105, 102, 113, 23, 207, 125, 181, 74, 174, 52}

func TestAddressBasic(t *testing.T) {
	privKey := &secp256k1.PrivKey{Key: nil}
	pubKey := privKey.PubKey()
	address := AddressBasic(pubKey)
	assert.NotNil(t, address)
	assert.True(t, len(address) == 20)
	assert.Equal(t, address, defaultBasicAddress)
}
