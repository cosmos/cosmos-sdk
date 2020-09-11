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
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

var (
	emptyPrefix = commitmenttypes.MerklePrefix{}
	emptyProof  = []byte{}
)

type MsgTestSuite struct {
	suite.Suite

	proof []byte
}

func (suite *MsgTestSuite) SetupTest() {
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

	merkleProof := commitmenttypes.MerkleProof{Proof: res.ProofOps}
	proof, err := app.AppCodec().MarshalBinaryBare(&merkleProof)
	suite.NoError(err)

	suite.proof = proof

}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}

func (suite *MsgTestSuite) TestNewMsgConnectionOpenInit() {
	prefix := commitmenttypes.NewMerklePrefix([]byte("storePrefixKey"))
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	testMsgs := []*types.MsgConnectionOpenInit{
		types.NewMsgConnectionOpenInit("test/conn1", "clienttotesta", "connectiontotest", "clienttotest", prefix, signer),
		types.NewMsgConnectionOpenInit("ibcconntest", "test/iris", "connectiontotest", "clienttotest", prefix, signer),
		types.NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "test/conn1", "clienttotest", prefix, signer),
		types.NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "test/conn1", prefix, signer),
		types.NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "clienttotest", emptyPrefix, signer),
		types.NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "clienttotest", prefix, nil),
		types.NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "clienttotest", prefix, signer),
	}

	var testCases = []struct {
		msg     *types.MsgConnectionOpenInit
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], false, "invalid connection ID"},
		{testMsgs[1], false, "invalid client ID"},
		{testMsgs[2], false, "invalid counterparty client ID"},
		{testMsgs[3], false, "invalid counterparty connection ID"},
		{testMsgs[4], false, "empty counterparty prefix"},
		{testMsgs[5], false, "empty singer"},
		{testMsgs[6], true, "success"},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			suite.Require().NoError(err, "Msg %d failed: %v", i, err)
		} else {
			suite.Require().Error(err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}

func (suite *MsgTestSuite) TestNewMsgConnectionOpenTry() {
	prefix := commitmenttypes.NewMerklePrefix([]byte("storePrefixKey"))
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	clientState := ibctmtypes.NewClientState(
		chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), false, false,
	)

	// Pack consensus state into any to test unpacking error
	consState := ibctmtypes.NewConsensusState(
		time.Now(), commitmenttypes.NewMerkleRoot([]byte("root")), clientHeight, []byte("nextValsHash"),
	)
	invalidAny := clienttypes.MustPackConsensusState(consState)
	counterparty := types.NewCounterparty("connectiontotest", "clienttotest", prefix)

	// invalidClientState fails validateBasic
	invalidClient := ibctmtypes.NewClientState(
		chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clienttypes.ZeroHeight(), commitmenttypes.GetSDKSpecs(), false, false,
	)

	testMsgs := []*types.MsgConnectionOpenTry{
		types.NewMsgConnectionOpenTry("test/conn1", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "test/iris", "connectiontotest", "clienttotest", clientState, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "ibc/test", "clienttotest", clientState, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "test/conn1", clientState, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", nil, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer),
		{"ibcconntest", "clienttotesta", invalidAny, counterparty, []string{ibctesting.ConnectionVersion}, clientHeight, suite.proof, suite.proof, suite.proof, clientHeight, signer},
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", invalidClient, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", clientState, emptyPrefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []string{}, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []string{ibctesting.ConnectionVersion}, emptyProof, suite.proof, suite.proof, clientHeight, clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, emptyProof, suite.proof, clientHeight, clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, emptyProof, clientHeight, clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, suite.proof, clienttypes.ZeroHeight(), clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, suite.proof, clientHeight, clienttypes.ZeroHeight(), signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, nil),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", clientState, prefix, []string{"(invalid version)"}, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, signer),
	}

	var testCases = []struct {
		msg     *types.MsgConnectionOpenTry
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], false, "invalid connection ID"},
		{testMsgs[1], false, "invalid client ID"},
		{testMsgs[2], false, "invalid counterparty connection ID"},
		{testMsgs[3], false, "invalid counterparty client ID"},
		{testMsgs[4], false, "invalid nil counterparty client"},
		{testMsgs[5], false, "invalid client unpacking"},
		{testMsgs[6], false, "counterparty failed Validate"},
		{testMsgs[7], false, "empty counterparty prefix"},
		{testMsgs[8], false, "empty counterpartyVersions"},
		{testMsgs[9], false, "empty proofInit"},
		{testMsgs[10], false, "empty proofClient"},
		{testMsgs[11], false, "empty proofConsensus"},
		{testMsgs[12], false, "invalid proofHeight"},
		{testMsgs[13], false, "invalid consensusHeight"},
		{testMsgs[14], false, "empty singer"},
		{testMsgs[15], true, "success"},
		{testMsgs[16], false, "invalid version"},
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

