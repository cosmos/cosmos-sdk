package crypto

import (
	"bytes"
	"testing"

	"github.com/tendermint/ed25519"
	"github.com/tendermint/go-wire"
)

func TestSignAndValidateEd25519(t *testing.T) {

	privKey := GenPrivKeyEd25519()
	pubKey := privKey.PubKey()

	msg := CRandBytes(128)
	sig := privKey.Sign(msg)
	t.Logf("msg: %X, sig: %X", msg, sig)

	// Test the signature
	if !pubKey.VerifyBytes(msg, sig) {
		t.Errorf("Account message signature verification failed")
	}

	// Mutate the signature, just one bit.
	sigEd := sig.(SignatureEd25519)
	sigEd[0] ^= byte(0x01)
	sig = Signature(sigEd)

	if pubKey.VerifyBytes(msg, sig) {
		t.Errorf("Account message signature verification should have failed but passed instead")
	}
}

func TestSignAndValidateSecp256k1(t *testing.T) {

	privKey := GenPrivKeySecp256k1()
	pubKey := privKey.PubKey()

	msg := CRandBytes(128)
	sig := privKey.Sign(msg)
	t.Logf("msg: %X, sig: %X", msg, sig)

	// Test the signature
	if !pubKey.VerifyBytes(msg, sig) {
		t.Errorf("Account message signature verification failed")
	}

	// Mutate the signature, just one bit.
	sigEd := sig.(SignatureSecp256k1)
	sigEd[0] ^= byte(0x01)
	sig = Signature(sigEd)

	if pubKey.VerifyBytes(msg, sig) {
		t.Errorf("Account message signature verification should have failed but passed instead")
	}
}

func TestBinaryDecodeEd25519(t *testing.T) {

	privKey := GenPrivKeyEd25519()
	pubKey := privKey.PubKey()

	msg := CRandBytes(128)
	sig := privKey.Sign(msg)
	t.Logf("msg: %X, sig: %X", msg, sig)

	buf, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteBinary(struct{ Signature }{sig}, buf, n, err)
	if *err != nil {
		t.Fatalf("Failed to write Signature: %v", err)
	}

	if len(buf.Bytes()) != ed25519.SignatureSize+1 {
		// 1 byte TypeByte, 64 bytes signature bytes
		t.Fatalf("Unexpected signature write size: %v", len(buf.Bytes()))
	}
	if buf.Bytes()[0] != SignatureTypeEd25519 {
		t.Fatalf("Unexpected signature type byte")
	}

	sigStruct := struct{ Signature }{}
	sig2 := wire.ReadBinary(sigStruct, buf, 0, n, err)
	if *err != nil {
		t.Fatalf("Failed to read Signature: %v", err)
	}

	// Test the signature
	if !pubKey.VerifyBytes(msg, sig2.(struct{ Signature }).Signature.(SignatureEd25519)) {
		t.Errorf("Account message signature verification failed")
	}
}

func TestBinaryDecodeSecp256k1(t *testing.T) {

	privKey := GenPrivKeySecp256k1()
	pubKey := privKey.PubKey()

	msg := CRandBytes(128)
	sig := privKey.Sign(msg)
	t.Logf("msg: %X, sig: %X", msg, sig)

	buf, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteBinary(struct{ Signature }{sig}, buf, n, err)
	if *err != nil {
		t.Fatalf("Failed to write Signature: %v", err)
	}

	if buf.Bytes()[0] != SignatureTypeSecp256k1 {
		t.Fatalf("Unexpected signature type byte")
	}

	sigStruct := struct{ Signature }{}
	sig2 := wire.ReadBinary(sigStruct, buf, 0, n, err)
	if *err != nil {
		t.Fatalf("Failed to read Signature: %v", err)
	}

	// Test the signature
	if !pubKey.VerifyBytes(msg, sig2.(struct{ Signature }).Signature.(SignatureSecp256k1)) {
		t.Errorf("Account message signature verification failed")
	}
}
