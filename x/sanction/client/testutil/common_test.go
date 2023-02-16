package testutil

import (
	"encoding/json"
	"fmt"

	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg       network.Config
	network   *network.Network
	clientCtx client.Context

	commonArgs []string
	valAddr    sdk.AccAddress
	authority  string

	sanctionGenesis *sanction.GenesisState
}

func NewIntegrationTestSuite(cfg network.Config, sanctionGenesis *sanction.GenesisState) *IntegrationTestSuite {
	return &IntegrationTestSuite{
		cfg:             cfg,
		sanctionGenesis: sanctionGenesis,
	}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.clientCtx = s.network.Validators[0].ClientCtx
	s.valAddr = s.network.Validators[0].Address

	s.commonArgs = []string{
		fmt.Sprintf("--%s", flags.FlagFrom), s.valAddr.String(),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, s.bondCoins(10).String()),
	}

}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// assertErrorContents calls AssertErrorContents using this suite's t.
func (s *IntegrationTestSuite) assertErrorContents(theError error, contains []string, msgAndArgs ...interface{}) bool {
	return testutil.AssertErrorContents(s.T(), theError, contains, msgAndArgs...)
}

// bondCoin creates an sdk.Coin with the bond-denom in the amount provided.
func (s *IntegrationTestSuite) bondCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(s.cfg.BondDenom, amt)
}

// bondCoins creates an sdk.Coins with the bond-denom in the amount provided.
func (s *IntegrationTestSuite) bondCoins(amt int64) sdk.Coins {
	return sdk.NewCoins(s.bondCoin(amt))
}

// appendCommonFlagsTo adds this suite's common flags to the end of the provided arguments.
func (s *IntegrationTestSuite) appendCommonArgsTo(args ...string) []string {
	return append(args, s.commonArgs...)
}

func (s *IntegrationTestSuite) getAuthority() string {
	if len(s.authority) > 0 {
		return s.authority
	}
	args := []string{"gov", "--" + tmcli.OutputFlag, "json"}
	outBW, err := cli.ExecTestCLICmd(s.clientCtx, authcli.QueryModuleAccountByNameCmd(), args)
	s.Require().NoError(err, "ExecTestCLICmd q auth module-account gov")
	outBz := outBW.Bytes()
	s.T().Logf("q auth module-account gov output:\n%s", string(outBz))
	// example output:
	// {
	//   "account": {
	//     "@type": "/cosmos.auth.v1beta1.ModuleAccount",
	//     "base_account": {
	//       "address": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	//       "pub_key": null,
	//       "account_number": "9",
	//       "sequence": "0"
	//     },
	//     "name": "gov",
	//     "permissions": ["burner"]
	//   }
	// }
	var output map[string]json.RawMessage
	err = json.Unmarshal(outBz, &output)
	s.Require().NoError(err, "Unmarshal output json")
	var account map[string]json.RawMessage
	err = json.Unmarshal(output["account"], &account)
	s.Require().NoError(err, "Unmarshal account")
	var baseAccount map[string]string
	err = json.Unmarshal(account["base_account"], &baseAccount)
	s.Require().NoError(err, "Unmarshal base_account")
	s.authority = baseAccount["address"]
	s.T().Logf("authority: %q", s.authority)
	return s.authority
}
