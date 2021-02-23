package ecdsa

import (
	"crypto/ecdsa"
	"math/big"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

var _ cryptotypes.PrivKey = &ecdsaSK{}

func (suite *EcdsaSuite) TestSkString() {
	suite.Require().Equal("secp256r1{-}", suite.sk.String())
}

func (suite *EcdsaSuite) TestSkEquals() {
	require := suite.Require()

	skOther, err := GenSecp256r1()
	require.NoError(err)
	// require.False(suite.sk.Equals(skOther))

	skOther2 := &ecdsaSK{skOther.(*ecdsaSK).PrivateKey}
	require.True(skOther.Equals(skOther2))
	// require.True(skOther2.Equals(skOther), "Equals must be reflexive")
}

func (suite *EcdsaSuite) TestSkPubKey() {
	pk := suite.sk.PubKey()
	suite.True(suite.sk.(*ecdsaSK).PublicKey.Equal(&pk.(*ecdsaPK).PublicKey))
}

func (suite *EcdsaSuite) Bytes() {
	bz := suite.sk.Bytes()
	suite.Len(bz, PrivKeySize)
	var sk *ecdsaSK
	suite.Nil(sk.Bytes())
}

func (suite *EcdsaSuite) TestSkReset() {
	var sk = &ecdsaSK{PrivateKey: ecdsa.PrivateKey{D: big.NewInt(1)}}
	sk.Reset()
	suite.Equal(0, sk.D.Cmp(big.NewInt(0)))
	suite.Equal(ecdsa.PublicKey{}, sk.PublicKey)
}

func (suite *EcdsaSuite) TestSkProtoMarshal() {
}

func (suite *EcdsaSuite) TestSkEquals2() {
}
