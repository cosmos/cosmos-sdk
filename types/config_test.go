package types_test

import (
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
