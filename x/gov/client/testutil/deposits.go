package testutil

import (
	"fmt"
	"time"

	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]

	deposits := sdk.Coins{
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(0)),
		sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens),
		sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens.Sub(sdk.NewInt(50))),
	}
	s.deposits = deposits

	// create 2 proposals for testing
	for i := 0; i < len(deposits); i++ {
		id := i + 1
		deposit := deposits[i]

		s.submitProposal(val, deposit, id)
		s.proposalIDs = append(s.proposalIDs, fmt.Sprintf("%d", id))
	}
}

func (s *DepositTestSuite) SetupNewSuite() {
	s.T().Log("setting up new test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *DepositTestSuite) submitProposal(val *network.Validator, initialDeposit sdk.Coin, id int) {
	var exactArgs []string

	if !initialDeposit.IsZero() {
		exactArgs = append(exactArgs, fmt.Sprintf("--%s=%s", cli.FlagDeposit, initialDeposit.String()))
	}

	_, err := MsgSubmitLegacyProposal(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("Text Proposal %d", id),
		"Where is the title!?",
		v1beta1.ProposalTypeText,
		exactArgs...,
	)

	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *DepositTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.network.Cleanup()
}

func (s *DepositTestSuite) TestQueryDepositsWithoutInitialDeposit() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	proposalID := s.proposalIDs[0]

	// deposit amount
	depositAmount := sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String()
	_, err := MsgDeposit(clientCtx, val.Address.String(), proposalID, depositAmount)
	s.Require().NoError(err)

	// query deposit
	deposit := s.queryDeposit(val, proposalID, false, "")
	s.Require().NotNil(deposit)
	s.Require().Equal(sdk.Coins(deposit.Amount).String(), depositAmount)

	// query deposits
	deposits := s.queryDeposits(val, proposalID, false, "")
	s.Require().NotNil(deposits)
	s.Require().Len(deposits.Deposits, 1)
	// verify initial deposit
	s.Require().Equal(sdk.Coins(deposits.Deposits[0].Amount).String(), depositAmount)
}

func (s *DepositTestSuite) TestQueryDepositsWithInitialDeposit() {
	val := s.network.Validators[0]
	proposalID := s.proposalIDs[1]

	// query deposit
	deposit := s.queryDeposit(val, proposalID, false, "")
	s.Require().NotNil(deposit)
	s.Require().Equal(sdk.Coins(deposit.Amount).String(), s.deposits[1].String())

	// query deposits
	deposits := s.queryDeposits(val, proposalID, false, "")
	s.Require().NotNil(deposits)
	s.Require().Len(deposits.Deposits, 1)
	// verify initial deposit
	s.Require().Equal(sdk.Coins(deposits.Deposits[0].Amount).String(), s.deposits[1].String())
}

func (s *DepositTestSuite) TestQueryProposalAfterVotingPeriod() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	proposalID := s.proposalIDs[2]

	// query proposal
	args := []string{proposalID, fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	cmd := cli.GetCmdQueryProposal()
	_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)

	// waiting for deposit and voting period to end
	time.Sleep(20 * time.Second)

	// query proposal
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), fmt.Sprintf("proposal %s doesn't exist", proposalID))

	// query deposits
	deposits := s.queryDeposits(val, proposalID, true, "proposal 3 doesn't exist")
	s.Require().Nil(deposits)
}

func (s *DepositTestSuite) queryDeposits(val *network.Validator, proposalID string, exceptErr bool, message string) *v1.QueryDepositsResponse {
	args := []string{proposalID, fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositsRes *v1.QueryDepositsResponse
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

func (s *DepositTestSuite) queryDeposit(val *network.Validator, proposalID string, exceptErr bool, message string) *v1.Deposit {
	args := []string{proposalID, val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
	var depositRes *v1.Deposit
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
