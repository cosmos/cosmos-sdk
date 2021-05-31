package testutil

import (
	"fmt"
	"time"

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
	_, err = MsgDeposit(clientCtx, val.Address.String(), "1", sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(50)).String())
	s.Require().NoError(err)

	// waiting for voting period to end
	time.Sleep(20 * time.Second)

	// query deposit & verify initial deposit
	deposit := s.queryDeposit(val, "1", false)
	s.Require().Equal(deposit.Amount.String(), initialDeposit)

	// query deposits
	deposits := s.queryDeposits(val, "1", false)
	s.Require().Equal(len(deposits), 2)
	// verify initial deposit
	s.Require().Equal(deposits[0].Amount.String(), initialDeposit)
}

func (s *DepositTestSuite) TestQueryDepositsWithoutInitialDeposit() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// create a proposal without deposit
	_, err := MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 2", "Where is the title!?", types.ProposalTypeText)
	s.Require().NoError(err)

	// deposit amount
	_, err = MsgDeposit(clientCtx, val.Address.String(), "2", sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String())
	s.Require().NoError(err)

	// waiting for voting period to end
	time.Sleep(20 * time.Second)

	// query deposit
	deposit := s.queryDeposit(val, "2", false)
	s.Require().Equal(deposit.Amount.String(), sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String())

	// query deposits
	deposits := s.queryDeposits(val, "2", false)
	s.Require().Equal(len(deposits), 1)
	// verify initial deposit
	s.Require().Equal(deposits[0].Amount.String(), sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String())
}

func (s *DepositTestSuite) TestQueryProposalNotEnoughDeposits() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	initialDeposit := sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Sub(sdk.NewInt(2000))).String()

	// create a proposal with deposit
	_, err := MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 3", "Where is the title!?", types.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, initialDeposit))
	s.Require().NoError(err)

	// query proposal
	args := []string{"3", fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	cmd := cli.GetCmdQueryProposal()
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)

	// waiting for deposit period to end
	time.Sleep(20 * time.Second)

	// query proposal
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "proposal 3 doesn't exist")
}

func (s *DepositTestSuite) TestRejectedProposalDeposits() {
	val0 := s.network.Validators[0]
	clientCtx := val0.ClientCtx
	initialDeposit := sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens)

	// create a proposal with deposit
	_, err := MsgSubmitProposal(clientCtx, val0.Address.String(),
		"Text Proposal 4", "Where is the title!?", types.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, initialDeposit))
	s.Require().NoError(err)

	// vote
	_, err = MsgVote(clientCtx, val0.Address.String(), "4", "no")
	s.Require().NoError(err)

	time.Sleep(20 * time.Second)

	args := []string{"4", fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	cmd := cli.GetCmdQueryProposal()
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)

	// query deposits
	deposits := s.queryDeposits(val0, "4", false)
	s.Require().Equal(len(deposits), 1)
	// verify initial deposit
	s.Require().Equal(deposits[0].Amount.String(), initialDeposit.String())

}

func (s *DepositTestSuite) queryDeposits(val *network.Validator, proposalID string, exceptErr bool) types.Deposits {
	args := []string{proposalID, fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositsRes types.Deposits
	cmd := cli.GetCmdQueryDeposits()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	if exceptErr {
		s.Require().Error(err)
		return nil
	} else {
		s.Require().NoError(err)
		s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(out.Bytes(), &depositsRes))
		return depositsRes
	}
}

func (s *DepositTestSuite) queryDeposit(val *network.Validator, proposalID string, exceptErr bool) *types.Deposit {
	args := []string{proposalID, val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositRes types.Deposit
	cmd := cli.GetCmdQueryDeposit()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	if exceptErr {
		s.Require().Error(err)
		return nil
	} else {
		s.Require().NoError(err)
		s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(out.Bytes(), &depositRes))
		return &depositRes
	}
}
