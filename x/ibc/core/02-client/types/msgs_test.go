package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

type TypesTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *TypesTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

func (suite *TypesTestSuite) TestMsgCreateClientGetSignBytes() {
	soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
	msg, err := types.NewMsgCreateClient(soloMachine.ClientState(), soloMachine.ConsensusState(), suite.chainA.SenderAccount.GetAddress())
	suite.Require().NoError(err)
	res := msg.GetSignBytes()

	expected := `{"type":"cosmos-sdk/MsgCreateClient","value":{"client_state":{"consensus_state":{"public_key":{"public_keys":["AibT/cJYde0lnMNJ/hb90Gsg9e5sXqB+aGOfC69Cl4VC","ArRUuKZ4p6+aIgKEuEsYSx4QHOqpNE9vstk2KzZH6Ig5"],"threshold":2},"timestamp":"10"},"sequence":"1"},"consensus_state":{"public_key":{"public_keys":["AibT/cJYde0lnMNJ/hb90Gsg9e5sXqB+aGOfC69Cl4VC","ArRUuKZ4p6+aIgKEuEsYSx4QHOqpNE9vstk2KzZH6Ig5"],"threshold":2},"timestamp":"10"},"signer":"cosmos1eulvtawa5ynj6qlxcv9ys7myahlyshhvl9vmlf"}}`
	suite.Require().Equal(expected, string(res))

	tendermintClient := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
	msg, err = types.NewMsgCreateClient(tendermintClient, suite.chainA.CurrentTMClientHeader().ConsensusState(), suite.chainA.SenderAccount.GetAddress())
	suite.Require().NoError(err)
	res = msg.GetSignBytes()

	expected = `{"type":"cosmos-sdk/MsgCreateClient","value":{"client_state":{"chain_id":"testchain0","frozen_height":{},"latest_height":{"revision_height":"10"},"max_clock_drift":"10000000000","proof_specs":[{"inner_spec":{"child_order":[0,1],"child_size":33,"hash":1,"max_prefix_length":12,"min_prefix_length":4},"leaf_spec":{"hash":1,"length":1,"prefix":"AA==","prehash_value":1}},{"inner_spec":{"child_order":[0,1],"child_size":32,"hash":1,"max_prefix_length":1,"min_prefix_length":1},"leaf_spec":{"hash":1,"length":1,"prefix":"AA==","prehash_value":1}}],"trust_level":{"denominator":"3","numerator":"1"},"trusting_period":"1209600000000000","unbonding_period":"1814400000000000","upgrade_path":["upgrade","upgradedIBCState"]},"consensus_state":{"next_validators_hash":"018119586C484CA72377680E19A8E848811DF312246D7A575ED030E2655B7BF7","root":{"hash":"5/1HBckYk1urdEnNfIduqeKbWzizaggikXRZ2wetyJM="},"timestamp":"2020-01-02T00:00:00Z"},"signer":"cosmos1rashze72a4sm6003nvn7u740kxulhyftawv0fk"}}`
	suite.Require().Equal(expected, string(res))
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
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
				msg, err = types.NewMsgCreateClient(soloMachine.ClientState(), soloMachine.ConsensusState(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
		},
		{
			"tendermint client", func() {
				tendermintClient := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				msg, err = types.NewMsgCreateClient(tendermintClient, suite.chainA.CurrentTMClientHeader().ConsensusState(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate()

			cdc := suite.chainA.App.AppCodec()

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
			"valid - tendermint client",
			func() {
				tendermintClient := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				msg, err = types.NewMsgCreateClient(tendermintClient, suite.chainA.CurrentTMClientHeader().ConsensusState(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid tendermint client",
			func() {
				msg, err = types.NewMsgCreateClient(&ibctmtypes.ClientState{}, suite.chainA.CurrentTMClientHeader().ConsensusState(), suite.chainA.SenderAccount.GetAddress())
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
				tendermintClient := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				msg, err = types.NewMsgCreateClient(tendermintClient, suite.chainA.CurrentTMClientHeader().ConsensusState(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
				msg.ConsensusState = nil
			},
			false,
		},
		{
			"invalid signer",
			func() {
				msg.Signer = ""
			},
			false,
		},
		{
			"valid - solomachine client",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
				msg, err = types.NewMsgCreateClient(soloMachine.ClientState(), soloMachine.ConsensusState(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid solomachine client",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
				msg, err = types.NewMsgCreateClient(&solomachinetypes.ClientState{}, soloMachine.ConsensusState(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"invalid solomachine consensus state",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
				msg, err = types.NewMsgCreateClient(soloMachine.ClientState(), &solomachinetypes.ConsensusState{}, suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"invalid - client state and consensus state client types do not match",
			func() {
				tendermintClient := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
				msg, err = types.NewMsgCreateClient(tendermintClient, soloMachine.ConsensusState(), suite.chainA.SenderAccount.GetAddress())
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
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
				msg, err = types.NewMsgUpdateClient(soloMachine.ClientID, soloMachine.CreateHeader(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
		},
		{
			"tendermint client", func() {
				msg, err = types.NewMsgUpdateClient("tendermint", suite.chainA.CurrentTMClientHeader(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)

			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate()

			cdc := suite.chainA.App.AppCodec()

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

func (suite *TypesTestSuite) TestMsgUpdateClientGetSignBytes() {
	soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
	msg, err := types.NewMsgUpdateClient(soloMachine.ClientID, soloMachine.CreateHeader(), suite.chainA.SenderAccount.GetAddress())
	suite.Require().NoError(err)
	res := msg.GetSignBytes()

	expected := fmt.Sprintf(`{"type":"cosmos-sdk/MsgUpdateClient","value":{"client_id":"solomachine","header":{"new_public_key":{"public_keys":["AzFlPey75R/Qj/bgzN2lyRUECssP6sxMRUj5p4npKL/p","A0mbyQaw2T+fvtar1baQkss2/TOCXCAxGK51UCPOu6ga"],"threshold":2},"sequence":"1","signature":"EpMBCgUIAhIBwBJECkISQOu8lGSWlSNiGpQYYpRMvlKG6NP19LRC37DCEXa6f0NNKU9VvA0EMwApAwhxagjnS87tRt8hZhn6UGQcVaNOR20SRApCEkDbgirlEJpPdYi6ZKdMkS36eoflXLmlsKhXYnRy/DGq1mWsgqYPiIjxNF7jw3WLxhbDPtfHBE/E239hEYRCjtjA","timestamp":"10"},"signer":"%s"}}`, suite.chainA.SenderAccount.GetAddress())
	// suite.Require().Equal(expected, string(res))

	msg, err = types.NewMsgUpdateClient("tendermint", suite.chainA.CurrentTMClientHeader(), suite.chainA.SenderAccount.GetAddress())
	suite.Require().NoError(err)
	res = msg.GetSignBytes()

	expected = `{"type":"cosmos-sdk/MsgCreateClient","value":{"client_state":{"chain_id":"testchain0","frozen_height":{},"latest_height":{"revision_height":"10"},"max_clock_drift":"10000000000","proof_specs":[{"inner_spec":{"child_order":[0,1],"child_size":33,"hash":1,"max_prefix_length":12,"min_prefix_length":4},"leaf_spec":{"hash":1,"length":1,"prefix":"AA==","prehash_value":1}},{"inner_spec":{"child_order":[0,1],"child_size":32,"hash":1,"max_prefix_length":1,"min_prefix_length":1},"leaf_spec":{"hash":1,"length":1,"prefix":"AA==","prehash_value":1}}],"trust_level":{"denominator":"3","numerator":"1"},"trusting_period":"1209600000000000","unbonding_period":"1814400000000000","upgrade_path":["upgrade","upgradedIBCState"]},"consensus_state":{"next_validators_hash":"018119586C484CA72377680E19A8E848811DF312246D7A575ED030E2655B7BF7","root":{"hash":"5/1HBckYk1urdEnNfIduqeKbWzizaggikXRZ2wetyJM="},"timestamp":"2020-01-02T00:00:00Z"},"signer":"cosmos1rashze72a4sm6003nvn7u740kxulhyftawv0fk"}}`
	suite.Require().Equal(expected, string(res)) // FIXME: this panics
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
				msg, err = types.NewMsgUpdateClient("tendermint", suite.chainA.CurrentTMClientHeader(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid tendermint header",
			func() {
				msg, err = types.NewMsgUpdateClient("tendermint", &ibctmtypes.Header{}, suite.chainA.SenderAccount.GetAddress())
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
				msg.Signer = ""
			},
			false,
		},
		{
			"valid - solomachine header",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
				msg, err = types.NewMsgUpdateClient(soloMachine.ClientID, soloMachine.CreateHeader(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid solomachine header",
			func() {
				msg, err = types.NewMsgUpdateClient("solomachine", &solomachinetypes.Header{}, suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"unsupported - localhost",
			func() {
				msg, err = types.NewMsgUpdateClient(exported.Localhost, suite.chainA.CurrentTMClientHeader(), suite.chainA.SenderAccount.GetAddress())
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

func (suite *TypesTestSuite) TestMarshalMsgUpgradeClient() {
	var (
		msg *types.MsgUpgradeClient
		err error
	)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"client upgrades to new tendermint client",
			func() {
				tendermintClient := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				tendermintConsState := &ibctmtypes.ConsensusState{NextValidatorsHash: []byte("nextValsHash")}
				msg, err = types.NewMsgUpgradeClient("clientid", tendermintClient, tendermintConsState, []byte("proofUpgradeClient"), []byte("proofUpgradeConsState"), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
		},
		{
			"client upgrades to new solomachine client",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 1)
				msg, err = types.NewMsgUpgradeClient("clientid", soloMachine.ClientState(), soloMachine.ConsensusState(), []byte("proofUpgradeClient"), []byte("proofUpgradeConsState"), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate()

			cdc := suite.chainA.App.AppCodec()

			// marshal message
			bz, err := cdc.MarshalJSON(msg)
			suite.Require().NoError(err)

			// unmarshal message
			newMsg := &types.MsgUpgradeClient{}
			err = cdc.UnmarshalJSON(bz, newMsg)
			suite.Require().NoError(err)
		})
	}
}

func (suite *TypesTestSuite) TestMsgUpgradeClientGetSignBytes() {
	soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 1)
	msg, err := types.NewMsgUpgradeClient("clientid", soloMachine.ClientState(), soloMachine.ConsensusState(), []byte("proofUpgradeClient"), []byte("proofUpgradeConsState"), suite.chainA.SenderAccount.GetAddress())
	suite.Require().NoError(err)
	res := msg.GetSignBytes()

	expected := fmt.Sprintf(`{"type":"cosmos-sdk/MsgUpgradeClient","value":{"client_id":"clientid","client_state":{"consensus_state":{"public_key":"AwrpSEaM4tX2+JlSU5F1KJeEUhuZvD1yAnW8P64HmTFV","timestamp":"10"},"sequence":"1"},"consensus_state":{"public_key":"AwrpSEaM4tX2+JlSU5F1KJeEUhuZvD1yAnW8P64HmTFV","timestamp":"10"},"proof_upgrade_client":"cHJvb2ZVcGdyYWRlQ2xpZW50","proof_upgrade_consensus_state":"cHJvb2ZVcGdyYWRlQ29uc1N0YXRl","signer":"%s"}}`, suite.chainA.SenderAccount.GetAddress())
	suite.Require().Equal(expected, string(res))

	tendermintClient := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
	tendermintConsState := &ibctmtypes.ConsensusState{NextValidatorsHash: []byte("nextValsHash")}
	msg, err = types.NewMsgUpgradeClient("clientid", tendermintClient, tendermintConsState, []byte("proofUpgradeClient"), []byte("proofUpgradeConsState"), suite.chainA.SenderAccount.GetAddress())
	suite.Require().NoError(err)
	res = msg.GetSignBytes()

	expected = fmt.Sprintf(`{"type":"cosmos-sdk/MsgUpgradeClient","value":{"client_id":"clientid","client_state":{"chain_id":"testchain0","frozen_height":{},"latest_height":{"revision_height":"10"},"max_clock_drift":"10000000000","proof_specs":[{"inner_spec":{"child_order":[0,1],"child_size":33,"hash":1,"max_prefix_length":12,"min_prefix_length":4},"leaf_spec":{"hash":1,"length":1,"prefix":"AA==","prehash_value":1}},{"inner_spec":{"child_order":[0,1],"child_size":32,"hash":1,"max_prefix_length":1,"min_prefix_length":1},"leaf_spec":{"hash":1,"length":1,"prefix":"AA==","prehash_value":1}}],"trust_level":{"denominator":"3","numerator":"1"},"trusting_period":"1209600000000000","unbonding_period":"1814400000000000","upgrade_path":["upgrade","upgradedIBCState"]},"consensus_state":{"next_validators_hash":"6E65787456616C7348617368","root":{},"timestamp":"0001-01-01T00:00:00Z"},"proof_upgrade_client":"cHJvb2ZVcGdyYWRlQ2xpZW50","proof_upgrade_consensus_state":"cHJvb2ZVcGdyYWRlQ29uc1N0YXRl","signer":"%s"}}`, suite.chainA.SenderAccount.GetAddress())
	suite.Require().Equal(expected, string(res))
}

func (suite *TypesTestSuite) TestMsgUpgradeClient_ValidateBasic() {
	cases := []struct {
		name     string
		malleate func(*types.MsgUpgradeClient)
		expPass  bool
	}{
		{
			name:     "success",
			malleate: func(msg *types.MsgUpgradeClient) {},
			expPass:  true,
		},
		{
			name: "client id empty",
			malleate: func(msg *types.MsgUpgradeClient) {
				msg.ClientId = ""
			},
			expPass: false,
		},
		{
			name: "invalid client id",
			malleate: func(msg *types.MsgUpgradeClient) {
				msg.ClientId = "invalid~chain/id"
			},
			expPass: false,
		},
		{
			name: "unpacking clientstate fails",
			malleate: func(msg *types.MsgUpgradeClient) {
				msg.ClientState = nil
			},
			expPass: false,
		},
		{
			name: "unpacking consensus state fails",
			malleate: func(msg *types.MsgUpgradeClient) {
				msg.ConsensusState = nil
			},
			expPass: false,
		},
		{
			name: "client and consensus type does not match",
			malleate: func(msg *types.MsgUpgradeClient) {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
				soloConsensus, err := types.PackConsensusState(soloMachine.ConsensusState())
				suite.Require().NoError(err)
				msg.ConsensusState = soloConsensus
			},
			expPass: false,
		},
		{
			name: "empty client proof",
			malleate: func(msg *types.MsgUpgradeClient) {
				msg.ProofUpgradeClient = nil
			},
			expPass: false,
		},
		{
			name: "empty consensus state proof",
			malleate: func(msg *types.MsgUpgradeClient) {
				msg.ProofUpgradeConsensusState = nil
			},
			expPass: false,
		},
		{
			name: "empty signer",
			malleate: func(msg *types.MsgUpgradeClient) {
				msg.Signer = "  "
			},
			expPass: false,
		},
	}

	for _, tc := range cases {
		tc := tc

		clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
		consState := &ibctmtypes.ConsensusState{NextValidatorsHash: []byte("nextValsHash")}
		msg, err := types.NewMsgUpgradeClient("testclientid", clientState, consState, []byte("proofUpgradeClient"), []byte("proofUpgradeConsState"), suite.chainA.SenderAccount.GetAddress())
		suite.Require().NoError(err)

		tc.malleate(msg)
		err = msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "valid case %s failed", tc.name)
		} else {
			suite.Require().Error(err, "invalid case %s passed", tc.name)
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
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
				msg, err = types.NewMsgSubmitMisbehaviour(soloMachine.ClientID, soloMachine.CreateMisbehaviour(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
		},
		{
			"tendermint client", func() {
				height := types.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height))
				heightMinus1 := types.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height)-1)
				header1 := suite.chainA.CreateTMClientHeader(suite.chainA.ChainID, int64(height.RevisionHeight), heightMinus1, suite.chainA.CurrentHeader.Time, suite.chainA.Vals, suite.chainA.Vals, suite.chainA.Signers)
				header2 := suite.chainA.CreateTMClientHeader(suite.chainA.ChainID, int64(height.RevisionHeight), heightMinus1, suite.chainA.CurrentHeader.Time.Add(time.Minute), suite.chainA.Vals, suite.chainA.Vals, suite.chainA.Signers)

				misbehaviour := ibctmtypes.NewMisbehaviour("tendermint", header1, header2)
				msg, err = types.NewMsgSubmitMisbehaviour("tendermint", misbehaviour, suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)

			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate()

			cdc := suite.chainA.App.AppCodec()

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

func (suite *TypesTestSuite) TestMsgSubmitMisbehaviourGetSignBytes() {
	soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
	msg, err := types.NewMsgSubmitMisbehaviour(soloMachine.ClientID, soloMachine.CreateMisbehaviour(), suite.chainA.SenderAccount.GetAddress())
	suite.Require().NoError(err)
	res := msg.GetSignBytes()

	expected := fmt.Sprintf(`{"type":"cosmos-sdk/MsgSubmitMisbehaviour","value":{"client_id":"solomachine","misbehaviour":{"client_id":"solomachine","sequence":"1","signature_one":{"data":"CikvaWJjL2NsaWVudHMlMkZjb3VudGVycGFydHklMkZjbGllbnRTdGF0ZRL7AQosL2liYy5saWdodGNsaWVudHMuc29sb21hY2hpbmUudjEuQ2xpZW50U3RhdGUSygEIARrFAQrAAQopL2Nvc21vcy5jcnlwdG8ubXVsdGlzaWcuTGVnYWN5QW1pbm9QdWJLZXkSkgEIAhJGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQPIWp26D7PQ7IftuP64ni0UcdoNxhMbm0NVleSTQzeW+xJGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQNK6n93W1oFBLaqwl/NScKvVcbDDuz/C/PYxOwMLgT/KhgK","data_type":1,"signature":"EpMBCgUIAhIBwBJECkISQNPIQU2h3tY5TfqbuUrx2zbkvj2DJG9NGUdt2UYRKEEWfZExGncI/za3EcFDAIoFG6I2u3Et/GKDvO1zmdMdwOUSRApCEkA87Vm/QeVJWY0DHvU+64uJLJQ76lGv8Xb1O7pCmOg1oEITeqrhjbIhZ5qviupoOUOxKQzdANfOBgxdEXWswsfR","timestamp":"10"},"signature_two":{"data":"CjMvaWJjL2NsaWVudHMlMkZjb3VudGVycGFydHklMkZjb25zZW5zdXNTdGF0ZXMlMkYwLTES+QEKLy9pYmMubGlnaHRjbGllbnRzLnNvbG9tYWNoaW5lLnYxLkNvbnNlbnN1c1N0YXRlEsUBCsABCikvY29zbW9zLmNyeXB0by5tdWx0aXNpZy5MZWdhY3lBbWlub1B1YktleRKSAQgCEkYKHy9jb3Ntb3MuY3J5cHRvLnNlY3AyNTZrMS5QdWJLZXkSIwohA8hanboPs9Dsh+24/rieLRRx2g3GExubQ1WV5JNDN5b7EkYKHy9jb3Ntb3MuY3J5cHRvLnNlY3AyNTZrMS5QdWJLZXkSIwohA0rqf3dbWgUEtqrCX81Jwq9VxsMO7P8L89jE7AwuBP8qGAo=","data_type":2,"signature":"EpMBCgUIAhIBwBJECkISQPuhj97kVBX79lp6a+iuq4MQu1hZiY8eAWfVWYtT0qyka+2+75GWsknhsC2ehdx+cTYhmGD701e4QS6Qcy3BJtwSRApCEkDuzVqNeRtnqKza3v5ciCLeACPSnz8W5UWKDzjbRuLDCXHN5rod8dE7hxBYWD5FT+w/9URs2EAfMdJ4pm4parS4","timestamp":"11"}},"signer":"%s"}}`, suite.chainA.SenderAccount.GetAddress())
	// suite.Require().Equal(expected, string(res))

	height := types.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height))
	heightMinus1 := types.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height)-1)
	header1 := suite.chainA.CreateTMClientHeader(suite.chainA.ChainID, int64(height.RevisionHeight), heightMinus1, suite.chainA.CurrentHeader.Time, suite.chainA.Vals, suite.chainA.Vals, suite.chainA.Signers)
	header2 := suite.chainA.CreateTMClientHeader(suite.chainA.ChainID, int64(height.RevisionHeight), heightMinus1, suite.chainA.CurrentHeader.Time.Add(time.Minute), suite.chainA.Vals, suite.chainA.Vals, suite.chainA.Signers)

	misbehaviour := ibctmtypes.NewMisbehaviour("tendermint", header1, header2)
	msg, err = types.NewMsgSubmitMisbehaviour("tendermint", misbehaviour, suite.chainA.SenderAccount.GetAddress())
	suite.Require().NoError(err)
	res = msg.GetSignBytes()

	expected = fmt.Sprintf(`{"type":"cosmos-sdk/MsgSubmitMisbehaviour","value":{"client_id":"clientid","client_state":{"chain_id":"testchain0","frozen_height":{},"latest_height":{"revision_height":"10"},"max_clock_drift":"10000000000","proof_specs":[{"inner_spec":{"child_order":[0,1],"child_size":33,"hash":1,"max_prefix_length":12,"min_prefix_length":4},"leaf_spec":{"hash":1,"length":1,"prefix":"AA==","prehash_value":1}},{"inner_spec":{"child_order":[0,1],"child_size":32,"hash":1,"max_prefix_length":1,"min_prefix_length":1},"leaf_spec":{"hash":1,"length":1,"prefix":"AA==","prehash_value":1}}],"trust_level":{"denominator":"3","numerator":"1"},"trusting_period":"1209600000000000","unbonding_period":"1814400000000000","upgrade_path":["upgrade","upgradedIBCState"]},"consensus_state":{"next_validators_hash":"6E65787456616C7348617368","root":{},"timestamp":"0001-01-01T00:00:00Z"},"proof_upgrade_client":"cHJvb2ZVcGdyYWRlQ2xpZW50","proof_upgrade_consensus_state":"cHJvb2ZVcGdyYWRlQ29uc1N0YXRl","signer":"%s"}}`, suite.chainA.SenderAccount.GetAddress())
	suite.Require().Equal(expected, string(res))
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
				height := types.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height))
				heightMinus1 := types.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height)-1)
				header1 := suite.chainA.CreateTMClientHeader(suite.chainA.ChainID, int64(height.RevisionHeight), heightMinus1, suite.chainA.CurrentHeader.Time, suite.chainA.Vals, suite.chainA.Vals, suite.chainA.Signers)
				header2 := suite.chainA.CreateTMClientHeader(suite.chainA.ChainID, int64(height.RevisionHeight), heightMinus1, suite.chainA.CurrentHeader.Time.Add(time.Minute), suite.chainA.Vals, suite.chainA.Vals, suite.chainA.Signers)

				misbehaviour := ibctmtypes.NewMisbehaviour("tendermint", header1, header2)
				msg, err = types.NewMsgSubmitMisbehaviour("tendermint", misbehaviour, suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid tendermint misbehaviour",
			func() {
				msg, err = types.NewMsgSubmitMisbehaviour("tendermint", &ibctmtypes.Misbehaviour{}, suite.chainA.SenderAccount.GetAddress())
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
				msg.Signer = ""
			},
			false,
		},
		{
			"valid - solomachine misbehaviour",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2)
				msg, err = types.NewMsgSubmitMisbehaviour(soloMachine.ClientID, soloMachine.CreateMisbehaviour(), suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid solomachine misbehaviour",
			func() {
				msg, err = types.NewMsgSubmitMisbehaviour("solomachine", &solomachinetypes.Misbehaviour{}, suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"client-id mismatch",
			func() {
				soloMachineMisbehaviour := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2).CreateMisbehaviour()
				msg, err = types.NewMsgSubmitMisbehaviour("external", soloMachineMisbehaviour, suite.chainA.SenderAccount.GetAddress())
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