func (suite *MsgTestSuite) TestNewMsgConnectionOpenAck() {
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")
	clientState := ibctmtypes.NewClientState(
		chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), false, false,
	)

	// Pack consensus state into any to test unpacking error
	consState := ibctmtypes.NewConsensusState(
		time.Now(), commitmenttypes.NewMerkleRoot([]byte("root")), clientHeight, []byte("nextValsHash"),
	)
	invalidAny := clienttypes.MustPackConsensusState(consState)

	// invalidClientState fails validateBasic
	invalidClient := ibctmtypes.NewClientState(
		chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clienttypes.ZeroHeight(), commitmenttypes.GetSDKSpecs(), false, false,
	)

	testMsgs := []*types.MsgConnectionOpenAck{
		types.NewMsgConnectionOpenAck("test/conn1", clientState, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", nil, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer),
		{"ibcconntest", ibctesting.ConnectionVersion, invalidAny, clientHeight, suite.proof, suite.proof, suite.proof, clientHeight, signer},
		types.NewMsgConnectionOpenAck("ibcconntest", invalidClient, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", clientState, emptyProof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", clientState, suite.proof, emptyProof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", clientState, suite.proof, suite.proof, emptyProof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", clientState, suite.proof, suite.proof, suite.proof, clienttypes.ZeroHeight(), clientHeight, ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", clientState, suite.proof, suite.proof, suite.proof, clientHeight, clienttypes.ZeroHeight(), ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", clientState, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, "", signer),
		types.NewMsgConnectionOpenAck("ibcconntest", clientState, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, nil),
		types.NewMsgConnectionOpenAck("ibcconntest", clientState, suite.proof, suite.proof, suite.proof, clientHeight, clientHeight, ibctesting.ConnectionVersion, signer),
	}
	var testCases = []struct {
		msg     *types.MsgConnectionOpenAck
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], false, "invalid connection ID"},
		{testMsgs[1], false, "invalid nil counterparty client"},
		{testMsgs[2], false, "invalid unpacking counterparty client"},
		{testMsgs[3], false, "counterparty client failed Validate"},
		{testMsgs[4], false, "empty proofTry"},
		{testMsgs[5], false, "empty proofClient"},
		{testMsgs[6], false, "empty proofConsensus"},
		{testMsgs[7], false, "invalid proofHeight"},
		{testMsgs[8], false, "invalid consensusHeight"},
		{testMsgs[9], false, "invalid version"},
		{testMsgs[10], false, "empty signer"},
		{testMsgs[11], true, "success"},
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

func (suite *MsgTestSuite) TestNewMsgConnectionOpenConfirm() {
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	testMsgs := []*types.MsgConnectionOpenConfirm{
		types.NewMsgConnectionOpenConfirm("test/conn1", suite.proof, clientHeight, signer),
		types.NewMsgConnectionOpenConfirm("ibcconntest", emptyProof, clientHeight, signer),
		types.NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, clienttypes.ZeroHeight(), signer),
		types.NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, clientHeight, nil),
		types.NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, clientHeight, signer),
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
