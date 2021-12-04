package types_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
)

type addressTestSuite struct {
	suite.Suite
}

func TestAddressTestSuite(t *testing.T) {
	suite.Run(t, new(addressTestSuite))
}

func (s *addressTestSuite) SetupSuite() {
	s.T().Parallel()
}

var invalidStrs = []string{
	"hello, world!",
	"0xAA",
	"AAA",
	types.Bech32PrefixAccAddr + "AB0C",
	types.Bech32PrefixAccPub + "1234",
	types.Bech32PrefixValAddr + "5678",
	types.Bech32PrefixValPub + "BBAB",
	types.Bech32PrefixConsAddr + "FF04",
	types.Bech32PrefixConsPub + "6789",
}

func (s *addressTestSuite) testMarshal(original interface{}, res interface{}, marshal func() ([]byte, error), unmarshal func([]byte) error) {
	bz, err := marshal()
	s.Require().Nil(err)
	s.Require().Nil(unmarshal(bz))
	s.Require().Equal(original, res)
}

func (s *addressTestSuite) TestEmptyAddresses() {
	s.T().Parallel()
	s.Require().Equal((types.AccAddress{}).String(), "")
	s.Require().Equal((types.ValAddress{}).String(), "")
	s.Require().Equal((types.ConsAddress{}).String(), "")

	accAddr, err := types.AccAddressFromBech32("")
	s.Require().True(accAddr.Empty())
	s.Require().Error(err)

	valAddr, err := types.ValAddressFromBech32("")
	s.Require().True(valAddr.Empty())
	s.Require().Error(err)

	consAddr, err := types.ConsAddressFromBech32("")
	s.Require().True(consAddr.Empty())
	s.Require().Error(err)
}

func (s *addressTestSuite) TestRandBech32PubkeyConsistency() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	for i := 0; i < 1000; i++ {
		rand.Read(pub.Key)

		mustBech32AccPub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeAccPub, pub)
		bech32AccPub, err := types.Bech32ifyPubKey(types.Bech32PubKeyTypeAccPub, pub)
		s.Require().Nil(err)
		s.Require().Equal(bech32AccPub, mustBech32AccPub)

		mustBech32ValPub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeValPub, pub)
		bech32ValPub, err := types.Bech32ifyPubKey(types.Bech32PubKeyTypeValPub, pub)
		s.Require().Nil(err)
		s.Require().Equal(bech32ValPub, mustBech32ValPub)

		mustBech32ConsPub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeConsPub, pub)
		bech32ConsPub, err := types.Bech32ifyPubKey(types.Bech32PubKeyTypeConsPub, pub)
		s.Require().Nil(err)
		s.Require().Equal(bech32ConsPub, mustBech32ConsPub)

		mustAccPub := types.MustGetPubKeyFromBech32(types.Bech32PubKeyTypeAccPub, bech32AccPub)
		accPub, err := types.GetPubKeyFromBech32(types.Bech32PubKeyTypeAccPub, bech32AccPub)
		s.Require().Nil(err)
		s.Require().Equal(accPub, mustAccPub)

		mustValPub := types.MustGetPubKeyFromBech32(types.Bech32PubKeyTypeValPub, bech32ValPub)
		valPub, err := types.GetPubKeyFromBech32(types.Bech32PubKeyTypeValPub, bech32ValPub)
		s.Require().Nil(err)
		s.Require().Equal(valPub, mustValPub)

		mustConsPub := types.MustGetPubKeyFromBech32(types.Bech32PubKeyTypeConsPub, bech32ConsPub)
		consPub, err := types.GetPubKeyFromBech32(types.Bech32PubKeyTypeConsPub, bech32ConsPub)
		s.Require().Nil(err)
		s.Require().Equal(consPub, mustConsPub)

		s.Require().Equal(valPub, accPub)
		s.Require().Equal(valPub, consPub)
	}
}

func (s *addressTestSuite) TestYAMLMarshalers() {
	addr := secp256k1.GenPrivKey().PubKey().Address()

	acc := types.AccAddress(addr)
	val := types.ValAddress(addr)
	cons := types.ConsAddress(addr)

	got, _ := yaml.Marshal(&acc)
	s.Require().Equal(acc.String()+"\n", string(got))

	got, _ = yaml.Marshal(&val)
	s.Require().Equal(val.String()+"\n", string(got))

	got, _ = yaml.Marshal(&cons)
	s.Require().Equal(cons.String()+"\n", string(got))
}

