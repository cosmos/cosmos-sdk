// +build norace

package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

type DepositTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewDepositTestSuite(cfg network.Config) *DepositTestSuite {
	return &DepositTestSuite{cfg: cfg}
}

func (s *DepositTestSuite) SetupSuite() {
	s.T().Log("setting up test suite")

	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

}

func (s *DepositTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.network.Cleanup()
}

func (s *DepositTestSuite) TestQueryWithInitialDeposit() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// create a proposal with deposit
	_, err := MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 1", "Where is the title!?", types.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Sub(sdk.NewInt(20))).String()))
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	// deposit more amount
	args1 := []string{
		"1",
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(50)).String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(20))).String()),
	}
	cmd := cli.NewCmdDeposit()
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args1)
	s.Require().NoError(err)

	// waiting for proposal to expires
	time.Sleep(30 * time.Second)

	args := []string{"1", val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositRes types.Deposit
	cmd = cli.GetCmdQueryDeposit()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), &depositRes))
	s.Require().Equal(depositRes.Amount.String(), sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(30))).String())
}

func (s *DepositTestSuite) TestQueryWithoutInitialDeposit() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// create a proposal without deposit
	_, err := MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 2", "Where is the title!?", types.ProposalTypeText)
	s.Require().NoError(err)

	// deposit amount
	args1 := []string{
		"2",
		sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(20))).String()),
	}
	cmd := cli.NewCmdDeposit()
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args1)
	s.Require().NoError(err)

	// waiting for proposal to expires
	time.Sleep(30 * time.Second)

	args := []string{"2", val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositRes types.Deposit
	cmd = cli.GetCmdQueryDeposit()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), &depositRes))
	s.Require().Equal(depositRes.Amount.String(), sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String())
}

func TestDepositTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	cfg.NumValidators = 1
	genesisState := types.DefaultGenesisState()
	genesisState.DepositParams = types.NewDepositParams(sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, types.DefaultMinDepositTokens)), time.Duration(30)*time.Second)
	bz, err := cfg.Codec.MarshalJSON(genesisState)
	require.NoError(t, err)
	cfg.GenesisState["gov"] = bz
	suite.Run(t, NewDepositTestSuite(cfg))
}
