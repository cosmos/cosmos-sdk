package ecdsa

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	proto "github.com/gogo/protobuf/proto"
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

func (suite *EcdsaSuite) TestSkMarshalProto() {
	require := suite.Require()

	/**** test structure marshalling ****/

	var sk ecdsaSK
	bz, err := proto.Marshal(suite.sk)
	require.NoError(err)
	require.NoError(proto.Unmarshal(bz, &sk))
	require.True(sk.Equals(suite.sk))

	/**** test structure marshalling with codec ****/

	sk = ecdsaSK{}
	registry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	bz, err = cdc.MarshalBinaryBare(suite.sk.(*ecdsaSK))
	require.NoError(err)
	require.NoError(cdc.UnmarshalBinaryBare(bz, &sk))
	require.True(sk.Equals(suite.sk))

	const bufSize = 100
	bz2 := make([]byte, bufSize)
	skCpy := suite.sk.(*ecdsaSK)
	_, err = skCpy.MarshalTo(bz2)
	require.NoError(err)
	require.Len(bz2, bufSize)
	require.Equal(bz, bz2[:sovPrivKeySize])

	bz2 = make([]byte, bufSize)
	_, err = skCpy.MarshalToSizedBuffer(bz2)
	require.NoError(err)
	require.Len(bz2, bufSize)
	require.Equal(bz, bz2[(bufSize-sovPrivKeySize):])
}

func (suite *EcdsaSuite) TestSign() {
}
