package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	"github.com/stretchr/testify/suite"
)

type TypesTestSuite struct {
	suite.Suite

	chain  *ibctesting.TestChain
	signer sdk.AccAddress
}

func (suite *TypesTestSuite) SetupTest() {
	coordinator := ibctesting.NewCoordinator(suite.T(), 1)
	suite.chain = coordinator.GetChain(ibctesting.GetChainID(0))
	privKey := secp256k1.GenPrivKey()
	suite.signer = sdk.AccAddress(privKey.PubKey().Address())
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
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
				tendermintClient := ibctmtypes.NewClientState(suite.chain.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, 10, commitmenttypes.GetSDKSpecs())
				msg, err = types.NewMsgCreateClient("tendermint", tendermintClient, suite.chain.CreateTMClientHeader().ConsensusState(), suite.signer)
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid tendermint client",
			func() {
				msg, err = types.NewMsgCreateClient("tendermint", &ibctmtypes.ClientState{}, suite.chain.CreateTMClientHeader().ConsensusState(), suite.signer)
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
				tendermintClient := ibctmtypes.NewClientState(suite.chain.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, 10, commitmenttypes.GetSDKSpecs())
				msg, err = types.NewMsgCreateClient("tendermint", tendermintClient, suite.chain.CreateTMClientHeader().ConsensusState(), suite.signer)
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
				soloMachine := ibctesting.NewSolomachine(suite.T(), "solomachine")
				msg, err = types.NewMsgCreateClient(soloMachine.ClientID, soloMachine.ClientState(), soloMachine.ConsensusState(), suite.signer)
				suite.Require().NoError(err)
			},
			true,
		},
		{
			"invalid solomachine client",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), "solomachine")
				msg, err = types.NewMsgCreateClient(soloMachine.ClientID, &solomachinetypes.ClientState{}, soloMachine.ConsensusState(), suite.signer)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"invalid solomachine consensus state",
			func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), "solomachine")
				msg, err = types.NewMsgCreateClient(soloMachine.ClientID, soloMachine.ClientState(), &solomachinetypes.ConsensusState{}, suite.signer)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"unsupported - localhost client",
			func() {
				localhostClient := localhosttypes.NewClientState(suite.chain.ChainID, suite.chain.LastHeader.Header.Height)
				msg, err = types.NewMsgCreateClient("localhost", localhostClient, suite.chain.LastHeader.ConsensusState(), suite.signer)
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
