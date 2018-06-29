package lcd

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	abci "github.com/tendermint/tendermint/abci/types"
	p2p "github.com/tendermint/tendermint/p2p"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tmlibs/common"

	client "github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	rpc "github.com/cosmos/cosmos-sdk/client/rpc"
	tests "github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/stake"
	stakerest "github.com/cosmos/cosmos-sdk/x/stake/client/rest"
)

func init() {
	cryptoKeys.BcryptSecurityParameter = 1
}

func TestKeys(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{addr})
	defer cleanup()

	// get seed
	// TODO Do we really need this endpoint?
	res, body := Request(t, port, "GET", "/keys/seed", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	reg, err := regexp.Compile(`([a-z]+ ){12}`)
	require.Nil(t, err)
	match := reg.MatchString(seed)
	assert.True(t, match, "Returned seed has wrong format", seed)

	newName := "test_newname"
	newPassword := "0987654321"

	// add key
	jsonStr := []byte(fmt.Sprintf(`{"name":"%s", "password":"%s"}`, newName, newPassword))
	res, body = Request(t, port, "POST", "/keys", jsonStr)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.NewKeyResponse
	err = wire.Cdc.UnmarshalJSON([]byte(body), &resp)
	require.Nil(t, err)
	addr2 := resp.Address
	assert.Len(t, addr2, 40, "Returned address has wrong format", addr2)

	// existing keys
	res, body = Request(t, port, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m [2]keys.KeyOutput
	err = cdc.UnmarshalJSON([]byte(body), &m)
	require.Nil(t, err)

	addr2Acc, err := sdk.GetAccAddressHex(addr2)
	require.Nil(t, err)
	addr2Bech32 := sdk.MustBech32ifyAcc(addr2Acc)
	addrBech32 := sdk.MustBech32ifyAcc(addr)

	assert.Equal(t, name, m[0].Name, "Did not serve keys name correctly")
	assert.Equal(t, addrBech32, m[0].Address, "Did not serve keys Address correctly")
	assert.Equal(t, newName, m[1].Name, "Did not serve keys name correctly")
	assert.Equal(t, addr2Bech32, m[1].Address, "Did not serve keys Address correctly")

	// select key
	keyEndpoint := fmt.Sprintf("/keys/%s", newName)
	res, body = Request(t, port, "GET", keyEndpoint, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m2 keys.KeyOutput
	err = cdc.UnmarshalJSON([]byte(body), &m2)
	require.Nil(t, err)

	assert.Equal(t, newName, m2.Name, "Did not serve keys name correctly")
	assert.Equal(t, addr2Bech32, m2.Address, "Did not serve keys Address correctly")

	// update key
	jsonStr = []byte(fmt.Sprintf(`{
		"old_password":"%s",
		"new_password":"12345678901"
	}`, newPassword))

	res, body = Request(t, port, "PUT", keyEndpoint, jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// here it should say unauthorized as we changed the password before
	res, body = Request(t, port, "PUT", keyEndpoint, jsonStr)
	require.Equal(t, http.StatusUnauthorized, res.StatusCode, body)

	// delete key
	jsonStr = []byte(`{"password":"12345678901"}`)
	res, body = Request(t, port, "DELETE", keyEndpoint, jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
}

func TestVersion(t *testing.T) {
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{})
	defer cleanup()

	// node info
	res, body := Request(t, port, "GET", "/version", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	reg, err := regexp.Compile(`\d+\.\d+\.\d+(-dev)?`)
	require.Nil(t, err)
	match := reg.MatchString(body)
	assert.True(t, match, body)

	// node info
	res, body = Request(t, port, "GET", "/node_version", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	reg, err = regexp.Compile(`\d+\.\d+\.\d+(-dev)?`)
	require.Nil(t, err)
	match = reg.MatchString(body)
	assert.True(t, match, body)
}

func TestNodeStatus(t *testing.T) {
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{})
	defer cleanup()

	// node info
	res, body := Request(t, port, "GET", "/node_info", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var nodeInfo p2p.NodeInfo
	err := cdc.UnmarshalJSON([]byte(body), &nodeInfo)
	require.Nil(t, err, "Couldn't parse node info")

	assert.NotEqual(t, p2p.NodeInfo{}, nodeInfo, "res: %v", res)

	// syncing
	res, body = Request(t, port, "GET", "/syncing", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// we expect that there is no other node running so the syncing state is "false"
	assert.Equal(t, "false", body)
}

func TestBlock(t *testing.T) {
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{})
	defer cleanup()

	var resultBlock ctypes.ResultBlock

	res, body := Request(t, port, "GET", "/blocks/latest", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultBlock)
	require.Nil(t, err, "Couldn't parse block")

	assert.NotEqual(t, ctypes.ResultBlock{}, resultBlock)

	// --

	res, body = Request(t, port, "GET", "/blocks/1", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = wire.Cdc.UnmarshalJSON([]byte(body), &resultBlock)
	require.Nil(t, err, "Couldn't parse block")

	assert.NotEqual(t, ctypes.ResultBlock{}, resultBlock)

	// --

	res, body = Request(t, port, "GET", "/blocks/1000000000", nil)
	require.Equal(t, http.StatusNotFound, res.StatusCode, body)
}

func TestValidators(t *testing.T) {
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{})
	defer cleanup()

	var resultVals rpc.ResultValidatorsOutput

	res, body := Request(t, port, "GET", "/validatorsets/latest", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultVals)
	require.Nil(t, err, "Couldn't parse validatorset")

	assert.NotEqual(t, rpc.ResultValidatorsOutput{}, resultVals)

	assert.Contains(t, resultVals.Validators[0].Address, "cosmosvaladdr")
	assert.Contains(t, resultVals.Validators[0].PubKey, "cosmosvalpub")

	// --

	res, body = Request(t, port, "GET", "/validatorsets/1", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultVals)
	require.Nil(t, err, "Couldn't parse validatorset")

	assert.NotEqual(t, rpc.ResultValidatorsOutput{}, resultVals)

	// --

	res, body = Request(t, port, "GET", "/validatorsets/1000000000", nil)
	require.Equal(t, http.StatusNotFound, res.StatusCode, body)
}

func TestCoinSend(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{addr})
	defer cleanup()

	bz, err := hex.DecodeString("8FA6AB57AD6870F6B5B2E57735F38F2F30E73CB6")
	require.NoError(t, err)
	someFakeAddr := sdk.MustBech32ifyAcc(bz)

	// query empty
	res, body := Request(t, port, "GET", "/accounts/"+someFakeAddr, nil)
	require.Equal(t, http.StatusNoContent, res.StatusCode, body)

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create TX
	receiveAddr, resultTx := doSend(t, port, seed, name, password, addr)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	mycoins := coins[0]

	assert.Equal(t, "steak", mycoins.Denom)
	assert.Equal(t, initialBalance[0].Amount.SubRaw(1), mycoins.Amount)

	// query receiver
	acc = getAccount(t, port, receiveAddr)
	coins = acc.GetCoins()
	mycoins = coins[0]

	assert.Equal(t, "steak", mycoins.Denom)
	assert.Equal(t, int64(1), mycoins.Amount.Int64())
}

func TestIBCTransfer(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{addr})
	defer cleanup()

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create TX
	resultTx := doIBCTransfer(t, port, seed, name, password, addr)

	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	mycoins := coins[0]

	assert.Equal(t, "steak", mycoins.Denom)
	assert.Equal(t, initialBalance[0].Amount.SubRaw(1), mycoins.Amount)

	// TODO: query ibc egress packet state
}

func TestTxs(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{addr})
	defer cleanup()

	// query wrong
	res, body := Request(t, port, "GET", "/txs", nil)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, body)

	// query empty
	res, body = Request(t, port, "GET", fmt.Sprintf("/txs?tag=sender_bech32='%s'", "cosmosaccaddr1jawd35d9aq4u76sr3fjalmcqc8hqygs9gtnmv3"), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	assert.Equal(t, "[]", body)

	// create TX
	receiveAddr, resultTx := doSend(t, port, seed, name, password, addr)

	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx is findable
	res, body = Request(t, port, "GET", fmt.Sprintf("/txs/%s", resultTx.Hash), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	type txInfo struct {
		Hash   common.HexBytes        `json:"hash"`
		Height int64                  `json:"height"`
		Tx     sdk.Tx                 `json:"tx"`
		Result abci.ResponseDeliverTx `json:"result"`
	}
	var indexedTxs []txInfo

	// check if tx is queryable
	res, body = Request(t, port, "GET", fmt.Sprintf("/txs?tag=tx.hash='%s'", resultTx.Hash), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	assert.NotEqual(t, "[]", body)

	err := cdc.UnmarshalJSON([]byte(body), &indexedTxs)
	require.NoError(t, err)
	assert.Equal(t, 1, len(indexedTxs))

	// XXX should this move into some other testfile for txs in general?
	// test if created TX hash is the correct hash
	assert.Equal(t, resultTx.Hash, indexedTxs[0].Hash)

	// query sender
	// also tests url decoding
	addrBech := sdk.MustBech32ifyAcc(addr)
	res, body = Request(t, port, "GET", "/txs?tag=sender_bech32=%27"+addrBech+"%27", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &indexedTxs)
	require.NoError(t, err)
	require.Equal(t, 1, len(indexedTxs), "%v", indexedTxs) // there are 2 txs created with doSend
	assert.Equal(t, resultTx.Height, indexedTxs[0].Height)

	// query recipient
	receiveAddrBech := sdk.MustBech32ifyAcc(receiveAddr)
	res, body = Request(t, port, "GET", fmt.Sprintf("/txs?tag=recipient_bech32='%s'", receiveAddrBech), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &indexedTxs)
	require.NoError(t, err)
	require.Equal(t, 1, len(indexedTxs))
	assert.Equal(t, resultTx.Height, indexedTxs[0].Height)
}

func TestValidatorsQuery(t *testing.T) {
	cleanup, pks, port := InitializeTestLCD(t, 2, []sdk.Address{})
	defer cleanup()
	require.Equal(t, 2, len(pks))

	validators := getValidators(t, port)
	assert.Equal(t, len(validators), 2)

	// make sure all the validators were found (order unknown because sorted by owner addr)
	foundVal1, foundVal2 := false, false
	pk1Bech := sdk.MustBech32ifyValPub(pks[0])
	pk2Bech := sdk.MustBech32ifyValPub(pks[1])
	if validators[0].PubKey == pk1Bech || validators[1].PubKey == pk1Bech {
		foundVal1 = true
	}
	if validators[0].PubKey == pk2Bech || validators[1].PubKey == pk2Bech {
		foundVal2 = true
	}
	assert.True(t, foundVal1, "pk1Bech %v, owner1 %v, owner2 %v", pk1Bech, validators[0].Owner, validators[1].Owner)
	assert.True(t, foundVal2, "pk2Bech %v, owner1 %v, owner2 %v", pk2Bech, validators[0].Owner, validators[1].Owner)
}

func TestBonding(t *testing.T) {
	name, password, denom := "test", "1234567890", "steak"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, pks, port := InitializeTestLCD(t, 1, []sdk.Address{addr})
	defer cleanup()

	validator1Owner := pks[0].Address()

	// create bond TX
	resultTx := doDelegate(t, port, seed, name, password, addr, validator1Owner)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc := getAccount(t, port, addr)
	coins := acc.GetCoins()

	assert.Equal(t, int64(40), coins.AmountOf(denom).Int64())

	// query validator
	bond := getDelegation(t, port, addr, validator1Owner)
	assert.Equal(t, "60/1", bond.Shares.String())

	//////////////////////
	// testing unbonding

	// create unbond TX
	resultTx = doBeginUnbonding(t, port, seed, name, password, addr, validator1Owner)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query validator
	bond = getDelegation(t, port, addr, validator1Owner)
	assert.Equal(t, "30/1", bond.Shares.String())

	// check if tx was committed
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// should the sender should have not received any coins as the unbonding has only just begun
	// query sender
	acc = getAccount(t, port, addr)
	coins = acc.GetCoins()
	assert.Equal(t, int64(40), coins.AmountOf("steak").Int64())

	// TODO add redelegation, need more complex capabilities such to mock context and
}

func TestSubmitProposal(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{addr})
	defer cleanup()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name, password, addr)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID int64
	cdc.UnmarshalBinaryBare(resultTx.DeliverTx.GetData(), &proposalID)

	// query proposal
	proposal := getProposal(t, port, proposalID)
	assert.Equal(t, "Test", proposal.Title)
}

func TestDeposit(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{addr})
	defer cleanup()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name, password, addr)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID int64
	cdc.UnmarshalBinaryBare(resultTx.DeliverTx.GetData(), &proposalID)

	// query proposal
	proposal := getProposal(t, port, proposalID)
	assert.Equal(t, "Test", proposal.Title)

	// create SubmitProposal TX
	resultTx = doDeposit(t, port, seed, name, password, addr, proposalID)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query proposal
	proposal = getProposal(t, port, proposalID)
	assert.True(t, proposal.TotalDeposit.IsEqual(sdk.Coins{sdk.NewCoin("steak", 10)}))

	// query deposit
	deposit := getDeposit(t, port, proposalID, addr)
	assert.True(t, deposit.Amount.IsEqual(sdk.Coins{sdk.NewCoin("steak", 10)}))
}

