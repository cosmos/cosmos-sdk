package cryptostore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-keys/cryptostore"
	"github.com/tendermint/go-keys/storage/memstorage"
)

// TestKeyManagement makes sure we can manipulate these keys well
func TestKeyManagement(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// make the storage with reasonable defaults
	cstore := cryptostore.New(
		cryptostore.GenSecp256k1,
		cryptostore.SecretBox,
		memstorage.New(),
	)

	n1, n2, n3 := "personal", "business", "other"
	p1, p2 := "1234", "really-secure!@#$"

	// Check empty state
	l, err := cstore.List()
	require.Nil(err)
	assert.Empty(l)

	// create some keys
	_, err = cstore.Get(n1)
	assert.NotNil(err)
	err = cstore.Create(n1, p1)
	require.Nil(err)
	err = cstore.Create(n2, p2)
	require.Nil(err)

	// we can get these keys
	i2, err := cstore.Get(n2)
	assert.Nil(err)
	_, err = cstore.Get(n3)
	assert.NotNil(err)

	// list shows them in order
	keys, err := cstore.List()
	require.Nil(err)
	require.Equal(2, len(keys))
	// note these are in alphabetical order
	assert.Equal(n2, keys[0].Name)
	assert.Equal(n1, keys[1].Name)
	assert.Equal(i2.PubKey, keys[0].PubKey)

	// deleting a key removes it
	err = cstore.Delete("bad name", "foo")
	require.NotNil(err)
	err = cstore.Delete(n1, p1)
	require.Nil(err)
	keys, err = cstore.List()
	require.Nil(err)
	assert.Equal(1, len(keys))
	_, err = cstore.Get(n1)
	assert.NotNil(err)

	// make sure that it only signs with the right password
	// tx := mock.NewSig([]byte("mytransactiondata"))
	// err = cstore.Sign(n2, p1, tx)
	// assert.NotNil(err)
	// err = cstore.Sign(n2, p2, tx)
	// assert.Nil(err, "%+v", err)
	// sigs, err := tx.Signers()
	// assert.Nil(err, "%+v", err)
	// if assert.Equal(1, len(sigs)) {
	// 	assert.Equal(i2.PubKey, sigs[0])
	// }
}

// TestSignVerify does some detailed checks on how we sign and validate
// signatures
// func TestSignVerify(t *testing.T) {
// 	assert, require := assert.New(t), require.New(t)

// 	// make the storage with reasonable defaults
// 	cstore := cryptostore.New(
// 		cryptostore.GenSecp256k1,
// 		cryptostore.SecretBox,
// 		memstorage.New(),
// 	)

// 	n1, n2 := "some dude", "a dudette"
// 	p1, p2 := "1234", "foobar"

// 	// create two users and get their info
// 	err := cstore.Create(n1, p1)
// 	require.Nil(err)
// 	i1, err := cstore.Get(n1)
// 	require.Nil(err)

// 	err = cstore.Create(n2, p2)
// 	require.Nil(err)
// 	i2, err := cstore.Get(n2)
// 	require.Nil(err)

// 	// let's try to sign some messages
// 	d1 := []byte("my first message")
// 	d2 := []byte("some other important info!")

// 	// try signing both data with both keys...
// 	s11, err := cstore.Signature(n1, p1, d1)
// 	require.Nil(err)
// 	s12, err := cstore.Signature(n1, p1, d2)
// 	require.Nil(err)
// 	s21, err := cstore.Signature(n2, p2, d1)
// 	require.Nil(err)
// 	s22, err := cstore.Signature(n2, p2, d2)
// 	require.Nil(err)

// 	// let's try to validate and make sure it only works when everything is proper
// 	keys := [][]byte{i1.PubKey, i2.PubKey}
// 	data := [][]byte{d1, d2}
// 	sigs := [][]byte{s11, s12, s21, s22}

