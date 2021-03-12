package ecdsa

import (
	"crypto/rand"
	"testing"

	"github.com/tendermint/tendermint/crypto"

	"github.com/stretchr/testify/suite"
)

func TestSKSuite(t *testing.T) {
	suite.Run(t, new(SKSuite))
}

type SKSuite struct{ CommonSuite }

func (suite *SKSuite) TestNewPrivKey() {
	require := suite.Require()
	cparams := secp256r1.Params()

	secret := make([]byte, cparams.BitSize)
	n, err := rand.Reader.Read(secret)
	require.NoError(err)
	require.Equal(cparams.BitSize, n)

	sk, err := NewPrivKey(secp256r1, secret)
	require.NoError(err)
	for i := 0; i < 10; i++ {
		sk2, err := NewPrivKey(secp256r1, secret)
		require.NoError(err)
		require.True(sk.Equal(&sk2.PrivateKey), "NewPrivKey must be deterministic")
	}

	// Test if NewPrivKeyFromSecret correctly creates a private key

	secret = suite.sk.D.Bytes()
	sk, err = NewPrivKeyFromSecret(secp256r1, secret)
	require.NoError(err)
	require.True(sk.Equal(&suite.sk.PrivateKey), "NewPrivKeyFromSecret must correctly recreate a private key")
}

func (suite *SKSuite) TestString() {
	const prefix = "abc"
	suite.Require().Equal(prefix+"{-}", suite.sk.String(prefix))
}

func (suite *SKSuite) TestPubKey() {
	pk := suite.sk.PubKey()
	suite.True(suite.sk.PublicKey.Equal(&pk.PublicKey))
}

func (suite *SKSuite) Bytes() {
	bz := suite.sk.Bytes()
	suite.Len(bz, 32)
	var sk *PrivKey
	suite.Nil(sk.Bytes())
}

func (suite *SKSuite) TestMarshal() {
	require := suite.Require()
	const size = 32

	var buffer = make([]byte, size)
	suite.sk.MarshalTo(buffer)

	var sk = new(PrivKey)
	err := sk.Unmarshal(buffer, secp256r1, size)
	require.NoError(err)
	require.True(sk.Equal(&suite.sk.PrivateKey))
}

func (suite *SKSuite) TestSign() {
	require := suite.Require()

	msg := crypto.CRandBytes(1000)
	sig, err := suite.sk.Sign(msg)
	require.NoError(err)
	sigCpy := make([]byte, len(sig))
	copy(sigCpy, sig)
	require.True(suite.pk.VerifySignature(msg, sigCpy))

	// Mutate the signature
	for i := range sig {
		sigCpy[i] ^= byte(i + 1)
		require.False(suite.pk.VerifySignature(msg, sigCpy))
	}

	// Mutate the message
	msg[1] ^= byte(2)
	require.False(suite.pk.VerifySignature(msg, sig))
}
