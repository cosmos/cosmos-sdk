package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
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

func TestValidateGenesis(t *testing.T) {
	privVal := ibctestingmock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	now := time.Now().UTC()

	val := tmtypes.NewValidator(pubKey, 10)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{val})

	header := ibctmtypes.CreateTestHeader(chainID, height, height-1, now, valSet, valSet, []tmtypes.PrivValidator{privVal})

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
			name: "valid genesis",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), false, false),
					),
					types.NewIdentifiedClientState(
						exported.ClientTypeLocalHost, localhosttypes.NewClientState("chainID", clientHeight),
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
				true,
			),
			expPass: true,
		},
		{
			name: "invalid clientid",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						"/~@$*", ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), false, false),
					),
					types.NewIdentifiedClientState(
						exported.ClientTypeLocalHost, localhosttypes.NewClientState("chainID", clientHeight),
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
				true,
			),
			expPass: false,
		},
		{
			name: "invalid client",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), false, false),
					),
					types.NewIdentifiedClientState(exported.ClientTypeLocalHost, localhosttypes.NewClientState("chaindID", types.ZeroHeight())),
				},
				nil,
				true,
			),
			expPass: false,
		},
		{
			name: "invalid consensus state",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), false, false),
					),
					types.NewIdentifiedClientState(
						exported.ClientTypeLocalHost, localhosttypes.NewClientState("chaindID", clientHeight),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						"(CLIENTID2)",
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
				true,
			),
			expPass: false,
		},
		{
			name: "invalid consensus state",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), false, false),
					),
					types.NewIdentifiedClientState(
						exported.ClientTypeLocalHost, localhosttypes.NewClientState("chaindID", clientHeight),
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
				true,
			),
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.genState.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}
