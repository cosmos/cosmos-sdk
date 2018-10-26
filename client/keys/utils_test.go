package keys

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestGetKeyBaseLocks(t *testing.T) {
	dir, err := ioutil.TempDir("", "cosmos-sdk-keys")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	// Acquire db
	kb, err := GetKeyBaseFromDirWithWritePerm(dir)
	require.Nil(t, err)
	_, _, err = kb.CreateMnemonic("foo", keys.English, "12345678", keys.Secp256k1)
	require.Nil(t, err)
	// Reset global variable
	keybase = nil
	// Try to acquire another keybase from the same storage
	_, err = GetKeyBaseFromDirWithWritePerm(dir)
	require.NotNil(t, err)
	_, err = GetKeyBaseFromDirWithWritePerm(dir)
	require.NotNil(t, err)

	// Close the db and try to acquire the lock
	kb.CloseDB()
	kb, err = GetKeyBaseFromDirWithWritePerm(dir)
	require.Nil(t, err)

	// Try to acquire another read-only keybase from the same storage
	_, err = GetKeyBaseFromDir(dir)
	require.Nil(t, err)

	kb.CloseDB()
}
