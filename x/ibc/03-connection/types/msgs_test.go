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

var (
	emptyPrefix = commitmenttypes.MerklePrefix{}
	emptyProof  = commitmenttypes.MerkleProof{Proof: nil}
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

	msg0, err := NewMsgConnectionOpenInit("test/conn1", "clienttotesta", "connectiontotest", "clienttotest", prefix, signer)
	suite.Require().NoError(err)
	msg1, err := NewMsgConnectionOpenInit("ibcconntest", "test/iris", "connectiontotest", "clienttotest", prefix, signer)
	suite.Require().NoError(err)
	msg2, err := NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "test/conn1", "clienttotest", prefix, signer)
	suite.Require().NoError(err)
	msg3, err := NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "test/conn1", prefix, signer)
	suite.Require().NoError(err)
	msg4, err := NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "clienttotest", emptyPrefix, signer)
	suite.Require().NoError(err)
	msg5, err := NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "clienttotest", prefix, nil)
	suite.Require().NoError(err)
	msg6, err := NewMsgConnectionOpenInit("ibcconntest", "clienttotest", "connectiontotest", "clienttotest", prefix, signer)
	suite.Require().NoError(err)

	var testCases = []struct {
		msg     MsgConnectionOpenInit
		expPass bool
		errMsg  string
	}{
		{msg0, false, "invalid connection ID"},
		{msg1, false, "invalid client ID"},
		{msg2, false, "invalid counterparty client ID"},
		{msg3, false, "invalid counterparty connection ID"},
		{msg4, false, "empty counterparty prefix"},
		{msg5, false, "empty singer"},
		{msg6, true, "success"},
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

	msg0, err := NewMsgConnectionOpenTry("test/conn1", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer)
	suite.Require().NoError(err)
	msg1, err := NewMsgConnectionOpenTry("ibcconntest", "test/iris", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer)
	suite.Require().NoError(err)
	msg2, err := NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "ibc/test", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer)
	suite.Require().NoError(err)
	msg3, err := NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "test/conn1", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer)
	suite.Require().NoError(err)
	msg4, err := NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", emptyPrefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer)
	suite.Require().NoError(err)
	msg5, err := NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{}, suite.proof, suite.proof, 10, 10, signer)
	suite.Require().NoError(err)
	msg6, err := NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, emptyProof, suite.proof, 10, 10, signer)
	suite.Require().NoError(err)
	msg7, err := NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, emptyProof, 10, 10, signer)
	suite.Require().NoError(err)
	msg8, err := NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 0, 10, signer)
	suite.Require().NoError(err)
	msg9, err := NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 0, signer)
	suite.Require().NoError(err)
	msg10, err := NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, nil)
	suite.Require().NoError(err)
	msg11, err := NewMsgConnectionOpenTry("ibcconntest", "clienttotesta", "connectiontotest", "clienttotest", prefix, []string{"1.0.0"}, suite.proof, suite.proof, 10, 10, signer)
	suite.Require().NoError(err)

	var testCases = []struct {
		msg     MsgConnectionOpenTry
		expPass bool
		errMsg  string
	}{
		{msg0, false, "invalid connection ID"},
		{msg1, false, "invalid client ID"},
		{msg2, false, "invalid counterparty connection ID"},
		{msg3, false, "invalid counterparty client ID"},
		{msg4, false, "empty counterparty prefix"},
		{msg5, false, "empty counterpartyVersions"},
		{msg6, false, "empty proofInit"},
		{msg7, false, "empty proofConsensus"},
		{msg8, false, "invalid proofHeight"},
		{msg9, false, "invalid consensusHeight"},
		{msg10, false, "empty singer"},
		{msg11, true, "success"},
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

	msg0, err := NewMsgConnectionOpenAck("test/conn1", suite.proof, suite.proof, 10, 10, "1.0.0", signer)
	suite.Require().NoError(err)
	msg1, err := NewMsgConnectionOpenAck("ibcconntest", emptyProof, suite.proof, 10, 10, "1.0.0", signer)
	suite.Require().NoError(err)
	msg2, err := NewMsgConnectionOpenAck("ibcconntest", suite.proof, emptyProof, 10, 10, "1.0.0", signer)
	suite.Require().NoError(err)
	msg3, err := NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 0, 10, "1.0.0", signer)
	suite.Require().NoError(err)
	msg4, err := NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 0, "1.0.0", signer)
	suite.Require().NoError(err)
	msg5, err := NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 10, "", signer)
	suite.Require().NoError(err)
	msg6, err := NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 10, "1.0.0", nil)
	suite.Require().NoError(err)
	msg7, err := NewMsgConnectionOpenAck("ibcconntest", suite.proof, suite.proof, 10, 10, "1.0.0", signer)
	suite.Require().NoError(err)

	var testCases = []struct {
		msg     MsgConnectionOpenAck
		expPass bool
		errMsg  string
	}{
		{msg0, false, "invalid connection ID"},
		{msg1, false, "empty proofTry"},
		{msg2, false, "empty proofConsensus"},
		{msg3, false, "invalid proofHeight"},
		{msg4, false, "invalid consensusHeight"},
		{msg5, false, "invalid version"},
		{msg6, false, "empty signer"},
		{msg7, true, "success"},
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

	msg0, err := NewMsgConnectionOpenConfirm("test/conn1", suite.proof, 10, signer)
	suite.Require().NoError(err)
	msg1, err := NewMsgConnectionOpenConfirm("ibcconntest", emptyProof, 10, signer)
	suite.Require().NoError(err)
	msg2, err := NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, 0, signer)
	suite.Require().NoError(err)
	msg3, err := NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, 10, nil)
	suite.Require().NoError(err)
	msg4, err := NewMsgConnectionOpenConfirm("ibcconntest", suite.proof, 10, signer)
	suite.Require().NoError(err)

	var testCases = []struct {
		msg     MsgConnectionOpenConfirm
		expPass bool
		errMsg  string
	}{
		{msg0, false, "invalid connection ID"},
		{msg1, false, "empty proofTry"},
		{msg2, false, "invalid proofHeight"},
		{msg3, false, "empty signer"},
		{msg4, true, "success"},
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