func TestVote(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{addr})
	defer cleanup()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name, password, addr)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID int64
	cdc.UnmarshalBinaryBare(resultTx.DeliverTx.GetData(), &proposalID)

	// query proposal
	proposal := getProposal(t, port, proposalID)
	assert.Equal(t, "Test", proposal.Title)

	// create SubmitProposal TX
	resultTx = doDeposit(t, port, seed, name, password, addr, proposalID)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query proposal
	proposal = getProposal(t, port, proposalID)
	assert.Equal(t, gov.StatusToString(gov.StatusVotingPeriod), proposal.Status)

	// create SubmitProposal TX
	resultTx = doVote(t, port, seed, name, password, addr, proposalID)
	tests.WaitForHeight(resultTx.Height+1, port)

	vote := getVote(t, port, proposalID, addr)
	assert.Equal(t, proposalID, vote.ProposalID)
	assert.Equal(t, gov.VoteOptionToString(gov.OptionYes), vote.Option)
}

func TestProposalsQuery(t *testing.T) {
	name, password1 := "test", "1234567890"
	name2, password2 := "test2", "1234567890"
	addr, seed := CreateAddr(t, "test", password1, GetKB(t))
	addr2, seed2 := CreateAddr(t, "test2", password2, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.Address{addr, addr2})
	defer cleanup()

	// Addr1 proposes (and deposits) proposals #1 and #2
	resultTx := doSubmitProposal(t, port, seed, name, password1, addr)
	var proposalID1 int64
	cdc.UnmarshalBinaryBare(resultTx.DeliverTx.GetData(), &proposalID1)
	tests.WaitForHeight(resultTx.Height+1, port)
	resultTx = doSubmitProposal(t, port, seed, name, password1, addr)
	var proposalID2 int64
	cdc.UnmarshalBinaryBare(resultTx.DeliverTx.GetData(), &proposalID2)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 proposes (and deposits) proposals #3
	resultTx = doSubmitProposal(t, port, seed2, name2, password2, addr2)
	var proposalID3 int64
	cdc.UnmarshalBinaryBare(resultTx.DeliverTx.GetData(), &proposalID3)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 deposits on proposals #2 & #3
	resultTx = doDeposit(t, port, seed2, name2, password2, addr2, proposalID2)
	tests.WaitForHeight(resultTx.Height+1, port)
	resultTx = doDeposit(t, port, seed2, name2, password2, addr2, proposalID3)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr1 votes on proposals #2 & #3
	resultTx = doVote(t, port, seed, name, password1, addr, proposalID2)
	tests.WaitForHeight(resultTx.Height+1, port)
	resultTx = doVote(t, port, seed, name, password1, addr, proposalID3)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 votes on proposal #3
	resultTx = doVote(t, port, seed2, name2, password2, addr2, proposalID3)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Test query all proposals
	proposals := getProposalsAll(t, port)
	assert.Equal(t, proposalID1, (proposals[0]).ProposalID)
	assert.Equal(t, proposalID2, (proposals[1]).ProposalID)
	assert.Equal(t, proposalID3, (proposals[2]).ProposalID)

	// Test query deposited by addr1
	proposals = getProposalsFilterDepositer(t, port, addr)
	assert.Equal(t, proposalID1, (proposals[0]).ProposalID)

	// Test query deposited by addr2
	proposals = getProposalsFilterDepositer(t, port, addr2)
	assert.Equal(t, proposalID2, (proposals[0]).ProposalID)
	assert.Equal(t, proposalID3, (proposals[1]).ProposalID)

	// Test query voted by addr1
	proposals = getProposalsFilterVoter(t, port, addr)
	assert.Equal(t, proposalID2, (proposals[0]).ProposalID)
	assert.Equal(t, proposalID3, (proposals[1]).ProposalID)

	// Test query voted by addr2
	proposals = getProposalsFilterVoter(t, port, addr2)
	assert.Equal(t, proposalID3, (proposals[0]).ProposalID)

	// Test query voted and deposited by addr1
	proposals = getProposalsFilterVoterDepositer(t, port, addr, addr)
	assert.Equal(t, proposalID2, (proposals[0]).ProposalID)
}