// 	// loop over keys and data
// 	for k := 0; k < 2; k++ {
// 		for d := 0; d < 2; d++ {
// 			// make sure only the proper sig works
// 			good := 2*k + d
// 			for s := 0; s < 4; s++ {
// 				err = cstore.Verify(data[d], sigs[s], keys[k])
// 				if s == good {
// 					assert.Nil(err, "%+v", err)
// 				} else {
// 					assert.NotNil(err)
// 				}
// 			}
// 		}
// 	}
// }

func assertPassword(assert *assert.Assertions, cstore cryptostore.Manager, name, pass, badpass string) {
	err := cstore.Update(name, badpass, pass)
	assert.NotNil(err)
	err = cstore.Update(name, pass, pass)
	assert.Nil(err, "%+v", err)
}

// TestAdvancedKeyManagement verifies update, import, export functionality
func TestAdvancedKeyManagement(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// make the storage with reasonable defaults
	cstore := cryptostore.New(
		cryptostore.GenSecp256k1,
		cryptostore.SecretBox,
		memstorage.New(),
	)

	n1, n2 := "old-name", "new name"
	p1, p2, p3, pt := "1234", "foobar", "ding booms!", "really-secure!@#$"

	// make sure key works with initial password
	err := cstore.Create(n1, p1)
	require.Nil(err, "%+v", err)
	assertPassword(assert, cstore, n1, p1, p2)

	// update password requires the existing password
	err = cstore.Update(n1, "jkkgkg", p2)
	assert.NotNil(err)
	assertPassword(assert, cstore, n1, p1, p2)

	// then it changes the password when correct
	err = cstore.Update(n1, p1, p2)
	assert.Nil(err)
	// p2 is now the proper one!
	assertPassword(assert, cstore, n1, p2, p1)

	// exporting requires the proper name and passphrase
	_, err = cstore.Export(n2, p2, pt)
	assert.NotNil(err)
	_, err = cstore.Export(n1, p1, pt)
	assert.NotNil(err)
	exported, err := cstore.Export(n1, p2, pt)
	require.Nil(err, "%+v", err)

	// import fails on bad transfer pass
	err = cstore.Import(n2, p3, p2, exported)
	assert.NotNil(err)
	// import cannot overwrite existing keys
	err = cstore.Import(n1, p3, pt, exported)
	assert.NotNil(err)
	// we can now import under another name
	err = cstore.Import(n2, p3, pt, exported)
	require.Nil(err, "%+v", err)

	// make sure both passwords are now properly set (not to the transfer pass)
	assertPassword(assert, cstore, n1, p2, pt)
	assertPassword(assert, cstore, n2, p3, pt)
}

// func ExampleStore() {
// 	// Select the encryption and storage for your cryptostore
// 	cstore := cryptostore.New(
// 		cryptostore.GenEd25519,
// 		cryptostore.SecretBox,
// 		// Note: use filestorage.New(dir) for real data
// 		memstorage.New(),
// 	)

// 	// Add keys and see they return in alphabetical order
// 	cstore.Create("Bob", "friend")
// 	cstore.Create("Alice", "secret")
// 	cstore.Create("Carl", "mitm")
// 	info, _ := cstore.List()
// 	for _, i := range info {
// 		fmt.Println(i.Name)
// 	}

// 	// We need to use passphrase to generate a signature
// 	tx := mock.NewSig([]byte("deadbeef"))
// 	err := cstore.Sign("Bob", "friend", tx)
// 	if err != nil {
// 		fmt.Println("don't accept real passphrase")
// 	}

// 	// and we can validate the signature with publically available info
// 	binfo, _ := cstore.Get("Bob")
// 	sigs, err := tx.Signers()
// 	if err != nil {
// 		fmt.Println("badly signed")
// 	} else if bytes.Equal(sigs[0].Bytes(), binfo.PubKey.Bytes()) {
// 		fmt.Println("signed by Bob")
// 	} else {
// 		fmt.Println("signed by someone else")
// 	}

// 	// Output:
// 	// Alice
// 	// Bob
// 	// Carl
// 	// signed by Bob
// }
