package hd

import (
	"testing"

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/stretchr/testify/require"
)

const (
	mnemonic    = "picnic rent average infant boat squirrel federal assault mercy purity very motor fossil wheel verify upset box fresh horse vivid copy predict square regret"
	BIP44HDPath = "m/44'/60'/0'/0/0"
)

func TestDerivation(t *testing.T) {
	bz, err := EthSecp256k1.Derive()(mnemonic, "", BIP44HDPath)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	badBz, err := EthSecp256k1.Derive()(mnemonic, "", "44'/118'/0'/0/0")
	require.NoError(t, err)
	require.NotEmpty(t, badBz)

	require.NotEqual(t, bz, badBz)

	privkey := EthSecp256k1.Generate()(bz)
	badPrivKey := EthSecp256k1.Generate()(badBz)

	require.False(t, privkey.Equals(badPrivKey))

	// require.Equal(t, account.Address.String(), "0xA588C66983a81e800Db4dF74564F09f91c026351")
	require.Equal(t, common.BytesToAddress(privkey.PubKey().Address().Bytes()).String(), "0xA588C66983a81e800Db4dF74564F09f91c026351")
	require.NotEqual(t, common.BytesToAddress(badPrivKey.PubKey().Address().Bytes()).String(), "0xA588C66983a81e800Db4dF74564F09f91c026351")

}

func TestDefaults(t *testing.T) {
	require.Equal(t, PubKeyType("multi"), MultiType)
	require.Equal(t, PubKeyType("secp256k1"), Secp256k1Type)
	require.Equal(t, PubKeyType("ethsecp256k1"), EthSecp256k1Type)
	require.Equal(t, PubKeyType("ed25519"), Ed25519Type)
	require.Equal(t, PubKeyType("sr25519"), Sr25519Type)
}