//_____________________________________________________________________________
// get the account to get the sequence
func getAccount(t *testing.T, port string, addr sdk.Address) auth.Account {
	addrBech32 := sdk.MustBech32ifyAcc(addr)
	res, body := Request(t, port, "GET", "/accounts/"+addrBech32, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var acc auth.Account
	err := cdc.UnmarshalJSON([]byte(body), &acc)
	require.Nil(t, err)
	return acc
}

func doSend(t *testing.T, port, seed, name, password string, addr sdk.Address) (receiveAddr sdk.Address, resultTx ctypes.ResultBroadcastTxCommit) {

	// create receive address
	kb := client.MockKeyBase()
	receiveInfo, _, err := kb.CreateMnemonic("receive_address", cryptoKeys.English, "1234567890", cryptoKeys.SigningAlgo("secp256k1"))
	require.Nil(t, err)
	receiveAddr = receiveInfo.GetPubKey().Address()
	receiveAddrBech := sdk.MustBech32ifyAcc(receiveAddr)

	acc := getAccount(t, port, addr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)

	// send
	coinbz, err := cdc.MarshalJSON(sdk.NewCoin("steak", 1))
	if err != nil {
		panic(err)
	}

	jsonStr := []byte(fmt.Sprintf(`{
		"name":"%s",
		"password":"%s",
		"account_number":"%d",
		"sequence":"%d",
		"gas": "10000",
		"amount":[%s],
		"chain_id":"%s"
	}`, name, password, accnum, sequence, coinbz, chainID))
	res, body := Request(t, port, "POST", "/accounts/"+receiveAddrBech+"/send", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	return receiveAddr, resultTx
}

func doIBCTransfer(t *testing.T, port, seed, name, password string, addr sdk.Address) (resultTx ctypes.ResultBroadcastTxCommit) {
	// create receive address
	kb := client.MockKeyBase()
	receiveInfo, _, err := kb.CreateMnemonic("receive_address", cryptoKeys.English, "1234567890", cryptoKeys.SigningAlgo("secp256k1"))
	require.Nil(t, err)
	receiveAddr := receiveInfo.GetPubKey().Address()
	receiveAddrBech := sdk.MustBech32ifyAcc(receiveAddr)

	chainID := viper.GetString(client.FlagChainID)

	// get the account to get the sequence
	acc := getAccount(t, port, addr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	// send
	jsonStr := []byte(fmt.Sprintf(`{
		"name":"%s",
		"password": "%s",
		"account_number":"%d",
		"sequence": "%d",
		"gas": "100000",
		"chain_id": "%s",
		"amount":[
			{
				"denom": "%s",
				"amount": "1"
			}
		]
	}`, name, password, accnum, sequence, chainID, "steak"))
	res, body := Request(t, port, "POST", "/ibc/testchain/"+receiveAddrBech+"/send", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	return resultTx
}

func getDelegation(t *testing.T, port string, delegatorAddr, validatorAddr sdk.Address) stake.Delegation {

	delegatorAddrBech := sdk.MustBech32ifyAcc(delegatorAddr)
	validatorAddrBech := sdk.MustBech32ifyVal(validatorAddr)

	// get the account to get the sequence
	res, body := Request(t, port, "GET", "/stake/"+delegatorAddrBech+"/delegation/"+validatorAddrBech, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var bond stake.Delegation
	err := cdc.UnmarshalJSON([]byte(body), &bond)
	require.Nil(t, err)
	return bond
}

func doDelegate(t *testing.T, port, seed, name, password string, delegatorAddr, validatorAddr sdk.Address) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, port, delegatorAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	delegatorAddrBech := sdk.MustBech32ifyAcc(delegatorAddr)
	validatorAddrBech := sdk.MustBech32ifyVal(validatorAddr)

	chainID := viper.GetString(client.FlagChainID)

	// send
	jsonStr := []byte(fmt.Sprintf(`{
		"name": "%s",
		"password": "%s",
		"account_number": "%d",
		"sequence": "%d",
		"gas": "10000",
		"chain_id": "%s",
		"delegations": [
			{
				"delegator_addr": "%s",
				"validator_addr": "%s",
				"bond": { "denom": "%s", "amount": "60" }
			}
		],
		"begin_unbondings": [],
		"complete_unbondings": [],
		"begin_redelegates": [],
		"complete_redelegates": []
	}`, name, password, accnum, sequence, chainID, delegatorAddrBech, validatorAddrBech, "steak"))
	res, body := Request(t, port, "POST", "/stake/delegations", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func doBeginUnbonding(t *testing.T, port, seed, name, password string,
	delegatorAddr, validatorAddr sdk.Address) (resultTx ctypes.ResultBroadcastTxCommit) {

	// get the account to get the sequence
	acc := getAccount(t, port, delegatorAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	delegatorAddrBech := sdk.MustBech32ifyAcc(delegatorAddr)
	validatorAddrBech := sdk.MustBech32ifyVal(validatorAddr)

	chainID := viper.GetString(client.FlagChainID)

	// send
	jsonStr := []byte(fmt.Sprintf(`{
		"name": "%s",
		"password": "%s",
		"account_number": "%d",
		"sequence": "%d",
		"gas": "10000",
		"chain_id": "%s",
		"delegations": [],
		"begin_unbondings": [
			{
				"delegator_addr": "%s",
				"validator_addr": "%s",
				"shares": "30"
			}
		],
		"complete_unbondings": [],
		"begin_redelegates": [],
		"complete_redelegates": []
	}`, name, password, accnum, sequence, chainID, delegatorAddrBech, validatorAddrBech))
	res, body := Request(t, port, "POST", "/stake/delegations", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func doBeginRedelegation(t *testing.T, port, seed, name, password string,
	delegatorAddr, validatorSrcAddr, validatorDstAddr sdk.Address) (resultTx ctypes.ResultBroadcastTxCommit) {

	// get the account to get the sequence
	acc := getAccount(t, port, delegatorAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	delegatorAddrBech := sdk.MustBech32ifyAcc(delegatorAddr)
	validatorSrcAddrBech := sdk.MustBech32ifyVal(validatorSrcAddr)
	validatorDstAddrBech := sdk.MustBech32ifyVal(validatorDstAddr)

	chainID := viper.GetString(client.FlagChainID)

	// send
	jsonStr := []byte(fmt.Sprintf(`{
		"name": "%s",
		"password": "%s",
		"account_number": "%d",
		"sequence": "%d",
		"gas": "10000",
		"chain_id": "%s",
		"delegations": [],
		"begin_unbondings": [],
		"complete_unbondings": [],
		"begin_redelegates": [
			{
				"delegator_addr": "%s",
				"validator_src_addr": "%s",
				"validator_dst_addr": "%s",
				"shares": "30"
			}
		],
		"complete_redelegates": []
	}`, name, password, accnum, sequence, chainID, delegatorAddrBech, validatorSrcAddrBech, validatorDstAddrBech))
	res, body := Request(t, port, "POST", "/stake/delegations", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func getValidators(t *testing.T, port string) []stakerest.StakeValidatorOutput {
	// get the account to get the sequence
	res, body := Request(t, port, "GET", "/stake/validators", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var validators []stakerest.StakeValidatorOutput
	err := cdc.UnmarshalJSON([]byte(body), &validators)
	require.Nil(t, err)
	return validators
}

func getProposal(t *testing.T, port string, proposalID int64) gov.ProposalRest {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var proposal gov.ProposalRest
	err := cdc.UnmarshalJSON([]byte(body), &proposal)
	require.Nil(t, err)
	return proposal
}

func getDeposit(t *testing.T, port string, proposalID int64, depositerAddr sdk.Address) gov.DepositRest {
	bechDepositerAddr := sdk.MustBech32ifyAcc(depositerAddr)
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/deposits/%s", proposalID, bechDepositerAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var deposit gov.DepositRest
	err := cdc.UnmarshalJSON([]byte(body), &deposit)
	require.Nil(t, err)
	return deposit
}

func getVote(t *testing.T, port string, proposalID int64, voterAddr sdk.Address) gov.VoteRest {
	bechVoterAddr := sdk.MustBech32ifyAcc(voterAddr)
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/votes/%s", proposalID, bechVoterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var vote gov.VoteRest
	err := cdc.UnmarshalJSON([]byte(body), &vote)
	require.Nil(t, err)
	return vote
}

func getProposalsAll(t *testing.T, port string) []gov.ProposalRest {
	res, body := Request(t, port, "GET", "/gov/proposals", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.ProposalRest
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

func getProposalsFilterDepositer(t *testing.T, port string, depositerAddr sdk.Address) []gov.ProposalRest {
	bechDepositerAddr := sdk.MustBech32ifyAcc(depositerAddr)

	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?depositer=%s", bechDepositerAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.ProposalRest
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

func getProposalsFilterVoter(t *testing.T, port string, voterAddr sdk.Address) []gov.ProposalRest {
	bechVoterAddr := sdk.MustBech32ifyAcc(voterAddr)

	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?voter=%s", bechVoterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.ProposalRest
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

func getProposalsFilterVoterDepositer(t *testing.T, port string, voterAddr sdk.Address, depositerAddr sdk.Address) []gov.ProposalRest {
	bechVoterAddr := sdk.MustBech32ifyAcc(voterAddr)
	bechDepositerAddr := sdk.MustBech32ifyAcc(depositerAddr)

	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?depositer=%s&voter=%s", bechDepositerAddr, bechVoterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.ProposalRest
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

func doSubmitProposal(t *testing.T, port, seed, name, password string, proposerAddr sdk.Address) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)

	bechProposerAddr := sdk.MustBech32ifyAcc(proposerAddr)

	// submitproposal
	jsonStr := []byte(fmt.Sprintf(`{
		"title": "Test",
		"description": "test",
		"proposal_type": "Text",
		"proposer": "%s",
		"initial_deposit": [{ "denom": "steak", "amount": "5" }],
		"base_req": {
			"name": "%s",
			"password": "%s",
			"chain_id": "%s",
			"account_number":"%d",
			"sequence":"%d",
			"gas":"100000"
		}
	}`, bechProposerAddr, name, password, chainID, accnum, sequence))
	res, body := Request(t, port, "POST", "/gov/proposals", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

func doDeposit(t *testing.T, port, seed, name, password string, proposerAddr sdk.Address, proposalID int64) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)

	bechProposerAddr := sdk.MustBech32ifyAcc(proposerAddr)

	// deposit on proposal
	jsonStr := []byte(fmt.Sprintf(`{
		"depositer": "%s",
		"amount": [{ "denom": "steak", "amount": "5" }],
		"base_req": {
			"name": "%s",
			"password": "%s",
			"chain_id": "%s",
			"account_number":"%d",
			"sequence": "%d",
			"gas":"100000"
		}
	}`, bechProposerAddr, name, password, chainID, accnum, sequence))
	res, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/deposits", proposalID), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

func doVote(t *testing.T, port, seed, name, password string, proposerAddr sdk.Address, proposalID int64) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)

	bechProposerAddr := sdk.MustBech32ifyAcc(proposerAddr)

	// vote on proposal
	jsonStr := []byte(fmt.Sprintf(`{
		"voter": "%s",
		"option": "Yes",
		"base_req": {
			"name": "%s",
			"password": "%s",
			"chain_id": "%s",
			"account_number": "%d",
			"sequence": "%d",
			"gas":"100000"
		}
	}`, bechProposerAddr, name, password, chainID, accnum, sequence))
	res, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/votes", proposalID), jsonStr)
	fmt.Println(res)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}
