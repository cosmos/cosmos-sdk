package ecdsa

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/suite"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type EcdsaSuite struct {
	suite.Suite
	pk cryptotypes.PubKey
	sk ecdsaSK
}

func TestEcdsaSuite(t *testing.T) {
	suite.Run(t, new(EcdsaSuite))
}

func (suite *EcdsaSuite) SetupSuite() {
	sk, err := GenSecp256r1()
	suite.Assert().NoError(err)
	suite.pk = sk.PubKey()
}

func (suite *EcdsaSuite) TestString() {
	assert := suite.Assert()
	require := suite.Require()

	pkStr := suite.pk.String()
	prefix := "secp256r1{"
	require.Len(pkStr, len(prefix)+PubKeySize*2+1) // prefix + hex_len + "}"
	assert.Equal(prefix, pkStr[:len(prefix)])
	assert.EqualValues('}', pkStr[len(pkStr)-1])

	bz, err := hex.DecodeString(pkStr[len(prefix) : len(pkStr)-1])
	require.NoError(err)
	assert.EqualValues(suite.pk.Bytes(), bz)
}

func (suite *EcdsaSuite) TestEqual() {
	assert := suite.Assert()

	skOther, err := GenSecp256r1()
	assert.NoError(err)
	pkOther := skOther.PubKey()
	pkOther2 := ecdsaPK{&skOther.PublicKey, nil}

	assert.False(suite.pk.Equals(pkOther))

	assert.True(pkOther.Equals(pkOther2))
	assert.True(pkOther2.Equals(pkOther), "Equals must be reflexive")
}

func (suite *EcdsaSuite) TestMarshalAmino() {
	require := suite.Require()
	type AminoPubKey interface {
		cryptotypes.PubKey
		MarshalAmino() ([]byte, error)
	}

	pk := suite.pk.(AminoPubKey)
	bz, err := pk.MarshalAmino()
	require.NoError(err)

	var pk2 = new(ecdsaPK)
	require.NoError(pk2.UnmarshalAmino(bz))
	require.True(pk2.Equals(suite.pk))
}
