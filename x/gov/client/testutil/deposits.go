package testutil

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

type DepositTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
	fees    string
}

func NewDepositTestSuite(cfg network.Config) *DepositTestSuite {
	return &DepositTestSuite{cfg: cfg}
}

func (s *DepositTestSuite) SetupSuite() {
	s.T().Log("setting up test suite")

	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
	s.fees = sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(20))).String()

}

func (s *DepositTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.network.Cleanup()
}

func (s *DepositTestSuite) TestQueryDepositsInitialDeposit() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	initialDeposit := sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Sub(sdk.NewInt(20))).String()

	// create a proposal with deposit
	_, err := MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 1", "Where is the title!?", types.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, initialDeposit))
	s.Require().NoError(err)

	// deposit more amount
	args := []string{
		"1",
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(50)).String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, s.fees),
	}
	cmd := cli.NewCmdDeposit()
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)

	// waiting for voting period to end
	time.Sleep(20 * time.Second)

	// query deposit
	args = []string{"1", val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositRes types.Deposit
	cmd = cli.GetCmdQueryDeposit()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(out.Bytes(), &depositRes))

	// verify initial deposit
	s.Require().Equal(depositRes.Amount.String(), initialDeposit)

	// query deposits
	args = []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositsRes types.Deposits
	cmd = cli.GetCmdQueryDeposits()
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(out.Bytes(), &depositsRes))
	s.Require().Equal(len(depositsRes), 2)
	// verify initial deposit
	s.Require().Equal(depositsRes[0].Amount.String(), initialDeposit)
}

func (s *DepositTestSuite) TestQueryDepositsWithoutInitialDeposit() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// create a proposal without deposit
	_, err := MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 2", "Where is the title!?", types.ProposalTypeText)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	// deposit amount
	args := []string{
		"2",
		sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, s.fees),
	}
	cmd := cli.NewCmdDeposit()
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)

	// waiting for voting period to end
	time.Sleep(20 * time.Second)

	// query deposit
	var depositRes types.Deposit
	args = []string{"2", val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	cmd = cli.GetCmdQueryDeposit()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(out.Bytes(), &depositRes))
	s.Require().Equal(depositRes.Amount.String(), sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String())

	// query deposits
	args = []string{"2", fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositsRes types.Deposits
	cmd = cli.GetCmdQueryDeposits()
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(out.Bytes(), &depositsRes))
	s.Require().Equal(len(depositsRes), 1)
	// verify initial deposit
	s.Require().Equal(depositsRes[0].Amount.String(), sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String())
}
