package types_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"strings"
	"testing"
	"unsafe"



	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32" //nolint:staticcheck // we're using this to support the legacy way of dealing with bech32
)

const (
	pubStr     = "pub"
	valoper    = "valoper"
	valoperpub = "valoperpub"
	valcons    = "valcons"
	valconspub = "valconspub"
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

func (s *addressTestSuite) testMarshal(original, res interface{}, marshal func() ([]byte, error), unmarshal func([]byte) error) {
	bz, err := marshal()
	s.Require().Nil(err)
	s.Require().Nil(unmarshal(bz))
	s.Require().Equal(original, res)
}

func (s *addressTestSuite) testMarshalYAML(original, res interface{}, marshal func() (interface{}, error), unmarshal func([]byte) error) {
	bz, err := marshal()
	s.Require().Nil(err)
	s.Require().Nil(unmarshal([]byte(bz.(string))))
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
		_, err := rand.Read(pub.Key)
		s.Require().NoError(err)
		acc := types.AccAddress(pub.Address())
		res := types.AccAddress{}

		s.testMarshal(&acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		s.testMarshal(&acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err = types.AccAddressFromBech32(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)

		str = hex.EncodeToString(acc)
		res, err = types.AccAddressFromHexUnsafe(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)
	}

	for _, str := range invalidStrs {
		_, err := types.AccAddressFromHexUnsafe(str)
		s.Require().NotNil(err)

		_, err = types.AccAddressFromBech32(str)
		s.Require().NotNil(err)

		err = (*types.AccAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		s.Require().NotNil(err)
	}

	_, err := types.AccAddressFromHexUnsafe("")
	s.Require().Equal(types.ErrEmptyHexAddress, err)
}

func (s *addressTestSuite) TestRandBech32AccAddrConsistencyYAML() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	for i := 0; i < 1000; i++ {
		_, err := rand.Read(pub.Key)
		s.Require().NoError(err)
		acc := types.AccAddress(pub.Address())
		res := &types.AccAddress{}

		s.testMarshalYAML(&acc, res, acc.MarshalYAML, res.UnmarshalYAML)
		s.testMarshalYAML(&acc, res, acc.MarshalYAML, res.UnmarshalYAML)

		str := acc.String()
		*res, err = types.AccAddressFromBech32(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, *res)

		str = hex.EncodeToString(acc)
		*res, err = types.AccAddressFromHexUnsafe(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, *res)
	}

	for _, str := range invalidStrs {
		_, err := types.AccAddressFromHexUnsafe(str)
		s.Require().NotNil(err)

		_, err = types.AccAddressFromBech32(str)
		s.Require().NotNil(err)

		err = (*types.AccAddress)(nil).UnmarshalYAML([]byte("\"" + str + "\""))
		s.Require().NotNil(err)
	}

	_, err := types.AccAddressFromHexUnsafe("")
	s.Require().Equal(types.ErrEmptyHexAddress, err)
}

// Test that the account address cache ignores the bech32 prefix setting, retrieving bech32 addresses from the cache.
// This will cause the AccAddress.String() to print out unexpected prefixes if the config was changed between bech32 lookups.
// See https://github.com/cosmos/cosmos-sdk/issues/15317.
func (s *addressTestSuite) TestAddrCache() {
	// Use a random key
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}
	_, err := rand.Read(pub.Key)
	s.Require().NoError(err)
	// Set SDK bech32 prefixes to 'osmo'
	prefix := "osmo"
	conf := types.GetConfig()
	conf.SetBech32PrefixForAccount(prefix, prefix+pubStr)
	conf.SetBech32PrefixForValidator(prefix+valoper, prefix+valoperpub)
	conf.SetBech32PrefixForConsensusNode(prefix+valcons, prefix+valconspub)

	acc := types.AccAddress(pub.Address())
	osmoAddrBech32 := acc.String()

	// Set SDK bech32 to 'cosmos'
	prefix = "cosmos"
	conf.SetBech32PrefixForAccount(prefix, prefix+pubStr)
	conf.SetBech32PrefixForValidator(prefix+valoper, prefix+valoperpub)
	conf.SetBech32PrefixForConsensusNode(prefix+valcons, prefix+valconspub)

	// We name this 'addrCosmos' to prove a point, but the bech32 address will still begin with 'osmo' due to the cache behavior.
	addrCosmos := types.AccAddress(pub.Address())
	cosmosAddrBech32 := addrCosmos.String()

	// The default behavior will retrieve the bech32 address from the cache, ignoring the bech32 prefix change.
	s.Require().Equal(osmoAddrBech32, cosmosAddrBech32)
	s.Require().True(strings.HasPrefix(osmoAddrBech32, "osmo"))
	s.Require().True(strings.HasPrefix(cosmosAddrBech32, "osmo"))
}

// Test that the bech32 prefix is respected when the address cache is disabled.
// This causes AccAddress.String() to print out the expected prefixes if the config is changed between bech32 lookups.
// See https://github.com/cosmos/cosmos-sdk/issues/15317.
func (s *addressTestSuite) TestAddrCacheDisabled() {
	types.SetAddrCacheEnabled(false)

	// Use a random key
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}
	_, err := rand.Read(pub.Key)
	s.Require().NoError(err)
	// Set SDK bech32 prefixes to 'osmo'
	prefix := "osmo"
	conf := types.GetConfig()
	conf.SetBech32PrefixForAccount(prefix, prefix+pubStr)
	conf.SetBech32PrefixForValidator(prefix+valoper, prefix+valoperpub)
	conf.SetBech32PrefixForConsensusNode(prefix+valcons, prefix+valconspub)

	acc := types.AccAddress(pub.Address())
	osmoAddrBech32 := acc.String()

	// Set SDK bech32 to 'cosmos'
	prefix = "cosmos"
	conf.SetBech32PrefixForAccount(prefix, prefix+pubStr)
	conf.SetBech32PrefixForValidator(prefix+valoper, prefix+valoperpub)
	conf.SetBech32PrefixForConsensusNode(prefix+valcons, prefix+valconspub)

	addrCosmos := types.AccAddress(pub.Address())
	cosmosAddrBech32 := addrCosmos.String()

	// retrieve the bech32 address from the cache, respecting the bech32 prefix change.
	s.Require().NotEqual(osmoAddrBech32, cosmosAddrBech32)
	s.Require().True(strings.HasPrefix(osmoAddrBech32, "osmo"))
	s.Require().True(strings.HasPrefix(cosmosAddrBech32, "cosmos"))
}

