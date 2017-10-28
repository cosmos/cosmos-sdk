package cryptostore_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/tmlibs/common"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/cryptostore"
	"github.com/tendermint/go-crypto/keys/storage/memstorage"
	"github.com/tendermint/go-crypto/nano"
)

// TestKeyManagement makes sure we can manipulate these keys well
func TestKeyManagement(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// make the storage with reasonable defaults
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		memstorage.New(),
		keys.MustLoadCodec("english"),
	)

	algo := crypto.NameEd25519
	n1, n2, n3 := "personal", "business", "other"
	p1, p2 := "1234", "really-secure!@#$"

	// Check empty state
	l, err := cstore.List()
	require.Nil(err)
	assert.Empty(l)

	// create some keys
	_, err = cstore.Get(n1)
	assert.NotNil(err)
	i, _, err := cstore.Create(n1, p1, algo)
	require.Equal(n1, i.Name)
	require.Nil(err)
	_, _, err = cstore.Create(n2, p2, algo)
	require.Nil(err)

	// we can get these keys
	i2, err := cstore.Get(n2)
	assert.Nil(err)
	_, err = cstore.Get(n3)
	assert.NotNil(err)

	// list shows them in order
	keyS, err := cstore.List()
	require.Nil(err)
	require.Equal(2, len(keyS))
	// note these are in alphabetical order
	assert.Equal(n2, keyS[0].Name)
	assert.Equal(n1, keyS[1].Name)
	assert.Equal(i2.PubKey, keyS[0].PubKey)

	// deleting a key removes it
	err = cstore.Delete("bad name", "foo")
	require.NotNil(err)
	err = cstore.Delete(n1, p1)
	require.Nil(err)
	keyS, err = cstore.List()
	require.Nil(err)
	assert.Equal(1, len(keyS))
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
func TestSignVerify(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// make the storage with reasonable defaults
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		memstorage.New(),
		keys.MustLoadCodec("english"),
	)
	algo := crypto.NameSecp256k1

	n1, n2 := "some dude", "a dudette"
	p1, p2 := "1234", "foobar"

	// create two users and get their info
	i1, _, err := cstore.Create(n1, p1, algo)
	require.Nil(err)

	i2, _, err := cstore.Create(n2, p2, algo)
	require.Nil(err)

	// let's try to sign some messages
	d1 := []byte("my first message")
	d2 := []byte("some other important info!")

	// try signing both data with both keys...
	s11 := keys.NewMockSignable(d1)
	err = cstore.Sign(n1, p1, s11)
	require.Nil(err)
	s12 := keys.NewMockSignable(d2)
	err = cstore.Sign(n1, p1, s12)
	require.Nil(err)
	s21 := keys.NewMockSignable(d1)
	err = cstore.Sign(n2, p2, s21)
	require.Nil(err)
	s22 := keys.NewMockSignable(d2)
	err = cstore.Sign(n2, p2, s22)
	require.Nil(err)

	// let's try to validate and make sure it only works when everything is proper
	cases := []struct {
		key   crypto.PubKey
		data  []byte
		sig   crypto.Signature
		valid bool
	}{
		// proper matches
		{i1.PubKey, d1, s11.Signature, true},
		// change data, pubkey, or signature leads to fail
		{i1.PubKey, d2, s11.Signature, false},
		{i2.PubKey, d1, s11.Signature, false},
		{i1.PubKey, d1, s21.Signature, false},
		// make sure other successes
		{i1.PubKey, d2, s12.Signature, true},
		{i2.PubKey, d1, s21.Signature, true},
		{i2.PubKey, d2, s22.Signature, true},
	}

	for i, tc := range cases {
		valid := tc.key.VerifyBytes(tc.data, tc.sig)
		assert.Equal(tc.valid, valid, "%d", i)
	}
}

