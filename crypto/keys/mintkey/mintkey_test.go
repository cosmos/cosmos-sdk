package mintkey_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
)

func TestArmorUnarmorPrivKey(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	armor := mintkey.EncryptArmorPrivKey(priv, "passphrase")
	_, err := mintkey.UnarmorDecryptPrivKey(armor, "wrongpassphrase")
	require.Error(t, err)
	decrypted, err := mintkey.UnarmorDecryptPrivKey(armor, "passphrase")
	require.NoError(t, err)
	require.True(t, priv.Equals(decrypted))
}

func TestArmorUnarmorPubKey(t *testing.T) {
	// Select the encryption and storage for your cryptostore
	cstore := keys.NewInMemory()

	// Add keys and see they return in alphabetical order
	info, _, err := cstore.CreateMnemonic("Bob", keys.English, "passphrase", keys.Secp256k1)
	require.NoError(t, err)
	armor := mintkey.ArmorPubKeyBytes(info.GetPubKey().Bytes())
	pubBytes, err := mintkey.UnarmorPubKeyBytes(armor)
	require.NoError(t, err)
	pub, err := cryptoAmino.PubKeyFromBytes(pubBytes)
	require.NoError(t, err)
	require.True(t, pub.Equals(info.GetPubKey()))
}
