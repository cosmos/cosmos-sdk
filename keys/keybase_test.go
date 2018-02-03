package keys_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tmlibs/db"

	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/words"
)

// TestKeyManagement makes sure we can manipulate these keys well
func TestKeyManagement(t *testing.T) {

	// make the storage with reasonable defaults
	cstore := keys.New(
		dbm.NewMemDB(),
		words.MustLoadCodec("english"),
	)

	algo := keys.AlgoEd25519
	n1, n2, n3 := "personal", "business", "other"
	p1, p2 := "1234", "really-secure!@#$"

	// Check empty state
	l, err := cstore.List()
	require.Nil(t, err)
	assert.Empty(t, l)

	// create some keys
	_, err = cstore.Get(n1)
	assert.NotNil(t, err)
	i, _, err := cstore.Create(n1, p1, algo)
	require.Equal(t, n1, i.Name)
	require.Nil(t, err)
	_, _, err = cstore.Create(n2, p2, algo)
	require.Nil(t, err)

	// we can get these keys
	i2, err := cstore.Get(n2)
	assert.Nil(t, err)
	_, err = cstore.Get(n3)
	assert.NotNil(t, err)

	// list shows them in order
	keyS, err := cstore.List()
	require.Nil(t, err)
	require.Equal(t, 2, len(keyS))
	// note these are in alphabetical order
	assert.Equal(t, n2, keyS[0].Name)
	assert.Equal(t, n1, keyS[1].Name)
	assert.Equal(t, i2.PubKey, keyS[0].PubKey)

	// deleting a key removes it
	err = cstore.Delete("bad name", "foo")
	require.NotNil(t, err)
	err = cstore.Delete(n1, p1)
	require.Nil(t, err)
	keyS, err = cstore.List()
	require.Nil(t, err)
	assert.Equal(t, 1, len(keyS))
	_, err = cstore.Get(n1)
	assert.NotNil(t, err)

	// make sure that it only signs with the right password
	// tx := mock.NewSig([]byte("mytransactiondata"))
	// err = cstore.Sign(n2, p1, tx)
	// assert.NotNil(t, err)
	// err = cstore.Sign(n2, p2, tx)
	// assert.Nil(t, err, "%+v", err)
	// sigs, err := tx.Signers()
	// assert.Nil(t, err, "%+v", err)
	// if assert.Equal(t, 1, len(sigs)) {
	// 	assert.Equal(t, i2.PubKey, sigs[0])
	// }
}

// TestSignVerify does some detailed checks on how we sign and validate
// signatures
func TestSignVerify(t *testing.T) {

	// make the storage with reasonable defaults
	cstore := keys.New(
		dbm.NewMemDB(),
		words.MustLoadCodec("english"),
	)
	algo := keys.AlgoSecp256k1

	n1, n2 := "some dude", "a dudette"
	p1, p2 := "1234", "foobar"

	// create two users and get their info
	i1, _, err := cstore.Create(n1, p1, algo)
	require.Nil(t, err)

	i2, _, err := cstore.Create(n2, p2, algo)
	require.Nil(t, err)

	// let's try to sign some messages
	d1 := []byte("my first message")
	d2 := []byte("some other important info!")

	// try signing both data with both keys...
	s11, pub1, err := cstore.Sign(n1, p1, d1)
	require.Nil(t, err)
	require.Equal(t, i1.PubKey, pub1)

	s12, pub1, err := cstore.Sign(n1, p1, d2)
	require.Nil(t, err)
	require.Equal(t, i1.PubKey, pub1)

	s21, pub2, err := cstore.Sign(n2, p2, d1)
	require.Nil(t, err)
	require.Equal(t, i2.PubKey, pub2)

	s22, pub2, err := cstore.Sign(n2, p2, d2)
	require.Nil(t, err)
	require.Equal(t, i2.PubKey, pub2)

	// let's try to validate and make sure it only works when everything is proper
	cases := []struct {
		key   crypto.PubKey
		data  []byte
		sig   crypto.Signature
		valid bool
	}{
		// proper matches
		{i1.PubKey, d1, s11, true},
		// change data, pubkey, or signature leads to fail
		{i1.PubKey, d2, s11, false},
		{i2.PubKey, d1, s11, false},
		{i1.PubKey, d1, s21, false},
		// make sure other successes
		{i1.PubKey, d2, s12, true},
		{i2.PubKey, d1, s21, true},
		{i2.PubKey, d2, s22, true},
	}

	for i, tc := range cases {
		valid := tc.key.VerifyBytes(tc.data, tc.sig)
		assert.Equal(t, tc.valid, valid, "%d", i)
	}
}

