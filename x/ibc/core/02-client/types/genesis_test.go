package types_test

import (
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	client "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/09-localhost/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	ibctestingmock "github.com/cosmos/cosmos-sdk/x/ibc/testing/mock"
)

const (
	chainID  = "chainID"
	clientID = "ethbridge"

	height = 10
)

var clientHeight = types.NewHeight(0, 10)

func (suite *TypesTestSuite) TestMarshalGenesisState() {
	cdc := suite.chainA.App.AppCodec()
	clientA, _, _, _, _, _ := suite.coordinator.Setup(suite.chainA, suite.chainB, channeltypes.ORDERED)
	suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, exported.Tendermint)

	genesis := client.ExportGenesis(suite.chainA.GetContext(), suite.chainA.App.IBCKeeper.ClientKeeper)

	bz, err := cdc.MarshalJSON(&genesis)
	suite.Require().NoError(err)
	suite.Require().NotNil(bz)

	var gs types.GenesisState
	err = cdc.UnmarshalJSON(bz, &gs)
	suite.Require().NoError(err)
}

func (suite *TypesTestSuite) TestValidateGenesis() {
	privVal := ibctestingmock.NewPV()
	pubKey, err := privVal.GetPubKey()
	suite.Require().NoError(err)

	now := time.Now().UTC()

	val := tmtypes.NewValidator(pubKey, 10)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{val})

	heightMinus1 := types.NewHeight(0, height-1)
	header := suite.chainA.CreateTMClientHeader(chainID, int64(clientHeight.RevisionHeight), heightMinus1, now, valSet, valSet, []tmtypes.PrivValidator{privVal})

	testCases := []struct {
		name     string
		genState types.GenesisState
		expPass  bool
	}{
		{
			name:     "default",
			genState: types.DefaultGenesisState(),
			expPass:  true,
		},
		{
			name: "valid custom genesis",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false),
					),
					types.NewIdentifiedClientState(
						exported.Localhost, localhosttypes.NewClientState("chainID", clientHeight),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						clientID,
						[]types.ConsensusStateWithHeight{
							types.NewConsensusStateWithHeight(
								header.GetHeight().(types.Height),
								ibctmtypes.NewConsensusState(
									header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), header.Header.NextValidatorsHash,
								),
							),
						},
					),
				},
				types.NewParams(exported.Tendermint, exported.Localhost),
				false,
			),
			expPass: true,
		},
		{
			name: "invalid clientid",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						"/~@$*", ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false),
					),
					types.NewIdentifiedClientState(
						exported.Localhost, localhosttypes.NewClientState("chainID", clientHeight),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						"/~@$*",
						[]types.ConsensusStateWithHeight{
							types.NewConsensusStateWithHeight(
								header.GetHeight().(types.Height),
								ibctmtypes.NewConsensusState(
									header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), header.Header.NextValidatorsHash,
								),
							),
						},
					),
				},
				types.NewParams(exported.Tendermint),
				false,
			),
			expPass: false,
		},
		{
			name: "invalid client",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false),
					),
					types.NewIdentifiedClientState(exported.Localhost, localhosttypes.NewClientState("chaindID", types.ZeroHeight())),
				},
				nil,
				types.NewParams(exported.Tendermint),
				false,
			),
			expPass: false,
		},
		{
			name: "consensus state client id does not match client id in genesis clients",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false),
					),
					types.NewIdentifiedClientState(
						exported.Localhost, localhosttypes.NewClientState("chaindID", clientHeight),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						"wrongclientid",
						[]types.ConsensusStateWithHeight{
							types.NewConsensusStateWithHeight(
								types.ZeroHeight(),
								ibctmtypes.NewConsensusState(
									header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), header.Header.NextValidatorsHash,
								),
							),
						},
					),
				},
				types.NewParams(exported.Tendermint),
				false,
			),
			expPass: false,
		},
		{
			name: "invalid consensus state height",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false),
					),
					types.NewIdentifiedClientState(
						exported.Localhost, localhosttypes.NewClientState("chaindID", clientHeight),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						clientID,
						[]types.ConsensusStateWithHeight{
							types.NewConsensusStateWithHeight(
								types.ZeroHeight(),
								ibctmtypes.NewConsensusState(
									header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), header.Header.NextValidatorsHash,
								),
							),
						},
					),
				},
				types.NewParams(exported.Tendermint),
				false,
			),
			expPass: false,
		},
		{
			name: "invalid consensus state",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false),
					),
					types.NewIdentifiedClientState(
						exported.Localhost, localhosttypes.NewClientState("chaindID", clientHeight),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						clientID,
						[]types.ConsensusStateWithHeight{
							types.NewConsensusStateWithHeight(
								types.NewHeight(0, 1),
								ibctmtypes.NewConsensusState(
									time.Time{}, commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), header.Header.NextValidatorsHash,
								),
							),
						},
					),
				},
				types.NewParams(exported.Tendermint),
				false,
			),
			expPass: false,
		},
		{
			name: "client in genesis clients is disallowed by params",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false),
					),
					types.NewIdentifiedClientState(
						exported.Localhost, localhosttypes.NewClientState("chainID", clientHeight),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						clientID,
						[]types.ConsensusStateWithHeight{
							types.NewConsensusStateWithHeight(
								header.GetHeight().(types.Height),
								ibctmtypes.NewConsensusState(
									header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), header.Header.NextValidatorsHash,
								),
							),
						},
					),
				},
				types.NewParams(exported.Solomachine),
				false,
			),
			expPass: false,
		},
		{
			name: "invalid params",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false),
					),
					types.NewIdentifiedClientState(
						exported.Localhost, localhosttypes.NewClientState("chainID", clientHeight),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						clientID,
						[]types.ConsensusStateWithHeight{
							types.NewConsensusStateWithHeight(
								header.GetHeight().(types.Height),
								ibctmtypes.NewConsensusState(
									header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), header.Header.NextValidatorsHash,
								),
							),
						},
					),
				},
				types.NewParams(" "),
				false,
			),
			expPass: false,
		},
		{
			name: "invalid param",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false),
					),
					types.NewIdentifiedClientState(
						exported.Localhost, localhosttypes.NewClientState("chainID", clientHeight),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						clientID,
						[]types.ConsensusStateWithHeight{
							types.NewConsensusStateWithHeight(
								header.GetHeight().(types.Height),
								ibctmtypes.NewConsensusState(
									header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), header.Header.NextValidatorsHash,
								),
							),
						},
					),
				},
				types.NewParams(" "),
				true,
			),
			expPass: false,
		},
		{
			name: "localhost client not registered on allowlist",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false),
					),
					types.NewIdentifiedClientState(
						exported.Localhost, localhosttypes.NewClientState("chainID", clientHeight),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						clientID,
						[]types.ConsensusStateWithHeight{
							types.NewConsensusStateWithHeight(
								header.GetHeight().(types.Height),
								ibctmtypes.NewConsensusState(
									header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), header.Header.NextValidatorsHash,
								),
							),
						},
					),
				},
				types.NewParams(exported.Tendermint),
				true,
			),
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.genState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