func (s *addressTestSuite) TestRandBech32AccAddrConsistency() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	for i := 0; i < 1000; i++ {
		rand.Read(pub.Key)

		acc := types.AccAddress(pub.Address())
		res := types.AccAddress{}

		s.testMarshal(&acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		s.testMarshal(&acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := types.AccAddressFromBech32(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)

		str = hex.EncodeToString(acc)
		res, err = types.AccAddressFromHex(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)
	}

	for _, str := range invalidStrs {
		_, err := types.AccAddressFromHex(str)
		s.Require().NotNil(err)

		_, err = types.AccAddressFromBech32(str)
		s.Require().NotNil(err)

		err = (*types.AccAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		s.Require().NotNil(err)
	}

	_, err := types.AccAddressFromHex("")
	s.Require().Equal("decoding Bech32 address failed: must provide an address", err.Error())
}

func (s *addressTestSuite) TestValAddr() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	for i := 0; i < 20; i++ {
		rand.Read(pub.Key)

		acc := types.ValAddress(pub.Address())
		res := types.ValAddress{}

		s.testMarshal(&acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		s.testMarshal(&acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := types.ValAddressFromBech32(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)

		str = hex.EncodeToString(acc)
		res, err = types.ValAddressFromHex(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)

	}

	for _, str := range invalidStrs {
		_, err := types.ValAddressFromHex(str)
		s.Require().NotNil(err)

		_, err = types.ValAddressFromBech32(str)
		s.Require().NotNil(err)

		err = (*types.ValAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		s.Require().NotNil(err)
	}

	// test empty string
	_, err := types.ValAddressFromHex("")
	s.Require().Equal("decoding Bech32 address failed: must provide an address", err.Error())
}

func (s *addressTestSuite) TestConsAddress() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	for i := 0; i < 20; i++ {
		rand.Read(pub.Key[:])

		acc := types.ConsAddress(pub.Address())
		res := types.ConsAddress{}

		s.testMarshal(&acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		s.testMarshal(&acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := types.ConsAddressFromBech32(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)

		str = hex.EncodeToString(acc)
		res, err = types.ConsAddressFromHex(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)
	}

	for _, str := range invalidStrs {
		_, err := types.ConsAddressFromHex(str)
		s.Require().NotNil(err)

		_, err = types.ConsAddressFromBech32(str)
		s.Require().NotNil(err)

		err = (*types.ConsAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		s.Require().NotNil(err)
	}

	// test empty string
	_, err := types.ConsAddressFromHex("")
	s.Require().Equal("decoding Bech32 address failed: must provide an address", err.Error())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (s *addressTestSuite) TestConfiguredPrefix() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}
	for length := 1; length < 10; length++ {
		for times := 1; times < 20; times++ {
			rand.Read(pub.Key[:])
			// Test if randomly generated prefix of a given length works
			prefix := RandString(length)

			// Assuming that GetConfig is not sealed.
			config := types.GetConfig()
			config.SetBech32PrefixForAccount(
				prefix+types.PrefixAccount,
				prefix+types.PrefixPublic)

			acc := types.AccAddress(pub.Address())
			s.Require().True(strings.HasPrefix(
				acc.String(),
				prefix+types.PrefixAccount), acc.String())

			bech32Pub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeAccPub, pub)
			s.Require().True(strings.HasPrefix(
				bech32Pub,
				prefix+types.PrefixPublic))

			config.SetBech32PrefixForValidator(
				prefix+types.PrefixValidator+types.PrefixAddress,
				prefix+types.PrefixValidator+types.PrefixPublic)

			val := types.ValAddress(pub.Address())
			s.Require().True(strings.HasPrefix(
				val.String(),
				prefix+types.PrefixValidator+types.PrefixAddress))

			bech32ValPub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeValPub, pub)
			s.Require().True(strings.HasPrefix(
				bech32ValPub,
				prefix+types.PrefixValidator+types.PrefixPublic))

			config.SetBech32PrefixForConsensusNode(
				prefix+types.PrefixConsensus+types.PrefixAddress,
				prefix+types.PrefixConsensus+types.PrefixPublic)

			cons := types.ConsAddress(pub.Address())
			s.Require().True(strings.HasPrefix(
				cons.String(),
				prefix+types.PrefixConsensus+types.PrefixAddress))

			bech32ConsPub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeConsPub, pub)
			s.Require().True(strings.HasPrefix(
				bech32ConsPub,
				prefix+types.PrefixConsensus+types.PrefixPublic))
		}
	}
}

func (s *addressTestSuite) TestAddressInterface() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}
	rand.Read(pub.Key)

	addrs := []types.Address{
		types.ConsAddress(pub.Address()),
		types.ValAddress(pub.Address()),
		types.AccAddress(pub.Address()),
	}

	for _, addr := range addrs {
		switch addr := addr.(type) {
		case types.AccAddress:
			_, err := types.AccAddressFromBech32(addr.String())
			s.Require().Nil(err)
		case types.ValAddress:
			_, err := types.ValAddressFromBech32(addr.String())
			s.Require().Nil(err)
		case types.ConsAddress:
			_, err := types.ConsAddressFromBech32(addr.String())
			s.Require().Nil(err)
		default:
			s.T().Fail()
		}
	}

}

