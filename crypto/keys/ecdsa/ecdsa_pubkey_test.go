package ecdsa

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type EcdsaSuite struct {
	suite.Suite
	pk cryptotypes.PubKey
	sk cryptotypes.PrivKey
}

func TestEcdsaSuite(t *testing.T) {
	suite.Run(t, new(EcdsaSuite))
}

func (suite *EcdsaSuite) SetupSuite() {
	sk, err := GenSecp256r1()
	suite.Require().NoError(err)
	suite.sk = sk
	suite.pk = sk.PubKey()
}

func (suite *EcdsaSuite) TestPKString() {
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

func (suite *EcdsaSuite) TestPKBytes() {
	require := suite.Require()
	var pk *ecdsaPK
	require.Nil(pk.Bytes())
	require.Len(suite.pk.Bytes(), PubKeySize)
}

func (suite *EcdsaSuite) TestPKEquals() {
	require := suite.Require()

	skOther, err := GenSecp256r1()
	require.NoError(err)
	pkOther := skOther.PubKey()
	pkOther2 := &ecdsaPK{skOther.(ecdsaSK).PublicKey, nil}

	require.False(suite.pk.Equals(pkOther))
	require.True(pkOther.Equals(pkOther2))
	require.True(pkOther2.Equals(pkOther))
	require.True(pkOther.Equals(pkOther), "Equals must be reflexive")
}

func (suite *EcdsaSuite) TestPKMarshalAmino() {
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

func (suite *EcdsaSuite) TestPKSize() {
	require := suite.Require()
	bv := gogotypes.BytesValue{Value: suite.pk.Bytes()}
	require.Equal(bv.Size(), suite.pk.(*ecdsaPK).Size())

	var nilPk *ecdsaPK
	require.Equal(0, nilPk.Size(), "nil value must have zero size")
}

func (suite *EcdsaSuite) TestPKMarshalProto() {
	require := suite.Require()

	/**** test structure marshalling ****/

	var pk ecdsaPK
	bz, err := proto.Marshal(suite.pk)
	require.NoError(err)
	require.NoError(proto.Unmarshal(bz, &pk))
	require.True(pk.Equals(suite.pk))

	/**** test structure marshalling with codec ****/

	pk = ecdsaPK{}
	registry := types.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	bz, err = cdc.MarshalBinaryBare(suite.pk.(*ecdsaPK))
	require.NoError(err)
	require.NoError(cdc.UnmarshalBinaryBare(bz, &pk))
	require.True(pk.Equals(suite.pk))

	const bufSize = 100
	bz2 := make([]byte, bufSize)
	pkCpy := suite.pk.(*ecdsaPK)
	_, err = pkCpy.MarshalTo(bz2)
	require.NoError(err)
	require.Len(bz2, bufSize)
	require.Equal(bz, bz2[:sovPubKeySize])

	bz2 = make([]byte, bufSize)
	_, err = pkCpy.MarshalToSizedBuffer(bz2)
	require.NoError(err)
	require.Len(bz2, bufSize)
	require.Equal(bz, bz2[(bufSize-sovPubKeySize):])

	/**** test interface marshalling ****/

	bz, err = cdc.MarshalInterface(suite.pk)
	require.NoError(err)
	var pkI cryptotypes.PubKey
	err = cdc.UnmarshalInterface(bz, &pkI)
	require.EqualError(err, "no registered implementations of type types.PubKey")

	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), new(ecdsaPK))
	require.NoError(cdc.UnmarshalInterface(bz, &pkI))
	require.True(pkI.Equals(suite.pk))

	cdc.UnmarshalInterface(bz, nil)
	require.Error(err, "nil should fail")
}

func (suite *EcdsaSuite) TestPKReset() {
	pk := &ecdsaPK{ecdsa.PublicKey{X: big.NewInt(1)}, []byte{1}}
	pk.Reset()
	suite.Nil(pk.address)
	suite.Equal(pk.PublicKey, ecdsa.PublicKey{})
}