// TestSignWithLedger makes sure we have ledger compatibility with
// the crypto store.
//
// This test will only succeed with a ledger attached to the computer
// and the cosmos app open
func TestSignWithLedger(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	if os.Getenv("WITH_LEDGER") == "" {
		t.Skip("Set WITH_LEDGER to run code on real ledger")
	}

	// make the storage with reasonable defaults
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		memstorage.New(),
		keys.MustLoadCodec("english"),
	)
	n := "nano-s"
	p := "hard2hack"

	// create a nano user
	c, _, err := cstore.Create(n, p, nano.NameLedgerEd25519)
	require.Nil(err, "%+v", err)
	assert.Equal(c.Name, n)
	_, ok := c.PubKey.Unwrap().(nano.PubKeyLedgerEd25519)
	require.True(ok)

	// make sure we can get it back
	info, err := cstore.Get(n)
	require.Nil(err, "%+v", err)
	assert.Equal(info.Name, n)
	key := info.PubKey
	require.False(key.Empty())
	require.True(key.Equals(c.PubKey))

	// let's try to sign some messages
	d1 := []byte("welcome to cosmos")
	d2 := []byte("please turn on the app")

	// try signing both data with the ledger...
	s1 := keys.NewMockSignable(d1)
	err = cstore.Sign(n, p, s1)
	require.Nil(err)
	s2 := keys.NewMockSignable(d2)
	err = cstore.Sign(n, p, s2)
	require.Nil(err)

	// now, let's check those signatures work
	assert.True(key.VerifyBytes(d1, s1.Signature))
	assert.True(key.VerifyBytes(d2, s2.Signature))
	// and mismatched signatures don't
	assert.False(key.VerifyBytes(d1, s2.Signature))
}

func assertPassword(assert *assert.Assertions, cstore cryptostore.Manager, name, pass, badpass string) {
	err := cstore.Update(name, badpass, pass)
	assert.NotNil(err)
	err = cstore.Update(name, pass, pass)
	assert.Nil(err, "%+v", err)
}

// TestImportUnencrypted tests accepting raw priv keys bytes as input
func TestImportUnencrypted(t *testing.T) {
	require := require.New(t)

	// make the storage with reasonable defaults
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		memstorage.New(),
		keys.MustLoadCodec("english"),
	)

	key, err := cryptostore.GenEd25519.Generate(cmn.RandBytes(16))
	require.NoError(err)

	addr := key.PubKey().Address()
	name := "john"
	pass := "top-secret"

	// import raw bytes
	err = cstore.Import(name, pass, "", key.Bytes())
	require.Nil(err, "%+v", err)

	// make sure the address matches
	info, err := cstore.Get(name)
	require.Nil(err, "%+v", err)
	require.EqualValues(addr, info.Address)
}

// TestAdvancedKeyManagement verifies update, import, export functionality
func TestAdvancedKeyManagement(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// make the storage with reasonable defaults
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		memstorage.New(),
		keys.MustLoadCodec("english"),
	)

	algo := crypto.NameSecp256k1
	n1, n2 := "old-name", "new name"
	p1, p2, p3, pt := "1234", "foobar", "ding booms!", "really-secure!@#$"

	// make sure key works with initial password
	_, _, err := cstore.Create(n1, p1, algo)
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
}

// TestSeedPhrase verifies restoring from a seed phrase
func TestSeedPhrase(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// make the storage with reasonable defaults
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		memstorage.New(),
		keys.MustLoadCodec("english"),
	)

	algo := crypto.NameEd25519
	n1, n2 := "lost-key", "found-again"
	p1, p2 := "1234", "foobar"

	// make sure key works with initial password
	info, seed, err := cstore.Create(n1, p1, algo)
	require.Nil(err, "%+v", err)
	assert.Equal(n1, info.Name)
	assert.NotEmpty(seed)

	// now, let us delete this key
	err = cstore.Delete(n1, p1)
	require.Nil(err, "%+v", err)
	_, err = cstore.Get(n1)
	require.NotNil(err)

	// let us re-create it from the seed-phrase
	newInfo, err := cstore.Recover(n2, p2, seed)
	require.Nil(err, "%+v", err)
	assert.Equal(n2, newInfo.Name)
	assert.Equal(info.Address, newInfo.Address)
	assert.Equal(info.PubKey, newInfo.PubKey)
}

func ExampleNew() {
	// Select the encryption and storage for your cryptostore
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		// Note: use filestorage.New(dir) for real data
		memstorage.New(),
		keys.MustLoadCodec("english"),
	)
	ed := crypto.NameEd25519
	sec := crypto.NameSecp256k1

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
	tx := keys.NewMockSignable([]byte("deadbeef"))
	err = cstore.Sign("Bob", "friend", tx)
	if err != nil {
		fmt.Println("don't accept real passphrase")
	}

	// and we can validate the signature with publically available info
	binfo, _ := cstore.Get("Bob")
	if !binfo.PubKey.Equals(bob.PubKey) {
		fmt.Println("Get and Create return different keys")
	}

	sigs, err := tx.Signers()
	if err != nil {
		fmt.Println("badly signed")
	} else if bytes.Equal(sigs[0].Bytes(), binfo.PubKey.Bytes()) {
		fmt.Println("signed by Bob")
	} else {
		fmt.Println("signed by someone else")
	}

	// Output:
	// Bob
	// Alice
	// Bob
	// Carl
	// signed by Bob
}