func (s *addressTestSuite) TestVerifyAddressFormat() {
	addr0 := make([]byte, 0)
	addr5 := make([]byte, 5)
	addr20 := make([]byte, 20)
	addr32 := make([]byte, 32)

	err := types.VerifyAddressFormat(addr0)
	s.Require().EqualError(err, "incorrect address length 0")
	err = types.VerifyAddressFormat(addr5)
	s.Require().EqualError(err, "incorrect address length 5")
	err = types.VerifyAddressFormat(addr20)
	s.Require().Nil(err)
	err = types.VerifyAddressFormat(addr32)
	s.Require().EqualError(err, "incorrect address length 32")
}

func (s *addressTestSuite) TestCustomAddressVerifier() {
	// Create a 10 byte address
	addr := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	accBech := types.AccAddress(addr).String()
	valBech := types.ValAddress(addr).String()
	consBech := types.ConsAddress(addr).String()
	// Verifiy that the default logic rejects this 10 byte address
	err := types.VerifyAddressFormat(addr)
	s.Require().NotNil(err)
	_, err = types.AccAddressFromBech32(accBech)
	s.Require().NotNil(err)
	_, err = types.ValAddressFromBech32(valBech)
	s.Require().NotNil(err)
	_, err = types.ConsAddressFromBech32(consBech)
	s.Require().NotNil(err)

	// Set a custom address verifier that accepts 10 or 20 byte addresses
	types.GetConfig().SetAddressVerifier(func(bz []byte) error {
		n := len(bz)
		if n == 10 || n == types.AddrLen {
			return nil
		}
		return fmt.Errorf("incorrect address length %d", n)
	})

	// Verifiy that the custom logic accepts this 10 byte address
	err = types.VerifyAddressFormat(addr)
	s.Require().Nil(err)
	_, err = types.AccAddressFromBech32(accBech)
	s.Require().Nil(err)
	_, err = types.ValAddressFromBech32(valBech)
	s.Require().Nil(err)
	_, err = types.ConsAddressFromBech32(consBech)
	s.Require().Nil(err)
}

func (s *addressTestSuite) TestBech32ifyAddressBytes() {
	addr10byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	type args struct {
		prefix string
		bs     []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty address", args{"prefixa", []byte{}}, "", false},
		{"empty prefix", args{"", addr20byte}, "", true},
		{"10-byte address", args{"prefixa", addr10byte}, "prefixa1qqqsyqcyq5rqwzqf3953cc", false},
		{"10-byte address", args{"prefixb", addr10byte}, "prefixb1qqqsyqcyq5rqwzqf20xxpc", false},
		{"20-byte address", args{"prefixa", addr20byte}, "prefixa1qqqsyqcyq5rqwzqfpg9scrgwpugpzysn7hzdtn", false},
		{"20-byte address", args{"prefixb", addr20byte}, "prefixb1qqqsyqcyq5rqwzqfpg9scrgwpugpzysnrujsuw", false},
	}
	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			got, err := types.Bech32ifyAddressBytes(tt.args.prefix, tt.args.bs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bech32ifyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func (s *addressTestSuite) TestMustBech32ifyAddressBytes() {
	addr10byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	type args struct {
		prefix string
		bs     []byte
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantPanic bool
	}{
		{"empty address", args{"prefixa", []byte{}}, "", false},
		{"empty prefix", args{"", addr20byte}, "", true},
		{"10-byte address", args{"prefixa", addr10byte}, "prefixa1qqqsyqcyq5rqwzqf3953cc", false},
		{"10-byte address", args{"prefixb", addr10byte}, "prefixb1qqqsyqcyq5rqwzqf20xxpc", false},
		{"20-byte address", args{"prefixa", addr20byte}, "prefixa1qqqsyqcyq5rqwzqfpg9scrgwpugpzysn7hzdtn", false},
		{"20-byte address", args{"prefixb", addr20byte}, "prefixb1qqqsyqcyq5rqwzqfpg9scrgwpugpzysnrujsuw", false},
	}
	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() { types.MustBech32ifyAddressBytes(tt.args.prefix, tt.args.bs) })
				return
			}
			require.Equal(t, tt.want, types.MustBech32ifyAddressBytes(tt.args.prefix, tt.args.bs))
		})
	}
}