/*
// TestSignWithLedger makes sure we have ledger compatibility with
// the crypto store.
//
// This test will only succeed with a ledger attached to the computer
// and the cosmos app open
func TestSignWithLedger(t *testing.T) {
	if os.Getenv("WITH_LEDGER") == "" {
		t.Skip("Set WITH_LEDGER to run code on real ledger")
	}

	// make the storage with reasonable defaults
	cstore := keys.New(
		dbm.NewMemDB(),
		words.MustLoadCodec("english"),
	)
	n := "nano-s"
	p := "hard2hack"

	// create a nano user
	c, _, err := cstore.Create(n, p, nano.KeyLedgerEd25519)
	require.Nil(t, err, "%+v", err)
	assert.Equal(t, c.Key, n)
	_, ok := c.PubKey.Unwrap().(nano.PubKeyLedgerEd25519)
	require.True(t, ok)

	// make sure we can get it back
	info, err := cstore.Get(n)
	require.Nil(t, err, "%+v", err)
	assert.Equal(t, info.Key, n)
	key := info.PubKey
	require.False(t ,key.Empty())
	require.True(t, key.Equals(c.PubKey))

	// let's try to sign some messages
	d1 := []byte("welcome to cosmos")
	d2 := []byte("please turn on the app")

	// try signing both data with the ledger...
	s1, pub, err := cstore.Sign(n, p, d1)
	require.Nil(t, err)
	require.Equal(t, info.PubKey, pub)

	s2, pub, err := cstore.Sign(n, p, d2)
	require.Nil(t, err)
	require.Equal(t, info.PubKey, pub)

	// now, let's check those signatures work
	assert.True(t, key.VerifyBytes(d1, s1))
	assert.True(t, key.VerifyBytes(d2, s2))
	// and mismatched signatures don't
	assert.False(t, key.VerifyBytes(d1, s2))
}
*/

func assertPassword(t *testing.T, cstore keys.Keybase, name, pass, badpass string) {
	err := cstore.Update(name, badpass, pass)
	assert.NotNil(t, err)
	err = cstore.Update(name, pass, pass)
	assert.Nil(t, err, "%+v", err)
}

// TestExportImport tests exporting and importing keys.
func TestExportImport(t *testing.T) {

	// make the storage with reasonable defaults
	db := dbm.NewMemDB()
	cstore := keys.New(
		db,
		words.MustLoadCodec("english"),
	)

	info, _, err := cstore.Create("john", "passphrase", keys.AlgoEd25519)
	assert.Nil(t, err)
	assert.Equal(t, info.Name, "john")
	addr := info.PubKey.Address()

	john, err := cstore.Get("john")
	assert.Nil(t, err)
	assert.Equal(t, john.Name, "john")
	assert.Equal(t, john.PubKey.Address(), addr)

	armor, err := cstore.Export("john")
	assert.Nil(t, err)

	err = cstore.Import("john2", armor)
	assert.Nil(t, err)

	john2, err := cstore.Get("john2")
	assert.Nil(t, err)

	assert.Equal(t, john.PubKey.Address(), addr)
	assert.Equal(t, john.Name, "john")
	assert.Equal(t, john, john2)
}

