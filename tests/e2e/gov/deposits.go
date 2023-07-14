package gov

import (
	"fmt"
	"strconv"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govclitestutil "github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
}

func (s *DepositTestSuite) submitProposal(val *network.Validator, initialDeposit sdk.Coin, name string) uint64 {
	var exactArgs []string

	if !initialDeposit.IsZero() {
		exactArgs = append(exactArgs, fmt.Sprintf("--%s=%s", cli.FlagDeposit, initialDeposit.String()))
	}

	_, err := govclitestutil.MsgSubmitLegacyProposal(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("Text Proposal %s", name),
		"Where is the title!?",
		v1beta1.ProposalTypeText,
		exactArgs...,
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// query proposals, return the last's id
	res, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/gov/v1/proposals", val.APIAddress))
	s.Require().NoError(err)
	var proposals v1.QueryProposalsResponse
	err = s.cfg.Codec.UnmarshalJSON(res, &proposals)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(proposals.Proposals), 1)

	return proposals.Proposals[len(proposals.Proposals)-1].Id
}

func (s *DepositTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.network.Cleanup()
}

func (s *DepositTestSuite) TestQueryDepositsWithoutInitialDeposit() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	// submit proposal without initial deposit
	id := s.submitProposal(val, sdk.NewCoin(s.cfg.BondDenom, math.NewInt(0)), "TestQueryDepositsWithoutInitialDeposit")
	proposalID := strconv.FormatUint(id, 10)

	// deposit amount
	depositAmount := sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens.Add(math.NewInt(50))).String()
	_, err := govclitestutil.MsgDeposit(clientCtx, val.Address.String(), proposalID, depositAmount)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// query deposit
	deposit := s.queryDeposit(val, proposalID, false, "")
	s.Require().NotNil(deposit)
	s.Require().Equal(depositAmount, sdk.Coins(deposit.Deposit.Amount).String())

	// query deposits
	deposits := s.queryDeposits(val, proposalID, false, "")
	s.Require().NotNil(deposits)
	s.Require().Len(deposits.Deposits, 1)
	// verify initial deposit
	s.Require().Equal(depositAmount, sdk.Coins(deposits.Deposits[0].Amount).String())
}

func (s *DepositTestSuite) TestQueryDepositsWithInitialDeposit() {
	val := s.network.Validators[0]
	depositAmount := sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens)

	// submit proposal with an initial deposit
	id := s.submitProposal(val, depositAmount, "TestQueryDepositsWithInitialDeposit")
	proposalID := strconv.FormatUint(id, 10)

	// query deposit
	deposit := s.queryDeposit(val, proposalID, false, "")
	s.Require().NotNil(deposit)
	s.Require().Equal(depositAmount.String(), sdk.Coins(deposit.Deposit.Amount).String())

	// query deposits
	deposits := s.queryDeposits(val, proposalID, false, "")
	s.Require().NotNil(deposits)
	s.Require().Len(deposits.Deposits, 1)
	// verify initial deposit
	s.Require().Equal(depositAmount.String(), sdk.Coins(deposits.Deposits[0].Amount).String())
}

func (s *DepositTestSuite) TestQueryProposalAfterVotingPeriod() {
	val := s.network.Validators[0]
	depositAmount := sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens.Sub(math.NewInt(50)))

	// submit proposal with an initial deposit
	id := s.submitProposal(val, depositAmount, "TestQueryProposalAfterVotingPeriod")
	proposalID := strconv.FormatUint(id, 10)

	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/gov/v1/proposals", val.APIAddress))
	s.Require().NoError(err)
	var proposals v1.QueryProposalsResponse
	err = s.cfg.Codec.UnmarshalJSON(resp, &proposals)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(proposals.Proposals), 1)

	// query proposal
	resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s", val.APIAddress, proposalID))
	s.Require().NoError(err)
	var proposal v1.QueryProposalResponse
	err = s.cfg.Codec.UnmarshalJSON(resp, &proposal)
	s.Require().NoError(err)

	// waiting for deposit and voting period to end
	time.Sleep(25 * time.Second)

	// query proposal
	resp, err = testutil.GetRequest(fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s", val.APIAddress, proposalID))
	s.Require().NoError(err)
	s.Require().Contains(string(resp), fmt.Sprintf("proposal %s doesn't exist", proposalID))

	// query deposits
	deposits := s.queryDeposits(val, proposalID, false, "")
	s.Require().Len(deposits.Deposits, 0)
}

func (s *DepositTestSuite) queryDeposits(val *network.Validator, proposalID string, exceptErr bool, message string) *v1.QueryDepositsResponse {
	s.Require().NoError(s.network.WaitForNextBlock())

	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/deposits", val.APIAddress, proposalID))
	s.Require().NoError(err)

	if exceptErr {
		s.Require().Contains(string(resp), message)
		return nil
	}

	var depositsRes v1.QueryDepositsResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(resp, &depositsRes)
	s.Require().NoError(err)

	return &depositsRes
}

func (s *DepositTestSuite) queryDeposit(val *network.Validator, proposalID string, exceptErr bool, message string) *v1.QueryDepositResponse {
	s.Require().NoError(s.network.WaitForNextBlock())

	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s/deposits/%s", val.APIAddress, proposalID, val.Address.String()))
	s.Require().NoError(err)

	if exceptErr {
		s.Require().Contains(string(resp), message)
		return nil
	}

	var depositRes v1.QueryDepositResponse
	err = val.ClientCtx.Codec.UnmarshalJSON(resp, &depositRes)
	s.Require().NoError(err)

	return &depositRes
}
