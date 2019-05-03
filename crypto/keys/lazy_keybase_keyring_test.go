package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/99designs/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

)

func TestNewKeybaseKeyringFileOnly(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	kb := NewKeybaseKeyringFileOnly("keybasename", dir)
	lazykb, ok := kb.(lazyKeybaseKeyring)
	require.True(t, ok)
	require.Equal(t, lazykb.name, "keybasename")
}


// New creates a new instance of a lazy keybase.
func NewKeybaseKeyringFileOnly(name string, dir string) Keybase {

	_, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{"file"},
		//Keyring with encrypted application data
		ServiceName: name,
		FileDir:dir,
	})
	if err != nil{
		panic(err)
	}

	return lazyKeybaseKeyring{name: name, dir:dir}
}


func TestLazyKeyManagementKeyRing(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	kb := NewKeybaseKeyringFileOnly("keybasename", dir)

	algo := Secp256k1
	n1, n2, n3 := "personal", "business", "other"
	p1, p2 := "1234", "really-secure!@#$"

	// Check empty state
	l, err := kb.List()
	require.Nil(t, err)
	assert.Empty(t, l)

	_, _, err = kb.CreateMnemonic(n1, English, p1, Ed25519)
	require.Error(t, err, "ed25519 keys are currently not supported by keybase")

	// create some keys
	_, err = kb.Get(n1)
	require.Error(t, err)
	i, _, err := kb.CreateMnemonic(n1, English, p1, algo)

	require.NoError(t, err)
	require.Equal(t, n1, i.GetName())
	_, _, err = kb.CreateMnemonic(n2, English, p2, algo)
	require.NoError(t, err)

	// we can get these keys
	i2, err := kb.Get(n2)
	require.NoError(t, err)
	_, err = kb.Get(n3)
	require.NotNil(t, err)
	_, err = kb.GetByAddress(accAddr(i2))
	require.NoError(t, err)
	addr, err := sdk.AccAddressFromBech32("cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t")
	require.NoError(t, err)
	_, err = kb.GetByAddress(addr)
	require.NotNil(t, err)

	// list shows them in order
	keyS, err := kb.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keyS))
	// note these are in alphabetical order
	require.Equal(t, n2, keyS[0].GetName())
	require.Equal(t, n1, keyS[1].GetName())
	require.Equal(t, i2.GetPubKey(), keyS[0].GetPubKey())

	// deleting a key removes it
	err = kb.Delete("bad name", "foo", false)
	require.NotNil(t, err)
	err = kb.Delete(n1, p1, false)
	require.NoError(t, err)
	keyS, err = kb.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(keyS))
	_, err = kb.Get(n1)
	require.Error(t, err)

	// create an offline key
	o1 := "offline"
	priv1 := ed25519.GenPrivKey()
	pub1 := priv1.PubKey()
	i, err = kb.CreateOffline(o1, pub1)
	require.Nil(t, err)
	require.Equal(t, pub1, i.GetPubKey())
	require.Equal(t, o1, i.GetName())
	keyS, err = kb.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keyS))

	// delete the offline key
	err = kb.Delete(o1, "", false)
	require.NoError(t, err)
	keyS, err = kb.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(keyS))

	// addr cache gets nuked - and test skip flag
	err = kb.Delete(n2, "", true)
	require.NoError(t, err)
}