// TestAdvancedKeyManagement verifies update, import, export functionality
func TestAdvancedKeyManagement(t *testing.T) {

	// make the storage with reasonable defaults
	cstore := keys.New(
		dbm.NewMemDB(),
		words.MustLoadCodec("english"),
	)

	algo := keys.AlgoSecp256k1
	n1, n2 := "old-name", "new name"
	p1, p2 := "1234", "foobar"

	// make sure key works with initial password
	_, _, err := cstore.Create(n1, p1, algo)
	require.Nil(t, err, "%+v", err)
	assertPassword(t, cstore, n1, p1, p2)

	// update password requires the existing password
	err = cstore.Update(n1, "jkkgkg", p2)
	assert.NotNil(t, err)
	assertPassword(t, cstore, n1, p1, p2)

	// then it changes the password when correct
	err = cstore.Update(n1, p1, p2)
	assert.Nil(t, err)
	// p2 is now the proper one!
	assertPassword(t, cstore, n1, p2, p1)

	// exporting requires the proper name and passphrase
	_, err = cstore.Export(n1 + ".notreal")
	assert.NotNil(t, err)
	_, err = cstore.Export(" " + n1)
	assert.NotNil(t, err)
	_, err = cstore.Export(n1 + " ")
	assert.NotNil(t, err)
	_, err = cstore.Export("")
	assert.NotNil(t, err)
	exported, err := cstore.Export(n1)
	require.Nil(t, err, "%+v", err)

	// import succeeds
	err = cstore.Import(n2, exported)
	assert.Nil(t, err)

	// second import fails
	err = cstore.Import(n2, exported)
	assert.NotNil(t, err)
}

// TestSeedPhrase verifies restoring from a seed phrase
func TestSeedPhrase(t *testing.T) {

	// make the storage with reasonable defaults
	cstore := keys.New(
		dbm.NewMemDB(),
		words.MustLoadCodec("english"),
	)

	algo := keys.AlgoEd25519
	n1, n2 := "lost-key", "found-again"
	p1, p2 := "1234", "foobar"

	// make sure key works with initial password
	info, seed, err := cstore.Create(n1, p1, algo)
	require.Nil(t, err, "%+v", err)
	assert.Equal(t, n1, info.Name)
	assert.NotEmpty(t, seed)

	// now, let us delete this key
	err = cstore.Delete(n1, p1)
	require.Nil(t, err, "%+v", err)
	_, err = cstore.Get(n1)
	require.NotNil(t, err)

	// let us re-create it from the seed-phrase
	newInfo, err := cstore.Recover(n2, p2, seed)
	require.Nil(t, err, "%+v", err)
	assert.Equal(t, n2, newInfo.Name)
	assert.Equal(t, info.Address(), newInfo.Address())
	assert.Equal(t, info.PubKey, newInfo.PubKey)
}

func ExampleNew() {
	// Select the encryption and storage for your cryptostore
	cstore := keys.New(
		dbm.NewMemDB(),
		words.MustLoadCodec("english"),
	)
	ed := keys.AlgoEd25519
	sec := keys.AlgoSecp256k1

	// Add keys and see they return in alphabetical order
	bob, _, err := cstore.Create("Bob", "friend", ed)
	if err != nil {
		// this should never happen
		fmt.Println(err)
	} else {
		// return info here just like in List
		fmt.Println(bob.Name)
	}
	cstore.Create("Alice", "secret", sec)
	cstore.Create("Carl", "mitm", ed)
	info, _ := cstore.List()
	for _, i := range info {
		fmt.Println(i.Name)
	}

	// We need to use passphrase to generate a signature
	tx := []byte("deadbeef")
	sig, pub, err := cstore.Sign("Bob", "friend", tx)
	if err != nil {
		fmt.Println("don't accept real passphrase")
	}

	// and we can validate the signature with publically available info
	binfo, _ := cstore.Get("Bob")
	if !binfo.PubKey.Equals(bob.PubKey) {
		fmt.Println("Get and Create return different keys")
	}

	if pub.Equals(binfo.PubKey) {
		fmt.Println("signed by Bob")
	}
	if !pub.VerifyBytes(tx, sig) {
		fmt.Println("invalid signature")
	}

	// Output:
	// Bob
	// Alice
	// Bob
	// Carl
	// signed by Bob
}
