package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

var (
	emptyPrefix = commitmenttypes.MerklePrefix{}
	emptyProof  = []byte{}
)

type MsgTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	proof []byte
}

func (suite *MsgTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)

	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	app := simapp.Setup(false)
	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db)
	storeKey := storetypes.NewKVStoreKey("iavlStoreKey")

	store.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, nil)
	store.LoadVersion(0)
	iavlStore := store.GetCommitStore(storeKey).(*iavl.Store)

	iavlStore.Set([]byte("KEY"), []byte("VALUE"))
	_ = store.Commit()

	res := store.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("/%s/key", storeKey.Name()), // required path to get key/value+proof
		Data:  []byte("KEY"),
		Prove: true,
	})

	merkleProof, err := commitmenttypes.ConvertProofs(res.ProofOps)
	suite.Require().NoError(err)
	proof, err := app.AppCodec().MarshalBinaryBare(&merkleProof)
	suite.Require().NoError(err)

	suite.proof = proof

}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}

func (suite *MsgTestSuite) TestNewMsgConnectionOpenInit() {
	prefix := commitmenttypes.NewMerklePrefix([]byte("storePrefixKey"))
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")
	// empty versions are considered valid, the default compatible versions
	// will be used in protocol.
	var version *types.Version

	var testCases = []struct {
		name    string
		msg     *types.MsgConnectionOpenInit
		expPass bool
	}{
		{"invalid client ID", types.NewMsgConnectionOpenInit("test/iris", "clienttotest", prefix, version, 500, signer), false},
		{"invalid counterparty client ID", types.NewMsgConnectionOpenInit("clienttotest", "(clienttotest)", prefix, version, 500, signer), false},
		{"invalid counterparty connection ID", &types.MsgConnectionOpenInit{connectionID, types.NewCounterparty("clienttotest", "connectiontotest", prefix), version, 500, signer.String()}, false},
		{"empty counterparty prefix", types.NewMsgConnectionOpenInit("clienttotest", "clienttotest", emptyPrefix, version, 500, signer), false},
		{"supplied version fails basic validation", types.NewMsgConnectionOpenInit("clienttotest", "clienttotest", prefix, &types.Version{}, 500, signer), false},
		{"empty singer", types.NewMsgConnectionOpenInit("clienttotest", "clienttotest", prefix, version, 500, nil), false},
		{"success", types.NewMsgConnectionOpenInit("clienttotest", "clienttotest", prefix, version, 500, signer), true},
	}

	for _, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *MsgTestSuite) TestNewMsgConnectionOpenTry() {
	prefix := commitmenttypes.NewMerklePrefix([]byte("storePrefixKey"))
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	clientState := ibctmtypes.NewClientState(
		chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false,
	)

	// Pack consensus state into any to test unpacking error
	consState := ibctmtypes.NewConsensusState(
		time.Now(), commitmenttypes.NewMerkleRoot([]byte("root")), []byte("nextValsHash"),
	)
	invalidAny := clienttypes.MustPackConsensusState(consState)
	counterparty := types.NewCounterparty("connectiontotest", "clienttotest", prefix)

	// invalidClientState fails validateBasic
	invalidClient := ibctmtypes.NewClientState(
		chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clienttypes.ZeroHeight(), commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false,
	)

	var testCases = []struct {
		name    string
		msg     *types.MsgConnectionOpenTry
		expPass bool
	}{
		{"invalid connection ID", types.NewMsgConnectionOpenTry("test/conn1", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer), false},
		{"invalid connection ID", types.NewMsgConnectionOpenTry("(invalidconnection)", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer), false},
		{"invalid client ID", types.NewMsgConnectionOpenTry(connectionID, "test/iris", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer), false},
		{"invalid counterparty connection ID", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "ibc/test", "clienttotest", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer), false},
		{"invalid counterparty client ID", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "test/conn1", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer), false},
		{"invalid nil counterparty client", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", nil, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer), false},
		{"invalid client unpacking", &types.MsgConnectionOpenTry{connectionID, "clienttotesta", invalidAny, counterparty, 500, []*types.Version{ibctesting.ConnectionVersion}, clientHeight, suite.proof, suite.proof, suite.proof, clientHeight, signer.String()}, false},
		{"counterparty failed Validate", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", invalidClient, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer), false},
		{"empty counterparty prefix", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", clientState, emptyPrefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer), false},
		{"empty counterpartyVersions", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer), false},
		{"empty proofInit", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, emptyProof, suite.proof, suite.proof, clientHeight, clientHeight, signer), false},
		{"empty proofClient", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, emptyProof, suite.proof, clientHeight, clientHeight, signer), false},
		{"empty proofConsensus", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, emptyProof, clientHeight, clientHeight, signer), false},
		{"invalid proofHeight", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clienttypes.ZeroHeight(), clientHeight, signer), false},
		{"invalid consensusHeight", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clienttypes.ZeroHeight(), signer), false},
		{"empty singer", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, nil), false},
		{"success", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{ibctesting.ConnectionVersion}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer), true},
		{"invalid version", types.NewMsgConnectionOpenTry(connectionID, "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []*types.Version{{}}, 500, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer), false},
	}

	for _, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *MsgTestSuite) TestNewMsgConnectionOpenAck() {
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")
	clientState := ibctmtypes.NewClientState(
		chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false,
	)

	// Pack consensus state into any to test unpacking error
	consState := ibctmtypes.NewConsensusState(
		time.Now(), commitmenttypes.NewMerkleRoot([]byte("root")), []byte("nextValsHash"),
	)
	invalidAny := clienttypes.MustPackConsensusState(consState)

	// invalidClientState fails validateBasic
	invalidClient := ibctmtypes.NewClientState(
		chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clienttypes.ZeroHeight(), commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false,
	)
	connectionID := "connection-0"

	var testCases = []struct {
		name    string
		msg     *types.MsgConnectionOpenAck
		expPass bool
	}{
		{"invalid connection ID", types.NewMsgConnectionOpenAck("test/conn1", connectionID, clientState, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer), false},
		{"invalid counterparty connection ID", types.NewMsgConnectionOpenAck(connectionID, "test/conn1", clientState, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer), false},
		{"invalid nil counterparty client", types.NewMsgConnectionOpenAck(connectionID, connectionID, nil, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer), false},
		{"invalid unpacking counterparty client", &types.MsgConnectionOpenAck{connectionID, connectionID, ibctesting.ConnectionVersion, invalidAny, clientHeight, suite.proof, suite.proof, suite.proof, clientHeight, signer.String()}, false},
		{"counterparty client failed Validate", types.NewMsgConnectionOpenAck(connectionID, connectionID, invalidClient, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer), false},
		{"empty proofTry", types.NewMsgConnectionOpenAck(connectionID, connectionID, clientState, emptyProof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer), false},
		{"empty proofClient", types.NewMsgConnectionOpenAck(connectionID, connectionID, clientState, suite.proof, emptyProof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer), false},
		{"empty proofConsensus", types.NewMsgConnectionOpenAck(connectionID, connectionID, clientState, suite.proof, suite.proof, emptyProof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer), false},
		{"invalid proofHeight", types.NewMsgConnectionOpenAck(connectionID, connectionID, clientState, suite.proof, suite.proof, suite.proof, clienttypes.ZeroHeight(), clientHeight, ibctesting.ConnectionVersion, signer), false},
		{"invalid consensusHeight", types.NewMsgConnectionOpenAck(connectionID, connectionID, clientState, suite.proof, suite.proof, suite.proof, clientHeight, clienttypes.ZeroHeight(), ibctesting.ConnectionVersion, signer), false},
		{"invalid version", types.NewMsgConnectionOpenAck(connectionID, connectionID, clientState, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, &types.Version{}, signer), false},
		{"empty signer", types.NewMsgConnectionOpenAck(connectionID, connectionID, clientState, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, nil), false},
		{"success", types.NewMsgConnectionOpenAck(connectionID, connectionID, clientState, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer), true},
	}

	for _, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *MsgTestSuite) TestNewMsgConnectionOpenConfirm() {
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	testMsgs := []*types.MsgConnectionOpenConfirm{
		types.NewMsgConnectionOpenConfirm("test/conn1", suite.proof, clientHeight, signer),
		types.NewMsgConnectionOpenConfirm(connectionID, emptyProof, clientHeight, signer),
		types.NewMsgConnectionOpenConfirm(connectionID, suite.proof, clienttypes.ZeroHeight(), signer),
		types.NewMsgConnectionOpenConfirm(connectionID, suite.proof, clientHeight, nil),
		types.NewMsgConnectionOpenConfirm(connectionID, suite.proof, clientHeight, signer),
	}

	var testCases = []struct {
		msg     *types.MsgConnectionOpenConfirm
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], false, "invalid connection ID"},
		{testMsgs[1], false, "empty proofTry"},
		{testMsgs[2], false, "invalid proofHeight"},
		{testMsgs[3], false, "empty signer"},
		{testMsgs[4], true, "success"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}
