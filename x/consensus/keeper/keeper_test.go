package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx                   sdk.Context
	consensusParamsKeeper *consensusparamkeeper.Keeper

	queryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(consensusparamkeeper.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{})
	env := runtime.NewEnvironment(runtime.NewKVStoreService(key), log.NewNopLogger())

	keeper := consensusparamkeeper.NewKeeper(encCfg.Codec, env, authtypes.NewModuleAddress("gov").String())

	s.ctx = ctx
	s.consensusParamsKeeper = &keeper

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper)
	s.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestGRPCQueryConsensusParams() {
	defaultConsensusParams := cmttypes.DefaultConsensusParams().ToProto()

	testCases := []struct {
		msg      string
		req      types.QueryParamsRequest
		malleate func()
		response types.QueryParamsResponse
		expPass  bool
	}{
		{
			"success",
			types.QueryParamsRequest{},
			func() {
				input := &types.MsgUpdateParams{
					Authority: s.consensusParamsKeeper.GetAuthority(),
					Block:     defaultConsensusParams.Block,
					Validator: defaultConsensusParams.Validator,
					Evidence:  defaultConsensusParams.Evidence,
					Abci:      defaultConsensusParams.Abci,
				}
				_, err := s.consensusParamsKeeper.UpdateParams(s.ctx, input)
				s.Require().NoError(err)
			},
			types.QueryParamsResponse{
				Params: &cmtproto.ConsensusParams{
					Block:     defaultConsensusParams.Block,
					Validator: defaultConsensusParams.Validator,
					Evidence:  defaultConsensusParams.Evidence,
					Version:   defaultConsensusParams.Version,
					Abci:      defaultConsensusParams.Abci,
				},
			},
			true,
		},
		{
			"success with abci",
			types.QueryParamsRequest{},
			func() {
				input := &types.MsgUpdateParams{
					Authority: s.consensusParamsKeeper.GetAuthority(),
					Block:     defaultConsensusParams.Block,
					Validator: defaultConsensusParams.Validator,
					Evidence:  defaultConsensusParams.Evidence,
					Abci: &cmtproto.ABCIParams{
						VoteExtensionsEnableHeight: 1234,
					},
				}
				_, err := s.consensusParamsKeeper.UpdateParams(s.ctx, input)
				s.Require().NoError(err)
			},
			types.QueryParamsResponse{
				Params: &cmtproto.ConsensusParams{
					Block:     defaultConsensusParams.Block,
					Validator: defaultConsensusParams.Validator,
					Evidence:  defaultConsensusParams.Evidence,
					Version:   defaultConsensusParams.Version,
					Abci: &cmtproto.ABCIParams{
						VoteExtensionsEnableHeight: 1234,
					},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.msg, func() {
			s.SetupTest() // reset

			tc.malleate()
			res, err := s.consensusParamsKeeper.Params(s.ctx, &tc.req)

			if tc.expPass {
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().Equal(tc.response.Params, res.Params)
			} else {
				s.Require().Error(err)
				s.Require().Nil(res)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateParams() {
	defaultConsensusParams := cmttypes.DefaultConsensusParams().ToProto()
	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "valid params",
			input: &types.MsgUpdateParams{
				Authority: s.consensusParamsKeeper.GetAuthority(),
				Block:     defaultConsensusParams.Block,
				Validator: defaultConsensusParams.Validator,
				Evidence:  defaultConsensusParams.Evidence,
				Abci:      defaultConsensusParams.Abci,
			},
			expErr:    false,
			expErrMsg: "",
		},
		{
			name: "invalid  params",
			input: &types.MsgUpdateParams{
				Authority: s.consensusParamsKeeper.GetAuthority(),
				Block:     &cmtproto.BlockParams{MaxGas: -10, MaxBytes: -10},
				Validator: defaultConsensusParams.Validator,
				Evidence:  defaultConsensusParams.Evidence,
				Abci:      defaultConsensusParams.Abci,
			},
			expErr:    true,
			expErrMsg: "block.MaxBytes must be -1 or greater than 0. Got -10",
		},
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Block:     defaultConsensusParams.Block,
				Validator: defaultConsensusParams.Validator,
				Evidence:  defaultConsensusParams.Evidence,
				Abci:      defaultConsensusParams.Abci,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "nil evidence params",
			input: &types.MsgUpdateParams{
				Authority: s.consensusParamsKeeper.GetAuthority(),
				Block:     defaultConsensusParams.Block,
				Validator: defaultConsensusParams.Validator,
				Evidence:  nil,
				Abci:      defaultConsensusParams.Abci,
			},
			expErr:    true,
			expErrMsg: "all parameters must be present",
		},
		{
			name: "nil block params",
			input: &types.MsgUpdateParams{
				Authority: s.consensusParamsKeeper.GetAuthority(),
				Block:     nil,
				Validator: defaultConsensusParams.Validator,
				Evidence:  defaultConsensusParams.Evidence,
				Abci:      defaultConsensusParams.Abci,
			},
			expErr:    true,
			expErrMsg: "all parameters must be present",
		},
		{
			name: "nil validator params",
			input: &types.MsgUpdateParams{
				Authority: s.consensusParamsKeeper.GetAuthority(),
				Block:     defaultConsensusParams.Block,
				Validator: nil,
				Evidence:  defaultConsensusParams.Evidence,
				Abci:      defaultConsensusParams.Abci,
			},
			expErr:    true,
			expErrMsg: "all parameters must be present",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()
			_, err := s.consensusParamsKeeper.UpdateParams(s.ctx, tc.input)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)

				res, err := s.consensusParamsKeeper.Params(s.ctx, &types.QueryParamsRequest{})
				s.Require().NoError(err)

				s.Require().Equal(tc.input.Abci, res.Params.Abci)
				s.Require().Equal(tc.input.Block, res.Params.Block)
				s.Require().Equal(tc.input.Evidence, res.Params.Evidence)
				s.Require().Equal(tc.input.Validator, res.Params.Validator)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetParams() {
	defaultConsensusParams := cmttypes.DefaultConsensusParams().ToProto()
	testCases := []struct {
		name      string
		input     *types.ConsensusMsgParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "valid params",
			input: &types.ConsensusMsgParams{
				Abci:      defaultConsensusParams.Abci,
				Version:   &cmtproto.VersionParams{App: 1},
				Block:     defaultConsensusParams.Block,
				Validator: defaultConsensusParams.Validator,
				Evidence:  defaultConsensusParams.Evidence,
			},
			expErr:    false,
			expErrMsg: "",
		},
		{
			name: "invalid  params",
			input: &types.ConsensusMsgParams{
				Abci:      defaultConsensusParams.Abci,
				Version:   &cmtproto.VersionParams{App: 1},
				Block:     &cmtproto.BlockParams{MaxGas: -10, MaxBytes: -10},
				Validator: defaultConsensusParams.Validator,
				Evidence:  defaultConsensusParams.Evidence,
			},
			expErr:    true,
			expErrMsg: "block.MaxBytes must be -1 or greater than 0. Got -10",
		},
		{
			name: "nil version params",
			input: &types.ConsensusMsgParams{
				Abci:      defaultConsensusParams.Abci,
				Version:   nil,
				Block:     defaultConsensusParams.Block,
				Validator: defaultConsensusParams.Validator,
				Evidence:  defaultConsensusParams.Evidence,
			},
			expErr:    true,
			expErrMsg: "all parameters must be present",
		},
		{
			name: "nil evidence params",
			input: &types.ConsensusMsgParams{
				Abci:      defaultConsensusParams.Abci,
				Version:   &cmtproto.VersionParams{App: 1},
				Block:     defaultConsensusParams.Block,
				Validator: defaultConsensusParams.Validator,
				Evidence:  nil,
			},
			expErr:    true,
			expErrMsg: "all parameters must be present",
		},
		{
			name: "nil block params",
			input: &types.ConsensusMsgParams{
				Abci:      defaultConsensusParams.Abci,
				Version:   &cmtproto.VersionParams{App: 1},
				Block:     nil,
				Validator: defaultConsensusParams.Validator,
				Evidence:  defaultConsensusParams.Evidence,
			},
			expErr:    true,
			expErrMsg: "all parameters must be present",
		},
		{
			name: "nil validator params",
			input: &types.ConsensusMsgParams{
				Abci:      defaultConsensusParams.Abci,
				Version:   &cmtproto.VersionParams{App: 1},
				Block:     defaultConsensusParams.Block,
				Validator: nil,
				Evidence:  defaultConsensusParams.Evidence,
			},
			expErr:    true,
			expErrMsg: "all parameters must be present",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()
			_, err := s.consensusParamsKeeper.SetParams(s.ctx, tc.input)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)

				res, err := s.consensusParamsKeeper.Params(s.ctx, &types.QueryParamsRequest{})
				s.Require().NoError(err)

				s.Require().Equal(tc.input.Abci, res.Params.Abci)
				s.Require().Equal(tc.input.Block, res.Params.Block)
				s.Require().Equal(tc.input.Evidence, res.Params.Evidence)
				s.Require().Equal(tc.input.Validator, res.Params.Validator)
				s.Require().Equal(tc.input.Version, res.Params.Version)
			}
		})
	}
}
