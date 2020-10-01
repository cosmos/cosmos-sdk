package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"
)

type decimalInternalTestSuite struct {
	suite.Suite
}

func TestDecimalInternalTestSuite(t *testing.T) {
	suite.Run(t, new(decimalInternalTestSuite))
}

func (s *decimalInternalTestSuite) TestPrecisionMultiplier() {
	res := precisionMultiplier(5)
	exp := big.NewInt(10000000000000)
	s.Require().Equal(0, res.Cmp(exp), "equality was incorrect, res %v, exp %v", res, exp)
}

func (s *decimalInternalTestSuite) TestZeroDeserializationJSON() {
	d := Dec{new(big.Int)}
	err := cdc.UnmarshalJSON([]byte(`"0"`), &d)
	s.Require().Nil(err)
	err = cdc.UnmarshalJSON([]byte(`"{}"`), &d)
	s.Require().NotNil(err)
}

func (s *decimalInternalTestSuite) TestSerializationGocodecJSON() {
	d := MustNewDecFromStr("0.333")

	bz, err := cdc.MarshalJSON(d)
	s.Require().NoError(err)

	d2 := Dec{new(big.Int)}
	err = cdc.UnmarshalJSON(bz, &d2)
	s.Require().NoError(err)
	s.Require().True(d.Equal(d2), "original: %v, unmarshalled: %v", d, d2)
}

func (s *decimalInternalTestSuite) TestDecMarshalJSON() {
	decimal := func(i int64) Dec {
		d := NewDec(0)
		d.i = new(big.Int).SetInt64(i)
		return d
	}
	tests := []struct {
		name    string
		d       Dec
		want    string
		wantErr bool // if wantErr = false, will also attempt unmarshaling
	}{
		{"zero", decimal(0), "\"0.000000000000000000\"", false},
		{"one", decimal(1), "\"0.000000000000000001\"", false},
		{"ten", decimal(10), "\"0.000000000000000010\"", false},
		{"12340", decimal(12340), "\"0.000000000000012340\"", false},
		{"zeroInt", NewDec(0), "\"0.000000000000000000\"", false},
		{"oneInt", NewDec(1), "\"1.000000000000000000\"", false},
		{"tenInt", NewDec(10), "\"10.000000000000000000\"", false},
		{"12340Int", NewDec(12340), "\"12340.000000000000000000\"", false},
	}
	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			got, err := tt.d.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Dec.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				s.Require().Equal(tt.want, string(got), "incorrect marshalled value")
				unmarshalledDec := NewDec(0)
				err := unmarshalledDec.UnmarshalJSON(got)
				s.Require().NoError(err)
				s.Require().Equal(tt.d, unmarshalledDec, "incorrect unmarshalled value")
			}
		})
	}
}
