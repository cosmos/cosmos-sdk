package types_test

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

type TypesTestSuite struct {
	suite.Suite

	chain *ibctesting.TestChain
}

func (suite *TypesTestSuite) SetupTest() {
	coordinator := ibctesting.NewCoordinator(suite.T(), 1)
	suite.chain = coordinator.GetChain(ibctesting.GetChainID(0))
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

// tests that different clients within MsgCreateClient can be marshaled
// and unmarshaled.
func (suite *TypesTestSuite) TestMarshalMsgCreateClient() {
	var (
		msg *types.MsgCreateClient
		err error
	)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"solo machine client", func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chain.Codec, "solomachine", "")
				msg, err = types.NewMsgCreateClient(soloMachine.ClientID, soloMachine.ClientState(), soloMachine.ConsensusState(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
		},
		{
			"tendermint client", func() {
				tendermintClient := ibctmtypes.NewClientState(suite.chain.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				msg, err = types.NewMsgCreateClient("tendermint", tendermintClient, suite.chain.CreateTMClientHeader().ConsensusState(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate()

			cdc := suite.chain.App.AppCodec()

			// marshal message
			bz, err := cdc.MarshalJSON(msg)
			suite.Require().NoError(err)

			// unmarshal message
			newMsg := &types.MsgCreateClient{}
			err = cdc.UnmarshalJSON(bz, newMsg)
			suite.Require().NoError(err)

			suite.Require().True(proto.Equal(msg, newMsg))
		})
	}
}

func (suite *TypesTestSuite) TestMsgCreateClient_ValidateBasic() {
	var (
		msg = &types.MsgCreateClient{}
		err error
	)

	cases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"invalid client-id",
			func() {
				msg.ClientId = ""
			},
			false,
		},
		{
			"valid - tendermint client",
			func() {
				tendermintClient := ibctmtypes.NewClientState(suite.chain.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				msg, err = types.NewMsgCreateClient("tendermint", tendermintClient, suite.chain.CreateTMClientHeader().ConsensusState(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid tendermint client",
			func() {
				msg, err = types.NewMsgCreateClient("tendermint", &ibctmtypes.ClientState{}, suite.chain.CreateTMClientHeader().ConsensusState(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"failed to unpack client",
			func() {
				msg.ClientState = nil
			},
			false,
		},
		{
			"failed to unpack consensus state",
			func() {
				tendermintClient := ibctmtypes.NewClientState(suite.chain.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				msg, err = types.NewMsgCreateClient("tendermint", tendermintClient, suite.chain.CreateTMClientHeader().ConsensusState(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
				msg.ConsensusState = nil
			},
			false,
		},
		{
			"invalid signer",
			func() {
				msg.Signer = nil
			},
			false,
		},
		{
			"valid - solomachine client",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chain.Codec, "solomachine", "")
				msg, err = types.NewMsgCreateClient(soloMachine.ClientID, soloMachine.ClientState(), soloMachine.ConsensusState(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid solomachine client",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chain.Codec, "solomachine", "")
				msg, err = types.NewMsgCreateClient(soloMachine.ClientID, &solomachinetypes.ClientState{}, soloMachine.ConsensusState(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"invalid solomachine consensus state",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chain.Codec, "solomachine", "")
				msg, err = types.NewMsgCreateClient(soloMachine.ClientID, soloMachine.ClientState(), &solomachinetypes.ConsensusState{}, suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"unsupported - localhost client",
			func() {
				localhostClient := localhosttypes.NewClientState(suite.chain.ChainID, types.NewHeight(0, uint64(suite.chain.LastHeader.Header.Height)))
				msg, err = types.NewMsgCreateClient("localhost", localhostClient, suite.chain.LastHeader.ConsensusState(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
	}

	for _, tc := range cases {
		tc.malleate()
		err = msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

// tests that different header within MsgUpdateClient can be marshaled
// and unmarshaled.
func (suite *TypesTestSuite) TestMarshalMsgUpdateClient() {
	var (
		msg *types.MsgUpdateClient
		err error
	)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"solo machine client", func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chain.Codec, "solomachine", "")
				msg, err = types.NewMsgUpdateClient(soloMachine.ClientID, soloMachine.CreateHeader(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
		},
		{
			"tendermint client", func() {
				msg, err = types.NewMsgUpdateClient("tendermint", suite.chain.CreateTMClientHeader(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)

			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate()

			cdc := suite.chain.App.AppCodec()

			// marshal message
			bz, err := cdc.MarshalJSON(msg)
			suite.Require().NoError(err)

			// unmarshal message
			newMsg := &types.MsgUpdateClient{}
			err = cdc.UnmarshalJSON(bz, newMsg)
			suite.Require().NoError(err)

			suite.Require().True(proto.Equal(msg, newMsg))
		})
	}
}

func (suite *TypesTestSuite) TestMsgUpdateClient_ValidateBasic() {
	var (
		msg = &types.MsgUpdateClient{}
		err error
	)

	cases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"invalid client-id",
			func() {
				msg.ClientId = ""
			},
			false,
		},
		{
			"valid - tendermint header",
			func() {
				msg, err = types.NewMsgUpdateClient("tendermint", suite.chain.CreateTMClientHeader(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid tendermint header",
			func() {
				msg, err = types.NewMsgUpdateClient("tendermint", &ibctmtypes.Header{}, suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"failed to unpack header",
			func() {
				msg.Header = nil
			},
			false,
		},
		{
			"invalid signer",
			func() {
				msg.Signer = nil
			},
			false,
		},
		{
			"valid - solomachine header",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chain.Codec, "solomachine", "")
				msg, err = types.NewMsgUpdateClient(soloMachine.ClientID, soloMachine.CreateHeader(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid solomachine header",
			func() {
				msg, err = types.NewMsgUpdateClient("solomachine", &solomachinetypes.Header{}, suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"unsupported - localhost",
			func() {
				msg, err = types.NewMsgUpdateClient(exported.ClientTypeLocalHost, suite.chain.CreateTMClientHeader(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
	}

	for _, tc := range cases {
		tc.malleate()
		err = msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

// tests that different misbehaviours within MsgSubmitMisbehaviour can be marshaled
// and unmarshaled.
func (suite *TypesTestSuite) TestMarshalMsgSubmitMisbehaviour() {
	var (
		msg *types.MsgSubmitMisbehaviour
		err error
	)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"solo machine client", func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chain.Codec, "solomachine", "")
				msg, err = types.NewMsgSubmitMisbehaviour(soloMachine.ClientID, soloMachine.CreateMisbehaviour(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
		},
		{
			"tendermint client", func() {
				header1 := ibctmtypes.CreateTestHeader(suite.chain.ChainID, suite.chain.CurrentHeader.Height, suite.chain.CurrentHeader.Height-1, suite.chain.CurrentHeader.Time, suite.chain.Vals, suite.chain.Vals, suite.chain.Signers)
				header2 := ibctmtypes.CreateTestHeader(suite.chain.ChainID, suite.chain.CurrentHeader.Height, suite.chain.CurrentHeader.Height-1, suite.chain.CurrentHeader.Time.Add(time.Minute), suite.chain.Vals, suite.chain.Vals, suite.chain.Signers)

				misbehaviour := ibctmtypes.NewMisbehaviour("tendermint", suite.chain.ChainID, header1, header2)
				msg, err = types.NewMsgSubmitMisbehaviour("tendermint", misbehaviour, suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)

			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate()

			cdc := suite.chain.App.AppCodec()

			// marshal message
			bz, err := cdc.MarshalJSON(msg)
			suite.Require().NoError(err)

			// unmarshal message
			newMsg := &types.MsgSubmitMisbehaviour{}
			err = cdc.UnmarshalJSON(bz, newMsg)
			suite.Require().NoError(err)

			suite.Require().True(proto.Equal(msg, newMsg))
		})
	}
}

func (suite *TypesTestSuite) TestMsgSubmitMisbehaviour_ValidateBasic() {
	var (
		msg = &types.MsgSubmitMisbehaviour{}
		err error
	)

	cases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"invalid client-id",
			func() {
				msg.ClientId = ""
			},
			false,
		},
		{
			"valid - tendermint misbehaviour",
			func() {
				header1 := ibctmtypes.CreateTestHeader(suite.chain.ChainID, suite.chain.CurrentHeader.Height, suite.chain.CurrentHeader.Height-1, suite.chain.CurrentHeader.Time, suite.chain.Vals, suite.chain.Vals, suite.chain.Signers)
				header2 := ibctmtypes.CreateTestHeader(suite.chain.ChainID, suite.chain.CurrentHeader.Height, suite.chain.CurrentHeader.Height-1, suite.chain.CurrentHeader.Time.Add(time.Minute), suite.chain.Vals, suite.chain.Vals, suite.chain.Signers)

				misbehaviour := ibctmtypes.NewMisbehaviour("tendermint", suite.chain.ChainID, header1, header2)
				msg, err = types.NewMsgSubmitMisbehaviour("tendermint", misbehaviour, suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid tendermint misbehaviour",
			func() {
				msg, err = types.NewMsgSubmitMisbehaviour("tendermint", &ibctmtypes.Misbehaviour{}, suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"failed to unpack misbehaviourt",
			func() {
				msg.Misbehaviour = nil
			},
			false,
		},
		{
			"invalid signer",
			func() {
				msg.Signer = nil
			},
			false,
		},
		{
			"valid - solomachine misbehaviour",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chain.Codec, "solomachine", "")
				msg, err = types.NewMsgSubmitMisbehaviour(soloMachine.ClientID, soloMachine.CreateMisbehaviour(), suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid solomachine misbehaviour",
			func() {
				msg, err = types.NewMsgSubmitMisbehaviour("solomachine", &solomachinetypes.Misbehaviour{}, suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"client-id mismatch",
			func() {
				soloMachineMisbehaviour := ibctesting.NewSolomachine(suite.T(), suite.chain.Codec, "solomachine", "").CreateMisbehaviour()
				msg, err = types.NewMsgSubmitMisbehaviour("external", soloMachineMisbehaviour, suite.chain.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
	}

	for _, tc := range cases {
		tc.malleate()
		err = msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
