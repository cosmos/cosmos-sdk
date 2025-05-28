package types

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type configTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(configTestSuite))
}

func (s *configTestSuite) TestConfig_SetPurpose() {
	config := NewConfig()
	config.SetPurpose(44)
	s.Require().Equal(uint32(44), config.GetPurpose())

	config.SetPurpose(0)
	s.Require().Equal(uint32(0), config.GetPurpose())

	config.Seal()
	s.Require().Panics(func() { config.SetPurpose(10) })
}

func (s *configTestSuite) TestConfig_SetCoinType() {
	config := NewConfig()
	config.SetCoinType(1)
	s.Require().Equal(uint32(1), config.GetCoinType())
	config.SetCoinType(99)
	s.Require().Equal(uint32(99), config.GetCoinType())

	config.Seal()
	s.Require().Panics(func() { config.SetCoinType(99) })
}

func (s *configTestSuite) TestConfig_SetTxEncoder() {
	mockErr := errors.New("test")
	config := NewConfig()
	s.Require().Nil(config.GetTxEncoder())
	encFunc := TxEncoder(func(tx Tx) ([]byte, error) { return nil, mockErr })
	config.SetTxEncoder(encFunc)
	_, err := config.GetTxEncoder()(Tx(nil))
	s.Require().Equal(mockErr, err)

	config.Seal()
	s.Require().Panics(func() { config.SetTxEncoder(encFunc) })
}

func (s *configTestSuite) TestConfig_SetFullFundraiserPath() {
	config := NewConfig()
	config.SetFullFundraiserPath("test/path")
	s.Require().Equal("test/path", config.GetFullFundraiserPath())

	config.SetFullFundraiserPath("test/poth")
	s.Require().Equal("test/poth", config.GetFullFundraiserPath())

	config.Seal()
	s.Require().Panics(func() { config.SetFullFundraiserPath("x/test/path") })
}

func (s *configTestSuite) TestKeyringServiceName() {
	s.Require().Equal(DefaultKeyringServiceName, KeyringServiceName())
}

func (s *configTestSuite) TestConfig_ScopePerBinary_DefaultBehavior() {
	cfg1 := GetConfig()
	cfg2 := GetConfig()
	s.Require().Equal(cfg1, cfg2, "configs should be identical in same binary by default")
}

func (s *configTestSuite) TestConfig_ScopePerBinary_EnvOverride() {
	s.T().Setenv(EnvConfigScope, "test-scope-A")
	cfgA := GetConfig()

	s.T().Setenv(EnvConfigScope, "test-scope-B")
	cfgB := GetConfig()

	s.Require().NotEqual(cfgA, cfgB, "configs should differ for different env scopes")
}

func (s *configTestSuite) TestConfig_ScopePerBinary_EnvRestoration() {
	envKey := EnvConfigScope

	s.T().Setenv(envKey, "test-scope-Restore")
	cfg1 := GetConfig()

	s.T().Setenv(envKey, "test-scope-Restore")
	cfg2 := GetConfig()

	s.Require().Equal(cfg1, cfg2, "config should remain stable with same env scope")
}

func (s *configTestSuite) TestConfig_ScopeKeyFormat() {
	key := getConfigKey()
	s.Require().True(strings.Count(key, "|") == 2, "scope key should have 2 pipe separators")
}
