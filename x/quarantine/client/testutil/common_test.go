package testutil

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"

	. "github.com/cosmos/cosmos-sdk/x/quarantine/testutil"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg       network.Config
	network   *network.Network
	clientCtx client.Context

	commonFlags []string
	valAddr     sdk.AccAddress
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, s.bondCoins(10).String()),
	}
	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.clientCtx = s.network.Validators[0].ClientCtx
	s.valAddr = s.network.Validators[0].Address
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) stopIfFailed() {
	if s.T().Failed() {
		s.T().FailNow()
	}
}

// bondCoin creates an sdk.Coin with the bond-denom in the amount provided.
func (s *IntegrationTestSuite) bondCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(s.cfg.BondDenom, amt)
}

// bondCoins creates an sdk.Coins with the bond-denom in the amount provided.
func (s *IntegrationTestSuite) bondCoins(amt int64) sdk.Coins {
	return sdk.NewCoins(s.bondCoin(amt))
}

// createAndFundAccount creates an account, adding the key to the keyring, funded with the provided amount of bond-denom coins.
func (s *IntegrationTestSuite) createAndFundAccount(index int, bondCoinAmt int64) string {
	memberNumber := uuid.New().String()

	info, _, err := s.clientCtx.Keyring.NewMnemonic(
		fmt.Sprintf("member%s", memberNumber),
		keyring.English, sdk.FullFundraiserPath,
		keyring.DefaultBIP39Passphrase, hd.Secp256k1,
	)
	s.Require().NoError(err, "NewMnemonic[%d]", index)

	pk, err := info.GetPubKey()
	s.Require().NoError(err, "GetPubKey[%d]", index)

	account := sdk.AccAddress(pk.Address())

	_, err = banktestutil.MsgSendExec(
		s.clientCtx,
		s.valAddr,
		account,
		s.bondCoins(bondCoinAmt),
		s.commonFlags...,
	)
	s.Require().NoError(err, "MsgSendExec[%d]", index)

	return account.String()
}

// appendCommonFlagsTo adds this suite's common flags to the end of the provided arguments.
func (s *IntegrationTestSuite) appendCommonFlagsTo(args ...string) []string {
	return append(args, s.commonFlags...)
}

// assertErrorContents calls AssertErrorContents using this suite's t.
func (s *IntegrationTestSuite) assertErrorContents(theError error, contains []string, msgAndArgs ...interface{}) bool {
	return AssertErrorContents(s.T(), theError, contains, msgAndArgs...)
}

var _ fmt.Stringer = asStringer("")

// asStringer is a string that has a String() function on it so that we can provide a string to MsgSendExec.
type asStringer string

// String implements the Stringer interface.
func (s asStringer) String() string {
	return string(s)
}