func (s *addressTestSuite) TestValAddr() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	for i := 0; i < 20; i++ {
		_, err := rand.Read(pub.Key)
		s.Require().NoError(err)
		acc := types.ValAddress(pub.Address())
		res := types.ValAddress{}

		s.testMarshal(&acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		s.testMarshal(&acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err = types.ValAddressFromBech32(str)
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
	s.Require().Equal(types.ErrEmptyHexAddress, err)
}

func (s *addressTestSuite) TestConsAddress() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	for i := 0; i < 20; i++ {
		_, err := rand.Read(pub.Key[:])
		s.Require().NoError(err)
		acc := types.ConsAddress(pub.Address())
		res := types.ConsAddress{}

		s.testMarshal(&acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		s.testMarshal(&acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err = types.ConsAddressFromBech32(str)
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
	s.Require().Equal(types.ErrEmptyHexAddress, err)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[mathrand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (s *addressTestSuite) TestConfiguredPrefix() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}
	for length := 1; length < 10; length++ {
		for times := 1; times < 20; times++ {
			_, err := rand.Read(pub.Key[:])
			s.Require().NoError(err)
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

			bech32Pub := legacybech32.MustMarshalPubKey(legacybech32.AccPK, pub)
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

			bech32ValPub := legacybech32.MustMarshalPubKey(legacybech32.ValPK, pub)
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

			bech32ConsPub := legacybech32.MustMarshalPubKey(legacybech32.ConsPK, pub)
			s.Require().True(strings.HasPrefix(
				bech32ConsPub,
				prefix+types.PrefixConsensus+types.PrefixPublic))
		}
	}
}

func (s *addressTestSuite) TestAddressInterface() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}
	_, err := rand.Read(pub.Key)
	s.Require().NoError(err)
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
	s.Require().True(accAddr1.Equals(accAddr1))   //nolint:gocritic // checking if these are the same
	s.Require().True(consAddr1.Equals(consAddr1)) //nolint:gocritic // checking if these are the same
	s.Require().True(valAddr1.Equals(valAddr1))   //nolint:gocritic // checking if these are the same

	// emptiness
	s.Require().True(types.AccAddress{}.Equals(types.AccAddress{})) //nolint:gocritic // checking if these are the same
	s.Require().True(types.AccAddress{}.Equals(types.AccAddress(nil)))
	s.Require().True(types.AccAddress(nil).Equals(types.AccAddress{}))
	s.Require().True(types.AccAddress(nil).Equals(types.AccAddress(nil))) //nolint:gocritic // checking if these are the same

	s.Require().True(types.ConsAddress{}.Equals(types.ConsAddress{})) //nolint:gocritic // checking if these are the same
	s.Require().True(types.ConsAddress{}.Equals(types.ConsAddress(nil)))
	s.Require().True(types.ConsAddress(nil).Equals(types.ConsAddress{}))
	s.Require().True(types.ConsAddress(nil).Equals(types.ConsAddress(nil))) //nolint:gocritic // checking if these are the same

	s.Require().True(types.ValAddress{}.Equals(types.ValAddress{})) //nolint:gocritic // checking if these are the same
	s.Require().True(types.ValAddress{}.Equals(types.ValAddress(nil)))
	s.Require().True(types.ValAddress(nil).Equals(types.ValAddress{}))
	s.Require().True(types.ValAddress(nil).Equals(types.ValAddress(nil))) //nolint:gocritic // checking if these are the same

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
	s.Require().Equal("decoding Bech32 address failed: must provide a non empty address", err.Error())
	_, err = types.GetFromBech32("cosmos1qqqsyqcyq5rqwzqfys8f67", "x")
	s.Require().Error(err)
	s.Require().Equal("invalid Bech32 prefix; expected x, got cosmos", err.Error())
}

func (s *addressTestSuite) TestMustValAddressFromBech32() {
	bech32PrefixValAddr := types.GetConfig().GetBech32ValidatorAddrPrefix()
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	address := types.MustBech32ifyAddressBytes(bech32PrefixValAddr, addr20byte)

	valAddress1, err := types.ValAddressFromBech32(address)
	s.Require().Nil(err)

	valAddress2 := types.MustValAddressFromBech32(address)

	s.Require().Equal(valAddress1, valAddress2)
}

func (s *addressTestSuite) TestMustValAddressFromBech32Panic() {
	s.Require().Panics(func() {
		types.MustValAddressFromBech32("")
	})
}

func (s *addressTestSuite) TestGetBech32PrefixAccPub() {
	actual := types.GetBech32PrefixAccPub("")
	s.Require().Equal(types.PrefixPublic, actual)

	actual = types.GetBech32PrefixAccPub("cosmos")
	expected := "cosmos" + types.PrefixPublic
	s.Require().Equal(expected, actual)
}

func (s *addressTestSuite) TestGetBech32PrefixValAddress() {
	actual := types.GetBech32PrefixValAddr("")
	expected := types.PrefixValidator + types.PrefixOperator
	s.Require().Equal(expected, actual)

	actual = types.GetBech32PrefixValAddr("cosmos1")
	expected = "cosmos1" + types.PrefixValidator + types.PrefixOperator
	s.Require().Equal(expected, actual)
}

func (s *addressTestSuite) TestGetBech32PrefixValPub() {
	actual := types.GetBech32PrefixValPub("")
	expected := types.PrefixValidator + types.PrefixOperator + types.PrefixPublic
	s.Require().Equal(expected, actual)

	actual = types.GetBech32PrefixValPub("cosmos2")
	expected = "cosmos2" + types.PrefixValidator + types.PrefixOperator + types.PrefixPublic
	s.Require().Equal(expected, actual)
}

func (s *addressTestSuite) TestGetBech32PrefixConsAddr() {
	actual := types.GetBech32PrefixConsAddr("")
	expected := types.PrefixValidator + types.PrefixConsensus
	s.Require().Equal(expected, actual)

	actual = types.GetBech32PrefixConsAddr("cosmos3")
	expected = "cosmos3" + types.PrefixValidator + types.PrefixConsensus
	s.Require().Equal(expected, actual)
}

func (s *addressTestSuite) TestGetBech32PrefixConsPub() {
	actual := types.GetBech32PrefixConsPub("")
	expected := types.PrefixValidator + types.PrefixConsensus + types.PrefixPublic
	s.Require().Equal(expected, actual)

	actual = types.GetBech32PrefixConsPub("cosmos4")
	expected = "cosmos4" + types.PrefixValidator + types.PrefixConsensus + types.PrefixPublic
	s.Require().Equal(expected, actual)
}

func (s *addressTestSuite) TestMustAccAddressFromBech32() {
	bech32PrefixAccAddr := types.GetConfig().GetBech32AccountAddrPrefix()
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	address := types.MustBech32ifyAddressBytes(bech32PrefixAccAddr, addr20byte)

	accAddress1, err := types.AccAddressFromBech32(address)
	s.Require().NoError(err)

	accAddress2 := types.MustAccAddressFromBech32(address)
	s.Require().Equal(accAddress1, accAddress2)
}

func (s *addressTestSuite) TestMustAccAddressFromBech32Panic() {
	s.Require().Panics(func() {
		types.MustAccAddressFromBech32("no-valid")
	})
}

func (s *addressTestSuite) TestUnmarshalJSONAccAddressFailed() {
	addr := &types.AccAddress{}
	err := addr.UnmarshalJSON(nil)
	s.Require().Error(err)
}

func (s *addressTestSuite) TestUnmarshalJSONAccAddressWithEmptyString() {
	addr := &types.AccAddress{}
	err := addr.UnmarshalJSON([]byte{34, 34})
	s.Require().NoError(err)
	s.Require().Equal(&types.AccAddress{}, addr)
}

func (s *addressTestSuite) TestUnmarshalYAMLAccAddressFailed() {
	malformedYAML := []byte("k:k:K:")
	addr := &types.AccAddress{}
	err := addr.UnmarshalYAML(malformedYAML)
	s.Require().Error(err)
}

func (s *addressTestSuite) TestUnmarshalYAMLAccAddressWithEmptyString() {
	addr := &types.AccAddress{}
	err := addr.UnmarshalYAML([]byte{34, 34})
	s.Require().NoError(err)
	s.Require().Equal(&types.AccAddress{}, addr)
}

func (s *addressTestSuite) TestFormatAccAddressAsString() {
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	accAddr := types.AccAddress(addr20byte)

	actual := fmt.Sprintf("%s", accAddr) // this will call internally  the method AccAddress.Format

	hrp := types.GetConfig().GetBech32AccountAddrPrefix()
	expected, err := bech32.ConvertAndEncode(hrp, addr20byte)
	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
}

func (s *addressTestSuite) TestFormatAccAddressAsPointer() {
	bech32PrefixAccAddr := types.GetConfig().GetBech32AccountAddrPrefix()
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	address := types.MustBech32ifyAddressBytes(bech32PrefixAccAddr, addr20byte)
	accAddr := types.MustAccAddressFromBech32(address)

	ptrAddr := &accAddr
	actual := fmt.Sprintf("%p", ptrAddr) // this will call internally  the method AccAddress.Format
	expected := uintptr(unsafe.Pointer(&accAddr))
	s.Require().Equal(fmt.Sprintf("0x%x", expected), actual)
}

func (s *addressTestSuite) TestFormatAccAddressWhenVerbIsDifferentFromSOrP() {
	bech32PrefixAccAddr := types.GetConfig().GetBech32AccountAddrPrefix()
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	address := types.MustBech32ifyAddressBytes(bech32PrefixAccAddr, addr20byte)
	accAddr := types.MustAccAddressFromBech32(address)

	// When the verb is different from 's' or 'p', then the address is always formatted to base 16, with uppercase letters for A-F
	expected := fmt.Sprintf("%X", addr20byte)

	// GENERAL
	actual := fmt.Sprintf("%v", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%#v", accAddr)
	s.Require().Equal(expected, actual)

	// BOOLEAN
	actual = fmt.Sprintf("%t", accAddr)
	s.Require().Equal(expected, actual)

	// INTEGERS
	actual = fmt.Sprintf("%b", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%c", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%d", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%o", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%O", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%x", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%X", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%U", accAddr)
	s.Require().Equal(expected, actual)

	// Floating-point and complex constituents
	actual = fmt.Sprintf("%e", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%E", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%f", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%F", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%g", accAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%G", accAddr)
}

func (s *addressTestSuite) TestUnmarshalJSONValAddressFailed() {
	addr := &types.ValAddress{}
	err := addr.UnmarshalJSON(nil)
	s.Require().Error(err)
}

func (s *addressTestSuite) TestUnmarshalJSONValAddressWithEmptyString() {
	addr := &types.ValAddress{}
	err := addr.UnmarshalJSON([]byte{34, 34})
	s.Require().NoError(err)
	s.Require().Equal(&types.ValAddress{}, addr)
}

func (s *addressTestSuite) TestUnmarshalYAMLValAddressFailed() {
	malformedYAML := []byte("k:k:K:")
	addr := &types.ValAddress{}
	err := addr.UnmarshalYAML(malformedYAML)
	s.Require().Error(err)
}

func (s *addressTestSuite) TestUnmarshalYAMLValAddressWithEmptyString() {
	addr := &types.ValAddress{}
	err := addr.UnmarshalYAML([]byte{34, 34})
	s.Require().NoError(err)
	s.Require().Equal(&types.ValAddress{}, addr)
}

func (s *addressTestSuite) TestFormatValAddressAsString() {
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	accAddr := types.ValAddress(addr20byte)

	actual := fmt.Sprintf("%s", accAddr)

	hrp := types.GetConfig().GetBech32ValidatorAddrPrefix()
	expected, err := bech32.ConvertAndEncode(hrp, addr20byte)
	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
}

func (s *addressTestSuite) TestFormatValAddressAsPointer() {
	bech32PrefixValAddr := types.GetConfig().GetBech32ValidatorAddrPrefix()
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	address := types.MustBech32ifyAddressBytes(bech32PrefixValAddr, addr20byte)
	accAddr := types.MustValAddressFromBech32(address)

	ptrAddr := &accAddr
	actual := fmt.Sprintf("%p", ptrAddr)
	expected := uintptr(unsafe.Pointer(&accAddr))
	s.Require().Equal(fmt.Sprintf("0x%x", expected), actual)
}

func (s *addressTestSuite) TestFormatValAddressWhenVerbIsDifferentFromSOrP() {
	bech32PrefixValAddr := types.GetConfig().GetBech32ValidatorAddrPrefix()
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	address := types.MustBech32ifyAddressBytes(bech32PrefixValAddr, addr20byte)
	valAddr := types.MustValAddressFromBech32(address)

	// When the verb is different from 's' or 'p', then the address is always formatted to base 16, with uppercase letters for A-F
	expected := fmt.Sprintf("%X", types.AccAddress(addr20byte))

	// GENERAL
	actual := fmt.Sprintf("%v", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%#v", valAddr)
	s.Require().Equal(expected, actual)

	// BOOLEAN
	actual = fmt.Sprintf("%t", valAddr)
	s.Require().Equal(expected, actual)

	// INTEGERS
	actual = fmt.Sprintf("%b", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%c", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%d", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%o", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%O", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%x", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%X", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%U", valAddr)
	s.Require().Equal(expected, actual)

	// Floating-point and complex constituents
	actual = fmt.Sprintf("%e", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%E", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%f", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%F", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%g", valAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%G", valAddr)
}

func (s *addressTestSuite) TestUnmarshalJSONConsAddressFailed() {
	addr := &types.ConsAddress{}
	err := addr.UnmarshalJSON(nil)
	s.Require().Error(err)
}

func (s *addressTestSuite) TestUnmarshalJSONConsAddressWithEmptyString() {
	addr := &types.ConsAddress{}
	err := addr.UnmarshalJSON([]byte{34, 34})
	s.Require().NoError(err)
	s.Require().Equal(&types.ConsAddress{}, addr)
}

func (s *addressTestSuite) TestUnmarshalYAMLConsAddressFailed() {
	malformedYAML := []byte("k:k:K:")
	addr := &types.ConsAddress{}
	err := addr.UnmarshalYAML(malformedYAML)
	s.Require().Error(err)
}

func (s *addressTestSuite) TestUnmarshalYAMLConsAddressWithEmptyString() {
	addr := &types.ConsAddress{}
	err := addr.UnmarshalYAML([]byte{34, 34})
	s.Require().NoError(err)
	s.Require().Equal(&types.ConsAddress{}, addr)
}

func (s *addressTestSuite) TestFormatConsAddressAsString() {
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	consAddr := types.ConsAddress(addr20byte)

	actual := fmt.Sprintf("%s", consAddr)

	hrp := types.GetConfig().GetBech32ConsensusAddrPrefix()
	expected, err := bech32.ConvertAndEncode(hrp, addr20byte)
	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
}

func (s *addressTestSuite) TestFormatConsAddressAsPointer() {
	bech32PrefixConsAddr := types.GetConfig().GetBech32ConsensusAddrPrefix()
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	address := types.MustBech32ifyAddressBytes(bech32PrefixConsAddr, addr20byte)

	consAddr, err := types.ConsAddressFromBech32(address)
	s.Require().NoError(err)

	ptrAddr := &consAddr
	actual := fmt.Sprintf("%p", ptrAddr)
	expected := uintptr(unsafe.Pointer(&consAddr))
	s.Require().Equal(fmt.Sprintf("0x%x", expected), actual)
}

func (s *addressTestSuite) TestFormatConsAddressWhenVerbIsDifferentFromSOrP() {
	bech32PrefixAccAddr := types.GetConfig().GetBech32ConsensusAddrPrefix()
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	address := types.MustBech32ifyAddressBytes(bech32PrefixAccAddr, addr20byte)
	consAddr, err := types.ConsAddressFromBech32(address)
	s.Require().NoError(err)

	// When the verb is different from 's' or 'p', then the address is always formatted to base 16, with uppercase letters for A-F
	expected := fmt.Sprintf("%X", types.AccAddress(addr20byte))

	// GENERAL
	actual := fmt.Sprintf("%v", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%#v", consAddr)
	s.Require().Equal(expected, actual)

	// BOOLEAN
	actual = fmt.Sprintf("%t", consAddr)
	s.Require().Equal(expected, actual)

	// INTEGERS
	actual = fmt.Sprintf("%b", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%c", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%d", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%o", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%O", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%x", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%X", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%U", consAddr)
	s.Require().Equal(expected, actual)

	// Floating-point and complex constituents
	actual = fmt.Sprintf("%e", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%E", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%f", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%F", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%g", consAddr)
	s.Require().Equal(expected, actual)
	actual = fmt.Sprintf("%G", consAddr)
}
