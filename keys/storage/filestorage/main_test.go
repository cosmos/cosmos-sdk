package filestorage

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	crypto "github.com/tendermint/go-crypto"
	keys "github.com/tendermint/go-crypto/keys"
)

func TestBasicCRUD(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	dir, err := ioutil.TempDir("", "filestorage-test")
	assert.Nil(err)
	defer os.RemoveAll(dir)
	store := New(dir)

	name := "bar"
	key := []byte("secret-key-here")
	pubkey := crypto.GenPrivKeyEd25519().PubKey()
	info := keys.Info{
		Name:   name,
		PubKey: pubkey.Wrap(),
	}

	// No data: Get and Delete return nothing
	_, _, err = store.Get(name)
	assert.NotNil(err)
	err = store.Delete(name)
	assert.NotNil(err)
	// List returns empty list
	l, err := store.List()
	assert.Nil(err)
	assert.Empty(l)

	// Putting the key in the  store must work
	err = store.Put(name, key, info)
	assert.Nil(err)
	// But a second time is a failure
	err = store.Put(name, key, info)
	assert.NotNil(err)

	// Now, we can get and list properly
	k, i, err := store.Get(name)
	require.Nil(err, "%+v", err)
	assert.Equal(key, k)
	assert.Equal(info.Name, i.Name)
	assert.Equal(info.PubKey, i.PubKey)
	assert.NotEmpty(i.Address)
	l, err = store.List()
	require.Nil(err, "%+v", err)
	assert.Equal(1, len(l))
	assert.Equal(i, l[0])

	// querying a non-existent key fails
	_, _, err = store.Get("badname")
	assert.NotNil(err)

	// We can only delete once
	err = store.Delete(name)
	assert.Nil(err)
	err = store.Delete(name)
	assert.NotNil(err)

	// and then Get and List don't work
	_, _, err = store.Get(name)
	assert.NotNil(err)
	// List returns empty list
	l, err = store.List()
	assert.Nil(err)
	assert.Empty(l)
}

func TestDirectoryHandling(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// prepare a temp dir and make sure it is not there
	newDir := path.Join(os.TempDir(), "file-test-dir")
	_, err := os.Open(newDir)
	assert.True(os.IsNotExist(err))
	defer os.RemoveAll(newDir)

	// now, check with two levels deep....
	parentDir := path.Join(os.TempDir(), "missing-dir")
	nestedDir := path.Join(parentDir, "lots", "of", "levels", "here")
	_, err = os.Open(parentDir)
	assert.True(os.IsNotExist(err))
	defer os.RemoveAll(parentDir)

	// create a new storage, and verify it creates the directory with good permissions
	for _, dir := range []string{newDir, nestedDir, newDir} {
		New(dir)
		d, err := os.Open(dir)
		require.Nil(err)
		defer d.Close()

		stat, err := d.Stat()
		require.Nil(err)
		assert.Equal(dirPerm, stat.Mode()&os.ModePerm)
	}
}
