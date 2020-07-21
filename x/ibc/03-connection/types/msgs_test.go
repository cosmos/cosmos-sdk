package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
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

	merkleProof := commitmenttypes.MerkleProof{Proof: res.Proof}
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

	testMsgs := []*types.MsgConnectionOpenTry{
		types.NewMsgConnectionOpenTry("test/conn1", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, 10, 10, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "test/iris", "connectiontotest", "clienttotest", prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, 10, 10, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "ibc/test", "clienttotest", prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, 10, 10, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "test/conn1", prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, 10, 10, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", emptyPrefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, 10, 10, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{}, suite.proof, suite.proof, 10, 10, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{ibctesting.ConnectionVersion}, emptyProof, suite.proof, 10, 10, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{ibctesting.ConnectionVersion}, suite.proof, emptyProof, 10, 10, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, 0, 10, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, 10, 0, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, 10, 10, nil),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{ibctesting.ConnectionVersion}, suite.proof, suite.proof, 10, 10, signer),
		types.NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"(invalid version)"}, suite.proof, suite.proof, 10, 10, signer),
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
		{testMsgs[4], false, "empty counterparty prefix"},
		{testMsgs[5], false, "empty counterpartyVersions"},
		{testMsgs[6], false, "empty proofInit"},
		{testMsgs[7], false, "empty proofConsensus"},
		{testMsgs[8], false, "invalid proofHeight"},
		{testMsgs[9], false, "invalid consensusHeight"},
		{testMsgs[10], false, "empty singer"},
		{testMsgs[11], true, "success"},
		{testMsgs[12], false, "invalid version"},
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

	testMsgs := []*types.MsgConnectionOpenAck{
		types.NewMsgConnectionOpenAck("test/conn1", suite.proof, suite.proof, 10, 10, ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", emptyProof, suite.proof, 10, 10, ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", suite.proof, emptyProof, 10, 10, ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 0, 10, ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 0, ibctesting.ConnectionVersion, signer),
		types.NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 10, "", signer),
		types.NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 10, ibctesting.ConnectionVersion, nil),
		types.NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 10, ibctesting.ConnectionVersion, signer),
	}
	var testCases = []struct {
		msg     *types.MsgConnectionOpenAck
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], false, "invalid connection ID"},
		{testMsgs[1], false, "empty proofTry"},
		{testMsgs[2], false, "empty proofConsensus"},
		{testMsgs[3], false, "invalid proofHeight"},
		{testMsgs[4], false, "invalid consensusHeight"},
		{testMsgs[5], false, "invalid version"},
		{testMsgs[6], false, "empty signer"},
		{testMsgs[7], true, "success"},
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
		types.NewMsgConnectionOpenConfirm("test/conn1", suite.proof, 10, signer),
		types.NewMsgConnectionOpenConfirm("ibcconntest", emptyProof, 10, signer),
		types.NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, 0, signer),
		types.NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, 10, nil),
		types.NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, 10, signer),
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
