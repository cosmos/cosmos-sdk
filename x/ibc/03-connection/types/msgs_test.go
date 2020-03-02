package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

type MsgTestSuite struct {
	suite.Suite

	proof commitmenttypes.MerkleProof
}

func (suite *MsgTestSuite) SetupTest() {
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

	suite.proof = commitmenttypes.MerkleProof{Proof: res.Proof}
}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}

func (suite *MsgTestSuite) TestNewMsgConnectionOpenInit() {
	prefix := commitmenttypes.NewMerklePrefix([]byte("storePrefixKey"))
	signer, _ := sdk.AccAddressFromBech32("cosmos1ckgw5d7jfj7wwxjzs9fdrdev9vc8dzcw3n2lht")

	testMsgs := []MsgConnectionOpenInit{
		NewMsgConnectionOpenInit("test/conn1", "clienttotesta", "connectiontotest", "clienttotest", prefix, signer),
		NewMsgConnectionOpenInit("ibcconntest", "test/iris", "connectiontotest", "clienttotest", prefix, signer),
		NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "test/conn1", "clienttotest", prefix, signer),
		NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "test/conn1", prefix, signer),
		NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "clienttotest", nil, signer),
		NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "clienttotest", prefix, nil),
		NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "clienttotest", prefix, signer),
	}

	var testCases = []struct {
		msg     MsgConnectionOpenInit
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

	testMsgs := []MsgConnectionOpenTry{
		NewMsgConnectionOpenTry("test/conn1", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconntest", "test/iris", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "ibc/test", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "test/conn1", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", nil, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{}, suite.proof, suite.proof, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, nil, suite.proof, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, commitmenttypes.MerkleProof{Proof: nil}, suite.proof, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, nil, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, commitmenttypes.MerkleProof{Proof: nil}, 10, 10, signer),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 0, 10, signer),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 0, signer),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, nil),
		NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer),
	}

	var testCases = []struct {
		msg     MsgConnectionOpenTry
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
		{testMsgs[7], false, "empty proofInit"},
		{testMsgs[8], false, "empty proofConsensus"},
		{testMsgs[9], false, "empty proofConsensus"},
		{testMsgs[10], false, "invalid proofHeight"},
		{testMsgs[11], false, "invalid consensusHeight"},
		{testMsgs[12], false, "empty singer"},
		{testMsgs[13], true, "success"},
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

	testMsgs := []MsgConnectionOpenAck{
		NewMsgConnectionOpenAck("test/conn1", suite.proof, suite.proof, 10, 10, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconntest", nil, suite.proof, 10, 10, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconntest", commitmenttypes.MerkleProof{Proof: nil}, suite.proof, 10, 10, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconntest", suite.proof, nil, 10, 10, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconntest", suite.proof, commitmenttypes.MerkleProof{Proof: nil}, 10, 10, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 0, 10, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 0, "1.0.0", signer),
		NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 10, "", signer),
		NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 10, "1.0.0", nil),
		NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 10, "1.0.0", signer),
	}
	var testCases = []struct {
		msg     MsgConnectionOpenAck
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], false, "invalid connection ID"},
		{testMsgs[1], false, "empty proofTry"},
		{testMsgs[2], false, "empty proofTry"},
		{testMsgs[3], false, "empty proofConsensus"},
		{testMsgs[4], false, "empty proofConsensus"},
		{testMsgs[5], false, "invalid proofHeight"},
		{testMsgs[6], false, "invalid consensusHeight"},
		{testMsgs[7], false, "invalid version"},
		{testMsgs[8], false, "empty signer"},
		{testMsgs[9], true, "success"},
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

	testMsgs := []MsgConnectionOpenConfirm{
		NewMsgConnectionOpenConfirm("test/conn1", suite.proof, 10, signer),
		NewMsgConnectionOpenConfirm("ibcconntest", nil, 10, signer),
		NewMsgConnectionOpenConfirm("ibcconntest", commitmenttypes.MerkleProof{Proof: nil}, 10, signer),
		NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, 0, signer),
		NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, 10, nil),
		NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, 10, signer),
	}

	var testCases = []struct {
		msg     MsgConnectionOpenConfirm
		expPass bool
		errMsg  string
	}{
		{testMsgs[0], false, "invalid connection ID"},
		{testMsgs[1], false, "empty proofTry"},
		{testMsgs[2], false, "empty proofTry"},
		{testMsgs[3], false, "invalid proofHeight"},
		{testMsgs[4], false, "empty signer"},
		{testMsgs[5], true, "success"},
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
