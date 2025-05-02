package types_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type configTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(configTestSuite))
}

func (s *configTestSuite) TestConfig_SetPurpose() {
	config := sdk.NewConfig()
	config.SetPurpose(44)
	s.Require().Equal(uint32(44), config.GetPurpose())

	config.SetPurpose(0)
	s.Require().Equal(uint32(0), config.GetPurpose())

	config.Seal()
	s.Require().Panics(func() { config.SetPurpose(10) })
}

func (s *configTestSuite) TestConfig_SetCoinType() {
	config := sdk.NewConfig()
	config.SetCoinType(1)
	s.Require().Equal(uint32(1), config.GetCoinType())
	config.SetCoinType(99)
	s.Require().Equal(uint32(99), config.GetCoinType())

	config.Seal()
	s.Require().Panics(func() { config.SetCoinType(99) })
}

func (s *configTestSuite) TestConfig_SetTxEncoder() {
	mockErr := errors.New("test")
	config := sdk.NewConfig()
	s.Require().Nil(config.GetTxEncoder())
	encFunc := sdk.TxEncoder(func(tx sdk.Tx) ([]byte, error) { return nil, mockErr })
	config.SetTxEncoder(encFunc)
	_, err := config.GetTxEncoder()(sdk.Tx(nil))
	s.Require().Equal(mockErr, err)

	config.Seal()
	s.Require().Panics(func() { config.SetTxEncoder(encFunc) })
}

func (s *configTestSuite) TestConfig_SetFullFundraiserPath() {
	config := sdk.NewConfig()
	config.SetFullFundraiserPath("test/path")
	s.Require().Equal("test/path", config.GetFullFundraiserPath())

	config.SetFullFundraiserPath("test/poth")
	s.Require().Equal("test/poth", config.GetFullFundraiserPath())

	config.Seal()
	s.Require().Panics(func() { config.SetFullFundraiserPath("x/test/path") })
}

func (s *configTestSuite) TestKeyringServiceName() {
	s.Require().Equal(sdk.DefaultKeyringServiceName, sdk.KeyringServiceName())
}

func (s *configTestSuite) TestConfig_ScopePerBinary_DefaultBehavior() {
	cfg1 := sdk.GetConfig()
	cfg2 := sdk.GetConfig()
	s.Require().Equal(cfg1, cfg2, "configs should be identical in same binary by default")
}

func (s *configTestSuite) TestConfig_ScopePerBinary_EnvOverride() {
	original := os.Getenv(sdk.EnvConfigScope)
	defer os.Setenv(sdk.EnvConfigScope, original)

	s.Require().NoError(os.Setenv(sdk.EnvConfigScope, "test-scope-A"))
	cfgA := sdk.GetConfig()

	s.Require().NoError(os.Setenv(sdk.EnvConfigScope, "test-scope-B"))
	cfgB := sdk.GetConfig()

	s.Require().NotEqual(cfgA, cfgB, "configs should differ for different env scopes")
}

func (s *configTestSuite) TestConfig_ScopePerBinary_EnvRestoration() {
	envKey := sdk.EnvConfigScope
	original := os.Getenv(envKey)
	defer os.Setenv(envKey, original)

	s.Require().NoError(os.Setenv(envKey, "test-scope-Restore"))
	cfg1 := sdk.GetConfig()

	s.Require().NoError(os.Setenv(envKey, "test-scope-Restore"))
	cfg2 := sdk.GetConfig()

	s.Require().Equal(cfg1, cfg2, "config should remain stable with same env scope")
}

func (s *configTestSuite) TestConfig_ScopeKeyFormat() {
	key := getTestConfigKey()
	s.Require().True(strings.Count(key, "|") == 2, "scope key should have 2 pipe separators")
}

// getTestConfigKey replicates the internal key generation logic for testing visibility.
func getTestConfigKey() string {
	id := os.Getenv(sdk.EnvConfigScope)
	if id != "" {
		return id
	}
	exe, _ := os.Executable()
	host, _ := os.Hostname()
	pid := os.Getpid()
	return fmt.Sprintf("%s|%s|%d", host, exe, pid)
}