func (s *addressTestSuite) TestAddressTypesEquals() {
	addr1 := secp256k1.GenPrivKey().PubKey().Address()
	accAddr1 := types.AccAddress(addr1)
	consAddr1 := types.ConsAddress(addr1)
	valAddr1 := types.ValAddress(addr1)

	addr2 := secp256k1.GenPrivKey().PubKey().Address()
	accAddr2 := types.AccAddress(addr2)
	consAddr2 := types.ConsAddress(addr2)
	valAddr2 := types.ValAddress(addr2)

	// equality
	s.Require().True(accAddr1.Equals(accAddr1))
	s.Require().True(consAddr1.Equals(consAddr1))
	s.Require().True(valAddr1.Equals(valAddr1))

	// emptiness
	s.Require().True(types.AccAddress{}.Equals(types.AccAddress{}))
	s.Require().True(types.AccAddress{}.Equals(types.AccAddress(nil)))
	s.Require().True(types.AccAddress(nil).Equals(types.AccAddress{}))
	s.Require().True(types.AccAddress(nil).Equals(types.AccAddress(nil)))

	s.Require().True(types.ConsAddress{}.Equals(types.ConsAddress{}))
	s.Require().True(types.ConsAddress{}.Equals(types.ConsAddress(nil)))
	s.Require().True(types.ConsAddress(nil).Equals(types.ConsAddress{}))
	s.Require().True(types.ConsAddress(nil).Equals(types.ConsAddress(nil)))

	s.Require().True(types.ValAddress{}.Equals(types.ValAddress{}))
	s.Require().True(types.ValAddress{}.Equals(types.ValAddress(nil)))
	s.Require().True(types.ValAddress(nil).Equals(types.ValAddress{}))
	s.Require().True(types.ValAddress(nil).Equals(types.ValAddress(nil)))

	s.Require().False(accAddr1.Equals(accAddr2))
	s.Require().Equal(accAddr1.Equals(accAddr2), accAddr2.Equals(accAddr1))
	s.Require().False(consAddr1.Equals(consAddr2))
	s.Require().Equal(consAddr1.Equals(consAddr2), consAddr2.Equals(consAddr1))
	s.Require().False(valAddr1.Equals(valAddr2))
	s.Require().Equal(valAddr1.Equals(valAddr2), valAddr2.Equals(valAddr1))
}

func (s *addressTestSuite) TestNilAddressTypesEmpty() {
	s.Require().True(types.AccAddress(nil).Empty())
	s.Require().True(types.ConsAddress(nil).Empty())
	s.Require().True(types.ValAddress(nil).Empty())
}

func (s *addressTestSuite) TestGetConsAddress() {
	pk := secp256k1.GenPrivKey().PubKey()
	s.Require().NotEqual(types.GetConsAddress(pk), pk.Address())
	s.Require().True(bytes.Equal(types.GetConsAddress(pk).Bytes(), pk.Address().Bytes()))
	s.Require().Panics(func() { types.GetConsAddress(cryptotypes.PubKey(nil)) })
}

func (s *addressTestSuite) TestGetFromBech32() {
	_, err := types.GetFromBech32("", "prefix")
	s.Require().Error(err)
	s.Require().Equal("decoding Bech32 address failed: must provide an address", err.Error())
	_, err = types.GetFromBech32("cosmos1qqqsyqcyq5rqwzqfys8f67", "x")
	s.Require().Error(err)
	s.Require().Equal("invalid Bech32 prefix; expected x, got cosmos", err.Error())
}
