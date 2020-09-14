package cli_test

import (
	json "encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var lock sync.RWMutex

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	lock.Lock()

	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 2

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	unbond, err := sdk.ParseCoin("10stake")
	s.Require().NoError(err)

	val := s.network.Validators()[0]
	val2 := s.network.Validators()[1]

	// redelegate
	_, err = stakingtestutil.MsgRedelegateExec(val.ClientCtx, val.Address, val.ValAddress, val2.ValAddress, unbond)
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	// unbonding
	_, err = stakingtestutil.MsgUnbondExec(val.ClientCtx, val.Address, val.ValAddress, unbond)
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	defer lock.Unlock()

	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestNewCreateValidatorCmd() {
	val := s.network.Validators()[0]

	consPrivKey := ed25519.GenPrivKey()
	consPubKey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, consPrivKey.PubKey())
	s.Require().NoError(err)

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewValidator", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"invalid transaction (missing amount)",
			[]string{
				fmt.Sprintf("--%s=AFAF00C4", cli.FlagIdentity),
				fmt.Sprintf("--%s=https://newvalidator.io", cli.FlagWebsite),
				fmt.Sprintf("--%s=contact@newvalidator.io", cli.FlagSecurityContact),
				fmt.Sprintf("--%s='Hey, I am a new validator. Please delegate!'", cli.FlagDetails),
				fmt.Sprintf("--%s=0.5", cli.FlagCommissionRate),
				fmt.Sprintf("--%s=1.0", cli.FlagCommissionMaxRate),
				fmt.Sprintf("--%s=0.1", cli.FlagCommissionMaxChangeRate),
				fmt.Sprintf("--%s=1", cli.FlagMinSelfDelegation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"invalid transaction (missing pubkey)",
			[]string{
				fmt.Sprintf("--%s=100stake", cli.FlagAmount),
				fmt.Sprintf("--%s=AFAF00C4", cli.FlagIdentity),
				fmt.Sprintf("--%s=https://newvalidator.io", cli.FlagWebsite),
				fmt.Sprintf("--%s=contact@newvalidator.io", cli.FlagSecurityContact),
				fmt.Sprintf("--%s='Hey, I am a new validator. Please delegate!'", cli.FlagDetails),
				fmt.Sprintf("--%s=0.5", cli.FlagCommissionRate),
				fmt.Sprintf("--%s=1.0", cli.FlagCommissionMaxRate),
				fmt.Sprintf("--%s=0.1", cli.FlagCommissionMaxChangeRate),
				fmt.Sprintf("--%s=1", cli.FlagMinSelfDelegation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"invalid transaction (missing moniker)",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPubKey, consPubKey),
				fmt.Sprintf("--%s=100stake", cli.FlagAmount),
				fmt.Sprintf("--%s=AFAF00C4", cli.FlagIdentity),
				fmt.Sprintf("--%s=https://newvalidator.io", cli.FlagWebsite),
				fmt.Sprintf("--%s=contact@newvalidator.io", cli.FlagSecurityContact),
				fmt.Sprintf("--%s='Hey, I am a new validator. Please delegate!'", cli.FlagDetails),
				fmt.Sprintf("--%s=0.5", cli.FlagCommissionRate),
				fmt.Sprintf("--%s=1.0", cli.FlagCommissionMaxRate),
				fmt.Sprintf("--%s=0.1", cli.FlagCommissionMaxChangeRate),
				fmt.Sprintf("--%s=1", cli.FlagMinSelfDelegation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPubKey, consPubKey),
				fmt.Sprintf("--%s=100stake", cli.FlagAmount),
				fmt.Sprintf("--%s=NewValidator", cli.FlagMoniker),
				fmt.Sprintf("--%s=AFAF00C4", cli.FlagIdentity),
				fmt.Sprintf("--%s=https://newvalidator.io", cli.FlagWebsite),
				fmt.Sprintf("--%s=contact@newvalidator.io", cli.FlagSecurityContact),
				fmt.Sprintf("--%s='Hey, I am a new validator. Please delegate!'", cli.FlagDetails),
				fmt.Sprintf("--%s=0.5", cli.FlagCommissionRate),
				fmt.Sprintf("--%s=1.0", cli.FlagCommissionMaxRate),
				fmt.Sprintf("--%s=0.1", cli.FlagCommissionMaxChangeRate),
				fmt.Sprintf("--%s=1", cli.FlagMinSelfDelegation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCreateValidatorCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryValidator() {
	val := s.network.Validators()[0]
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"with invalid address ",
			[]string{"somethinginvalidaddress", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
		},
		{
			"with valid and not existing address",
			[]string{"cosmosvaloper15jkng8hytwt22lllv6mw4k89qkqehtahd84ptu", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
		},
		{
			"happy case",
			[]string{fmt.Sprintf("%s", val.ValAddress), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidator()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().NotEqual("internal", err.Error())
			} else {
				var result types.Validator
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &result))
				s.Require().Equal(val.ValAddress, result.OperatorAddress)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryValidators() {
	val := s.network.Validators()[0]

	testCases := []struct {
		name              string
		args              []string
		minValidatorCount int
	}{
		{
			"one validator case",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			1,
		},
		{
			"multi validator case",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			len(s.network.Validators()),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidators()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result []types.Validator
			s.Require().NoError(json.Unmarshal(out.Bytes(), &result))
			s.Require().Equal(len(s.network.Validators()), len(result))
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryDelegation() {
	val := s.network.Validators()[0]
	val2 := s.network.Validators()[1]

	testCases := []struct {
		name     string
		args     []string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"with wrong delegator address",
			[]string{
				"wrongDelAddr",
				val2.ValAddress.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true, nil, nil,
		},
		{
			"with wrong validator address",
			[]string{
				val.Address.String(),
				"wrongValAddr",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true, nil, nil,
		},
		{
			"with json output",
			[]string{
				val.Address.String(),
				val2.ValAddress.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			&types.DelegationResponse{},
			&types.DelegationResponse{
				Delegation: types.Delegation{
					DelegatorAddress: val.Address,
					ValidatorAddress: val2.ValAddress,
					Shares:           sdk.NewDec(10),
				},
				Balance: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10)),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryDelegation()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryDelegations() {
	val := s.network.Validators()[0]

	testCases := []struct {
		name     string
		args     []string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"with no delegator address",
			[]string{},
			true, nil, nil,
		},
		{
			"with wrong delegator address",
			[]string{"wrongDelAddr"},
			true, nil, nil,
		},
		{
			"valid request (height specific)",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			false,
			&types.QueryDelegatorDelegationsResponse{},
			&types.QueryDelegatorDelegationsResponse{
				DelegationResponses: types.DelegationResponses{
					types.NewDelegationResp(val.Address, val.ValAddress, sdk.NewDec(100000000), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000))),
				},
				Pagination: &query.PageResponse{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryDelegations()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryDelegationsTo() {
	val := s.network.Validators()[0]

	testCases := []struct {
		name     string
		args     []string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"with no validator address",
			[]string{},
			true, nil, nil,
		},
		{
			"wrong validator address",
			[]string{"wrongValAddr"},
			true, nil, nil,
		},
		{
			"valid request(height specific)",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			false,
			&types.QueryValidatorDelegationsResponse{},
			&types.QueryValidatorDelegationsResponse{
				DelegationResponses: types.DelegationResponses{
					types.NewDelegationResp(val.Address, val.ValAddress, sdk.NewDec(100000000), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000))),
				},
				Pagination: &query.PageResponse{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryDelegations()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryUnbondingDelegations() {
	val := s.network.Validators()[0]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			[]string{
				"wrongDelAddr",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryUnbondingDelegations()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var ubds types.QueryDelegatorUnbondingDelegationsResponse
				err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &ubds)

				s.Require().NoError(err)
				s.Require().Len(ubds.UnbondingResponses, 1)
				s.Require().Equal(ubds.UnbondingResponses[0].DelegatorAddress, val.Address)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryUnbondingDelegation() {
	val := s.network.Validators()[0]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			[]string{
				"wrongDelAddr",
				val.ValAddress.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"wrong validator address",
			[]string{
				val.Address.String(),
				"wrongValAddr",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.Address.String(),
				val.ValAddress.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryUnbondingDelegation()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var ubd types.UnbondingDelegation

				err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &ubd)
				s.Require().NoError(err)
				s.Require().Equal(ubd.DelegatorAddress, val.Address)
				s.Require().Equal(ubd.ValidatorAddress, val.ValAddress)
				s.Require().Len(ubd.Entries, 1)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryValidatorUnbondingDelegations() {
	val := s.network.Validators()[0]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong validator address",
			[]string{
				"wrongValAddr",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.ValAddress.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidatorUnbondingDelegations()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var ubds types.QueryValidatorUnbondingDelegationsResponse
				err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &ubds)

				s.Require().NoError(err)
				s.Require().Len(ubds.UnbondingResponses, 1)
				s.Require().Equal(ubds.UnbondingResponses[0].DelegatorAddress, val.Address)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryRedelegations() {
	val := s.network.Validators()[0]
	val2 := s.network.Validators()[1]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			[]string{
				"wrongdeladdr",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryRedelegations()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var redelegations types.QueryRedelegationsResponse
				err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &redelegations)

				s.Require().NoError(err)

				s.Require().Len(redelegations.RedelegationResponses, 1)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.DelegatorAddress, val.Address)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorSrcAddress, val.ValAddress)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorDstAddress, val2.ValAddress)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryRedelegation() {
	val := s.network.Validators()[0]
	val2 := s.network.Validators()[1]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			[]string{
				"wrongdeladdr",
				val.ValAddress.String(),
				val2.ValAddress.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"wrong source validator address address",
			[]string{
				val.Address.String(),
				"wrongSrcValAddress",
				val2.ValAddress.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"wrong destination validator address address",
			[]string{
				val.Address.String(),
				val.ValAddress.String(),
				"wrongDestValAddress",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.Address.String(),
				val.ValAddress.String(),
				val2.ValAddress.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryRedelegation()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var redelegations types.QueryRedelegationsResponse

				err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &redelegations)
				s.Require().NoError(err)

				s.Require().Len(redelegations.RedelegationResponses, 1)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.DelegatorAddress, val.Address)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorSrcAddress, val.ValAddress)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorDstAddress, val2.ValAddress)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryRedelegationsFrom() {
	val := s.network.Validators()[0]
	val2 := s.network.Validators()[1]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong validator address",
			[]string{
				"wrongValAddr",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid request",
			[]string{
				val.ValAddress.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidatorRedelegations()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var redelegations types.QueryRedelegationsResponse
				err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &redelegations)

				s.Require().NoError(err)

				s.Require().Len(redelegations.RedelegationResponses, 1)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.DelegatorAddress, val.Address)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorSrcAddress, val.ValAddress)
				s.Require().Equal(redelegations.RedelegationResponses[0].Redelegation.ValidatorDstAddress, val2.ValAddress)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryHistoricalInfo() {
	val := s.network.Validators()[0]

	testCases := []struct {
		name  string
		args  []string
		error bool
	}{
		{
			"wrong height",
			[]string{
				"-1",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid request",
			[]string{
				"1",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryHistoricalInfo()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.error {
				s.Require().Error(err)
			} else {
				var historical_info types.HistoricalInfo

				err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &historical_info)
				s.Require().NoError(err)
				s.Require().NotNil(historical_info)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryParams() {
	val := s.network.Validators()[0]
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"with text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`bond_denom: stake
historical_entries: 100
max_entries: 7
max_validators: 100
unbonding_time: 1814400s`,
		},
		{
			"with json output",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"unbonding_time":"1814400s","max_validators":100,"max_entries":7,"historical_entries":100,"bond_denom":"stake"}`,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryParams()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryPool() {
	val := s.network.Validators()[0]
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"with text",
			[]string{
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			`bonded_tokens: "200000000"
not_bonded_tokens: "0"`,
		},
		{
			"with json",
			[]string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			`{"not_bonded_tokens":"0","bonded_tokens":"200000000"}`,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryPool()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
