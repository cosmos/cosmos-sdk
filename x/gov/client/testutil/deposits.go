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

	cfg         network.Config
	network     *network.Network
	deposits    sdk.Coins
	proposalIDs []string
}

func NewDepositTestSuite(cfg network.Config) *DepositTestSuite {
	return &DepositTestSuite{cfg: cfg}
}

func (s *DepositTestSuite) SetupSuite() {
	s.T().Log("setting up test suite")

	s.network = network.New(s.T(), s.cfg)
	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]

	deposits := sdk.Coins{
		sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Sub(sdk.NewInt(120))),
		sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Sub(sdk.NewInt(50))),
		sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens),
	}
	s.deposits = deposits

	// create 4 proposals for testing
	for i := 0; i < 4; i++ {
		var exactArgs []string
		id := i + 1

		if !deposits[i].IsZero() {
			exactArgs = append(exactArgs, fmt.Sprintf("--%s=%s", cli.FlagDeposit, deposits[i]))
		}

		_, err := MsgSubmitProposal(
			val.ClientCtx,
			val.Address.String(),
			fmt.Sprintf("Text Proposal %d", id),
			"Where is the title!?",
			types.ProposalTypeText,
			exactArgs...,
		)

		s.Require().NoError(err)
		_, err = s.network.WaitForHeight(1)
		s.Require().NoError(err)
		s.proposalIDs = append(s.proposalIDs, fmt.Sprintf("%d", id))
	}
}

func (s *DepositTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.network.Cleanup()
}

func (s *DepositTestSuite) TestQueryDepositsWithoutInitialDeposit() {
	val := s.network.Validators[0]
	// val2 := s.network.Validators[1]
	clientCtx := val.ClientCtx
	proposalID := s.proposalIDs[1]

	// deposit amount
	depositAmount := sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String()
	_, err := MsgDeposit(clientCtx, val.Address.String(), proposalID, depositAmount)
	s.Require().NoError(err)

	// waiting for voting period to end
	// time.Sleep(20 * time.Second)
	_, err = s.network.WaitForHeight(2)
	s.Require().NoError(err)

	// query deposit
	deposit := s.queryDeposit(val, proposalID, false, "")
	s.Require().NotNil(deposit)
	s.Require().Equal(deposit.Amount.String(), depositAmount)

	// query deposits
	deposits := s.queryDeposits(val, proposalID, false, "")
	s.Require().NotNil(deposits)
	s.Require().Len(deposits.Deposits, 1)
	// verify initial deposit
	s.Require().Equal(deposits.Deposits[0].Amount.String(), depositAmount)
}

func (s *DepositTestSuite) TestQueryProposalNotEnoughDeposits() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	proposalID := s.proposalIDs[2]

	// query proposal
	args := []string{proposalID, fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	cmd := cli.GetCmdQueryProposal()
	_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)

	// waiting for deposit period to end
	time.Sleep(20 * time.Second)

	// query proposal
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), fmt.Sprintf("proposal %s doesn't exist", proposalID))
}

func (s *DepositTestSuite) TestRejectedProposalDeposits() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	initialDeposit := s.deposits[3]
	proposalID := s.proposalIDs[3]

	// query deposits
	var deposits types.QueryDepositsResponse
	args := []string{proposalID, fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	cmd := cli.GetCmdQueryDeposits()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(out.Bytes(), &deposits))
	s.Require().Equal(len(deposits.Deposits), 1)
	// verify initial deposit
	s.Require().Equal(deposits.Deposits[0].Amount.String(), sdk.NewCoin(s.cfg.BondDenom, types.DefaultMinDepositTokens).String())

	// vote
	_, err = MsgVote(clientCtx, val.Address.String(), proposalID, "no")
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(3)
	s.Require().NoError(err)

	// time.Sleep(20 * time.Second)

	args = []string{proposalID, fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	cmd = cli.GetCmdQueryProposal()
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)

	// query deposits
	depositsRes := s.queryDeposits(val, proposalID, false, "")
	s.Require().NotNil(depositsRes)
	s.Require().Len(depositsRes.Deposits, 1)
	// verify initial deposit
	s.Require().Equal(depositsRes.Deposits[0].Amount.String(), initialDeposit.String())
}

func (s *DepositTestSuite) queryDeposits(val *network.Validator, proposalID string, exceptErr bool, message string) *types.QueryDepositsResponse {
	args := []string{proposalID, fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositsRes *types.QueryDepositsResponse
	cmd := cli.GetCmdQueryDeposits()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)

	if exceptErr {
		s.Require().Error(err)
		s.Require().Contains(err.Error(), message)
		return nil
	}

	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(out.Bytes(), &depositsRes))
	return depositsRes
}

func (s *DepositTestSuite) queryDeposit(val *network.Validator, proposalID string, exceptErr bool, message string) *types.Deposit {
	args := []string{proposalID, val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositRes *types.Deposit
	cmd := cli.GetCmdQueryDeposit()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	if exceptErr {
		s.Require().Error(err)
		s.Require().Contains(err.Error(), message)
		return nil
	}
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(out.Bytes(), &depositRes))
	return depositRes
}
