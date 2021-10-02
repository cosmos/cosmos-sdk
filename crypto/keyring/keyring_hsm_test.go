package keyring

import (
	"testing"
	"fmt"
	"log"
	"crypto/rand"

	"github.com/stretchr/testify/require"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	hsmkeys "github.com/regen-network/keystone/keys"

)

// randomBytes returns n bytes obtained from a local source of
// crypto-secure randomness. This can be used for generating
// hard-to-guess key labels, for example.
func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)

	if err != nil {
		log.Printf("Error reading random bytes: %s", err.Error())
		return nil, err
	}

	return b, nil
}

func TestCreateHsmKeyAndSign(t *testing.T) {
	cdc := getCodec()
	kb := NewInMemory(cdc)
	
	// hardcoded path for now - might change this API to actually
	// take a JSON string and let caller decide how to get that
	// JSON
	kr, err := hsmkeys.NewPkcs11FromConfig("./testdata/keys/pkcs11-config")

	label, err := randomBytes(16)
	require.NoError(t, err)

	key, err := kr.NewKey(hsmkeys.KEYGEN_SECP256K1, string(label))
	require.NoError(t, err)
	require.NotNil(t, key)

	require.NoError(t, err)

	hsmrecord, err := kb.SaveHsmKey("testkey123", hd.Secp256k1, string(label), kr)
	require.NoError(t, err)

	hsmkey := hsmrecord.GetHsm()
	labelcopy := hsmkey.GetLabel()
	fmt.Printf("Key label: %s", labelcopy)
	pubkey, err := hsmrecord.GetPubKey()
	require.NoError(t, err)
	fmt.Printf("PubKey: %v", pubkey )
	msg := []byte("Signing this plaintext tells me what exactly?")
	sig, pub, err := SignWithHsm(hsmrecord, msg, kr)
	require.NoError(t, err)
	fmt.Printf("Signature: %v", sig)
	fmt.Printf("PubKey: %v", pub)
}
