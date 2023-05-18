package gov

import (
	"fmt"
	"strconv"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
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
	cmd := cli.GetCmdQueryProposals(address.NewBech32Codec("cosmos"))
	args := []string{fmt.Sprintf("--%s=json", flags.FlagOutput)}
	res, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().NoError(err)

	var proposals v1.QueryProposalsResponse
	err = s.cfg.Codec.UnmarshalJSON(res.Bytes(), &proposals)
	s.Require().NoError(err)

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
	id := s.submitProposal(val, sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(0)), "TestQueryDepositsWithoutInitialDeposit")
	proposalID := strconv.FormatUint(id, 10)

	// deposit amount
	depositAmount := sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens.Add(sdk.NewInt(50))).String()
	_, err := govclitestutil.MsgDeposit(clientCtx, val.Address.String(), proposalID, depositAmount)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

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
	depositAmount := sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens)

	// submit proposal with an initial deposit
	id := s.submitProposal(val, depositAmount, "TestQueryDepositsWithInitialDeposit")
	proposalID := strconv.FormatUint(id, 10)

	// query deposit
	deposit := s.queryDeposit(val, proposalID, false, "")
	s.Require().NotNil(deposit)
	s.Require().Equal(sdk.Coins(deposit.Amount).String(), depositAmount.String())

	// query deposits
	deposits := s.queryDeposits(val, proposalID, false, "")
	s.Require().NotNil(deposits)
	s.Require().Len(deposits.Deposits, 1)
	// verify initial deposit
	s.Require().Equal(sdk.Coins(deposits.Deposits[0].Amount).String(), depositAmount.String())
}

func (s *DepositTestSuite) TestQueryProposalAfterVotingPeriod() {
	val := s.network.Validators[0]
	depositAmount := sdk.NewCoin(s.cfg.BondDenom, v1.DefaultMinDepositTokens.Sub(sdk.NewInt(50)))

	// submit proposal with an initial deposit
	id := s.submitProposal(val, depositAmount, "TestQueryProposalAfterVotingPeriod")
	proposalID := strconv.FormatUint(id, 10)

	args := []string{fmt.Sprintf("--%s=json", flags.FlagOutput)}
	cmd := cli.GetCmdQueryProposals(address.NewBech32Codec("cosmos"))
	_, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().NoError(err)

	// query proposal
	args = []string{proposalID, fmt.Sprintf("--%s=json", flags.FlagOutput)}
	cmd = cli.GetCmdQueryProposal()
	_, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().NoError(err)

	// waiting for deposit and voting period to end
	time.Sleep(25 * time.Second)

	// query proposal
	_, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), fmt.Sprintf("proposal %s doesn't exist", proposalID))

	// query deposits
	deposits := s.queryDeposits(val, proposalID, true, "proposal 3 doesn't exist")
	s.Require().Nil(deposits)
}

func (s *DepositTestSuite) queryDeposits(val *network.Validator, proposalID string, exceptErr bool, message string) *v1.QueryDepositsResponse {
	args := []string{proposalID, fmt.Sprintf("--%s=json", flags.FlagOutput)}
	var depositsRes *v1.QueryDepositsResponse
	cmd := cli.GetCmdQueryDeposits()

	var (
		out testutil.BufferWriter
		err error
	)

	err = s.network.RetryForBlocks(func() error {
		out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
		if err == nil {
			err = val.ClientCtx.LegacyAmino.UnmarshalJSON(out.Bytes(), &depositsRes)
			return err
		}
		return err
	}, 3)

	if exceptErr {
		s.Require().Error(err)
		s.Require().Contains(err.Error(), message)
		return nil
	}

	s.Require().NoError(err)
	return depositsRes
}

func (s *DepositTestSuite) queryDeposit(val *network.Validator, proposalID string, exceptErr bool, message string) *v1.Deposit {
	args := []string{proposalID, val.Address.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)}
	var depositRes *v1.Deposit
	cmd := cli.GetCmdQueryDeposit()
	var (
		out testutil.BufferWriter
		err error
	)

	err = s.network.RetryForBlocks(func() error {
		out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
		return err
	}, 3)

	if exceptErr {
		s.Require().Error(err)
		s.Require().Contains(err.Error(), message)
		return nil
	}
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(out.Bytes(), &depositRes))

	return depositRes
}
