package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/codec/testdata"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

type TypeWithInterface struct {
	Animal testdata.Animal `json:"animal"`
	X      int64           `json:"x,omitempty"`
}

type Suite struct {
	suite.Suite
	cdc  *amino.Codec
	a    TypeWithInterface
	b    testdata.HasAnimal
	spot *testdata.Dog
}

func (s *Suite) SetupTest() {
	s.cdc = amino.NewCodec()
	s.cdc.RegisterInterface((*testdata.Animal)(nil), nil)
	s.cdc.RegisterConcrete(&testdata.Dog{}, "testdata/Dob", nil)

	s.spot = &testdata.Dog{Size_: "small", Name: "Spot"}
	s.a = TypeWithInterface{Animal: s.spot}

	any, err := types.NewAnyWithValue(s.spot)
	s.Require().NoError(err)
	s.b = testdata.HasAnimal{Animal: any}
}

func (s *Suite) TestAminoBinary() {
	bz, err := s.cdc.MarshalBinaryBare(s.a)
	s.Require().NoError(err)

	// expect plain amino marshal to fail
	_, err = s.cdc.MarshalBinaryBare(s.b)
	s.Require().Error(err)

	// expect unpack interfaces before amino marshal to succeed
	err = types.UnpackInterfaces(s.b, types.AminoPacker{Cdc: s.cdc})
	s.Require().NoError(err)
	bz2, err := s.cdc.MarshalBinaryBare(s.b)
	s.Require().NoError(err)
	s.Require().Equal(bz, bz2)

	var c testdata.HasAnimal
	err = s.cdc.UnmarshalBinaryBare(bz, &c)
	s.Require().NoError(err)
	err = types.UnpackInterfaces(c, types.AminoUnpacker{Cdc: s.cdc})
	s.Require().NoError(err)
	s.Require().Equal(s.spot, c.Animal.GetCachedValue())
}

func (s *Suite) TestAminoJSON() {
	bz, err := s.cdc.MarshalJSON(s.a)
	s.Require().NoError(err)

	// expect plain amino marshal to fail
	_, err = s.cdc.MarshalJSON(s.b)
	s.Require().Error(err)

	// expect unpack interfaces before amino marshal to succeed
	err = types.UnpackInterfaces(s.b, types.AminoJSONPacker{Cdc: s.cdc})
	s.Require().NoError(err)
	bz2, err := s.cdc.MarshalJSON(s.b)
	s.Require().NoError(err)
	s.Require().Equal(string(bz), string(bz2))

	var c testdata.HasAnimal
	err = s.cdc.UnmarshalJSON(bz, &c)
	s.Require().NoError(err)
	err = types.UnpackInterfaces(c, types.AminoJSONUnpacker{Cdc: s.cdc})
	s.Require().NoError(err)
	s.Require().Equal(s.spot, c.Animal.GetCachedValue())
}

func TestSuite(t *testing.T) {
	suite.Run(t, &Suite{})
}
