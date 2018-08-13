package lcd

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	p2p "github.com/tendermint/tendermint/p2p"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	client "github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	rpc "github.com/cosmos/cosmos-sdk/client/rpc"
	tests "github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/client/rest"
)

func init() {
	cryptoKeys.BcryptSecurityParameter = 1
}

func TestKeys(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// get seed
	// TODO Do we really need this endpoint?
	res, body := Request(t, port, "GET", "/keys/seed", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	reg, err := regexp.Compile(`([a-z]+ ){12}`)
	require.Nil(t, err)
	match := reg.MatchString(seed)
	require.True(t, match, "Returned seed has wrong format", seed)

	newName := "test_newname"
	newPassword := "0987654321"

	// add key
	jsonStr := []byte(fmt.Sprintf(`{"name":"%s", "password":"%s", "seed":"%s"}`, newName, newPassword, seed))
	res, body = Request(t, port, "POST", "/keys", jsonStr)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.KeyOutput
	err = wire.Cdc.UnmarshalJSON([]byte(body), &resp)
	require.Nil(t, err, body)

	addr2Bech32 := resp.Address.String()
	_, err = sdk.AccAddressFromBech32(addr2Bech32)
	require.NoError(t, err, "Failed to return a correct bech32 address")

	// test if created account is the correct account
	expectedInfo, _ := GetKeyBase(t).CreateKey(newName, seed, newPassword)
	expectedAccount := sdk.AccAddress(expectedInfo.GetPubKey().Address().Bytes())
	assert.Equal(t, expectedAccount.String(), addr2Bech32)

	// existing keys
	res, body = Request(t, port, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m [2]keys.KeyOutput
	err = cdc.UnmarshalJSON([]byte(body), &m)
	require.Nil(t, err)

	addrBech32 := addr.String()

	require.Equal(t, name, m[0].Name, "Did not serve keys name correctly")
	require.Equal(t, addrBech32, m[0].Address.String(), "Did not serve keys Address correctly")
	require.Equal(t, newName, m[1].Name, "Did not serve keys name correctly")
	require.Equal(t, addr2Bech32, m[1].Address.String(), "Did not serve keys Address correctly")

	// select key
	keyEndpoint := fmt.Sprintf("/keys/%s", newName)
	res, body = Request(t, port, "GET", keyEndpoint, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m2 keys.KeyOutput
	err = cdc.UnmarshalJSON([]byte(body), &m2)
	require.Nil(t, err)

	require.Equal(t, newName, m2.Name, "Did not serve keys name correctly")
	require.Equal(t, addr2Bech32, m2.Address.String(), "Did not serve keys Address correctly")

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
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()

	// node info
	res, body := Request(t, port, "GET", "/version", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	reg, err := regexp.Compile(`\d+\.\d+\.\d+(-dev)?`)
	require.Nil(t, err)
	match := reg.MatchString(body)
	require.True(t, match, body)

	// node info
	res, body = Request(t, port, "GET", "/node_version", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	reg, err = regexp.Compile(`\d+\.\d+\.\d+(-dev)?`)
	require.Nil(t, err)
	match = reg.MatchString(body)
	require.True(t, match, body)
}

func TestNodeStatus(t *testing.T) {
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()

	// node info
	res, body := Request(t, port, "GET", "/node_info", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var nodeInfo p2p.NodeInfo
	err := cdc.UnmarshalJSON([]byte(body), &nodeInfo)
	require.Nil(t, err, "Couldn't parse node info")

	require.NotEqual(t, p2p.NodeInfo{}, nodeInfo, "res: %v", res)

	// syncing
	res, body = Request(t, port, "GET", "/syncing", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// we expect that there is no other node running so the syncing state is "false"
	require.Equal(t, "false", body)
}

func TestBlock(t *testing.T) {
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()

	var resultBlock ctypes.ResultBlock

	res, body := Request(t, port, "GET", "/blocks/latest", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultBlock)
	require.Nil(t, err, "Couldn't parse block")

	require.NotEqual(t, ctypes.ResultBlock{}, resultBlock)

	// --

	res, body = Request(t, port, "GET", "/blocks/1", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = wire.Cdc.UnmarshalJSON([]byte(body), &resultBlock)
	require.Nil(t, err, "Couldn't parse block")

	require.NotEqual(t, ctypes.ResultBlock{}, resultBlock)

	// --

	res, body = Request(t, port, "GET", "/blocks/1000000000", nil)
	require.Equal(t, http.StatusNotFound, res.StatusCode, body)
}

func TestValidators(t *testing.T) {
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()

	var resultVals rpc.ResultValidatorsOutput

	res, body := Request(t, port, "GET", "/validatorsets/latest", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultVals)
	require.Nil(t, err, "Couldn't parse validatorset")

	require.NotEqual(t, rpc.ResultValidatorsOutput{}, resultVals)

	require.Contains(t, resultVals.Validators[0].Address.String(), "cosmosvaladdr")
	require.Contains(t, resultVals.Validators[0].PubKey, "cosmosvalpub")

	// --

	res, body = Request(t, port, "GET", "/validatorsets/1", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultVals)
	require.Nil(t, err, "Couldn't parse validatorset")

	require.NotEqual(t, rpc.ResultValidatorsOutput{}, resultVals)

	// --

	res, body = Request(t, port, "GET", "/validatorsets/1000000000", nil)
	require.Equal(t, http.StatusNotFound, res.StatusCode, body)
}

func TestCoinSend(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	bz, err := hex.DecodeString("8FA6AB57AD6870F6B5B2E57735F38F2F30E73CB6")
	require.NoError(t, err)
	someFakeAddr := sdk.AccAddress(bz)

	// query empty
	res, body := Request(t, port, "GET", fmt.Sprintf("/accounts/%s", someFakeAddr), nil)
	require.Equal(t, http.StatusNoContent, res.StatusCode, body)

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create TX
	receiveAddr, resultTx := doSend(t, port, seed, name, password, addr)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	mycoins := coins[0]

	require.Equal(t, "steak", mycoins.Denom)
	require.Equal(t, initialBalance[0].Amount.SubRaw(1), mycoins.Amount)

	// query receiver
	acc = getAccount(t, port, receiveAddr)
	coins = acc.GetCoins()
	mycoins = coins[0]

	require.Equal(t, "steak", mycoins.Denom)
	require.Equal(t, int64(1), mycoins.Amount.Int64())
}

func TestIBCTransfer(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create TX
	resultTx := doIBCTransfer(t, port, seed, name, password, addr)

	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	mycoins := coins[0]

	require.Equal(t, "steak", mycoins.Denom)
	require.Equal(t, initialBalance[0].Amount.SubRaw(1), mycoins.Amount)

	// TODO: query ibc egress packet state
}

func TestTxs(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// query wrong
	res, body := Request(t, port, "GET", "/txs", nil)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, body)

	// query empty
	res, body = Request(t, port, "GET", fmt.Sprintf("/txs?tag=sender_bech32='%s'", "cosmosaccaddr1jawd35d9aq4u76sr3fjalmcqc8hqygs9gtnmv3"), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Equal(t, "[]", body)

	// create TX
	receiveAddr, resultTx := doSend(t, port, seed, name, password, addr)

	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx is findable
	res, body = Request(t, port, "GET", fmt.Sprintf("/txs/%s", resultTx.Hash), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var indexedTxs []tx.Info

	// check if tx is queryable
	res, body = Request(t, port, "GET", fmt.Sprintf("/txs?tag=tx.hash='%s'", resultTx.Hash), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NotEqual(t, "[]", body)

	err := cdc.UnmarshalJSON([]byte(body), &indexedTxs)
	require.NoError(t, err)
	require.Equal(t, 1, len(indexedTxs))

	// XXX should this move into some other testfile for txs in general?
	// test if created TX hash is the correct hash
	require.Equal(t, resultTx.Hash, indexedTxs[0].Hash)

	// query sender
	// also tests url decoding
	res, body = Request(t, port, "GET", fmt.Sprintf("/txs?tag=sender_bech32=%%27%s%%27", addr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &indexedTxs)
	require.NoError(t, err)
	require.Equal(t, 1, len(indexedTxs), "%v", indexedTxs) // there are 2 txs created with doSend
	require.Equal(t, resultTx.Height, indexedTxs[0].Height)

	// query recipient
	res, body = Request(t, port, "GET", fmt.Sprintf("/txs?tag=recipient_bech32='%s'", receiveAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &indexedTxs)
	require.NoError(t, err)
	require.Equal(t, 1, len(indexedTxs))
	require.Equal(t, resultTx.Height, indexedTxs[0].Height)
}

func TestValidatorsQuery(t *testing.T) {
	cleanup, pks, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()
	require.Equal(t, 1, len(pks))

	validators := getValidators(t, port)
	require.Equal(t, len(validators), 1)

	// make sure all the validators were found (order unknown because sorted by owner addr)
	foundVal := false
	pkBech := sdk.MustBech32ifyValPub(pks[0])
	if validators[0].PubKey == pkBech {
		foundVal = true
	}
	require.True(t, foundVal, "pkBech %v, owner %v", pkBech, validators[0].Owner)
}

func TestValidatorQuery(t *testing.T) {
	cleanup, pks, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()
	require.Equal(t, 1, len(pks))

	validator1Owner := sdk.AccAddress(pks[0].Address())

	validator := getValidator(t, port, validator1Owner)
	bech32ValAddress, err := sdk.Bech32ifyValPub(pks[0])
	require.NoError(t, err)
	assert.Equal(t, validator.PubKey, bech32ValAddress, "The returned validator does not hold the correct data")
}

func TestBonding(t *testing.T) {
	name, password, denom := "test", "1234567890", "steak"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, pks, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	validator1Owner := sdk.AccAddress(pks[0].Address())

	// create bond TX
	resultTx := doDelegate(t, port, seed, name, password, addr, validator1Owner)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc := getAccount(t, port, addr)
	coins := acc.GetCoins()

	require.Equal(t, int64(40), coins.AmountOf(denom).Int64())

	// query validator
	bond := getDelegation(t, port, addr, validator1Owner)
	require.Equal(t, "60.0000000000", bond.Shares)

	//////////////////////
	// testing unbonding

	// create unbond TX
	resultTx = doBeginUnbonding(t, port, seed, name, password, addr, validator1Owner)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query validator
	bond = getDelegation(t, port, addr, validator1Owner)
	require.Equal(t, "30.0000000000", bond.Shares)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// should the sender should have not received any coins as the unbonding has only just begun
	// query sender
	acc = getAccount(t, port, addr)
	coins = acc.GetCoins()
	require.Equal(t, int64(40), coins.AmountOf("steak").Int64())

	// query unbonding delegation
	validatorAddr := sdk.AccAddress(pks[0].Address())
	unbondings := getUndelegations(t, port, addr, validatorAddr)
	assert.Len(t, unbondings, 1, "Unbondings holds all unbonding-delegations")
	assert.Equal(t, "30", unbondings[0].Balance.Amount.String())

	// query summary
	summary := getDelegationSummary(t, port, addr)

	assert.Len(t, summary.Delegations, 1, "Delegation summary holds all delegations")
	assert.Equal(t, "30.0000000000", summary.Delegations[0].Shares)
	assert.Len(t, summary.UnbondingDelegations, 1, "Delegation summary holds all unbonding-delegations")
	assert.Equal(t, "30", summary.UnbondingDelegations[0].Balance.Amount.String())

	// TODO add redelegation, need more complex capabilities such to mock context and
	// TODO check summary for redelegation
	// assert.Len(t, summary.Redelegations, 1, "Delegation summary holds all redelegations")

	// query txs
	txs := getBondingTxs(t, port, addr, "")
	assert.Len(t, txs, 2, "All Txs found")

	txs = getBondingTxs(t, port, addr, "bond")
	assert.Len(t, txs, 1, "All bonding txs found")

	txs = getBondingTxs(t, port, addr, "unbond")
	assert.Len(t, txs, 1, "All unbonding txs found")
}

func TestSubmitProposal(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name, password, addr)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID int64
	cdc.UnmarshalBinaryBare(resultTx.DeliverTx.GetData(), &proposalID)

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())
}

func TestDeposit(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name, password, addr)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID int64
	cdc.UnmarshalBinaryBare(resultTx.DeliverTx.GetData(), &proposalID)

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	// create SubmitProposal TX
	resultTx = doDeposit(t, port, seed, name, password, addr, proposalID)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query proposal
	proposal = getProposal(t, port, proposalID)
	require.True(t, proposal.GetTotalDeposit().IsEqual(sdk.Coins{sdk.NewInt64Coin("steak", 10)}))

	// query deposit
	deposit := getDeposit(t, port, proposalID, addr)
	require.True(t, deposit.Amount.IsEqual(sdk.Coins{sdk.NewInt64Coin("steak", 10)}))
}

func TestVote(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name, password, addr)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID int64
	cdc.UnmarshalBinaryBare(resultTx.DeliverTx.GetData(), &proposalID)

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	// create SubmitProposal TX
	resultTx = doDeposit(t, port, seed, name, password, addr, proposalID)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query proposal
	proposal = getProposal(t, port, proposalID)
	require.Equal(t, gov.StatusVotingPeriod, proposal.GetStatus())

	// create SubmitProposal TX
	resultTx = doVote(t, port, seed, name, password, addr, proposalID)
	tests.WaitForHeight(resultTx.Height+1, port)

	vote := getVote(t, port, proposalID, addr)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, gov.OptionYes, vote.Option)
}

func TestUnrevoke(t *testing.T) {
	_, password := "test", "1234567890"
	addr, _ := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, pks, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// XXX: any less than this and it fails
	tests.WaitForHeight(3, port)
	pkString, _ := sdk.Bech32ifyValPub(pks[0])
	signingInfo := getSigningInfo(t, port, pkString)
	tests.WaitForHeight(4, port)
	require.Equal(t, true, signingInfo.IndexOffset > 0)
	require.Equal(t, time.Unix(0, 0).UTC(), signingInfo.JailedUntil)
	require.Equal(t, true, signingInfo.SignedBlocksCounter > 0)
}

func TestProposalsQuery(t *testing.T) {
	name, password1 := "test", "1234567890"
	name2, password2 := "test2", "1234567890"
	addr, seed := CreateAddr(t, "test", password1, GetKeyBase(t))
	addr2, seed2 := CreateAddr(t, "test2", password2, GetKeyBase(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr, addr2})
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

	// Only proposals #1 should be in Deposit Period
	proposals := getProposalsFilterStatus(t, port, gov.StatusDepositPeriod)
	require.Len(t, proposals, 1)
	require.Equal(t, proposalID1, proposals[0].GetProposalID())
	// Only proposals #2 and #3 should be in Voting Period
	proposals = getProposalsFilterStatus(t, port, gov.StatusVotingPeriod)
	require.Len(t, proposals, 2)
	require.Equal(t, proposalID2, proposals[0].GetProposalID())
	require.Equal(t, proposalID3, proposals[1].GetProposalID())

	// Addr1 votes on proposals #2 & #3
	resultTx = doVote(t, port, seed, name, password1, addr, proposalID2)
	tests.WaitForHeight(resultTx.Height+1, port)
	resultTx = doVote(t, port, seed, name, password1, addr, proposalID3)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 votes on proposal #3
	resultTx = doVote(t, port, seed2, name2, password2, addr2, proposalID3)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Test query all proposals
	proposals = getProposalsAll(t, port)
	require.Equal(t, proposalID1, (proposals[0]).GetProposalID())
	require.Equal(t, proposalID2, (proposals[1]).GetProposalID())
	require.Equal(t, proposalID3, (proposals[2]).GetProposalID())

	// Test query deposited by addr1
	proposals = getProposalsFilterDepositer(t, port, addr)
	require.Equal(t, proposalID1, (proposals[0]).GetProposalID())

	// Test query deposited by addr2
	proposals = getProposalsFilterDepositer(t, port, addr2)
	require.Equal(t, proposalID2, (proposals[0]).GetProposalID())
	require.Equal(t, proposalID3, (proposals[1]).GetProposalID())

	// Test query voted by addr1
	proposals = getProposalsFilterVoter(t, port, addr)
	require.Equal(t, proposalID2, (proposals[0]).GetProposalID())
	require.Equal(t, proposalID3, (proposals[1]).GetProposalID())

	// Test query voted by addr2
	proposals = getProposalsFilterVoter(t, port, addr2)
	require.Equal(t, proposalID3, (proposals[0]).GetProposalID())

	// Test query voted and deposited by addr1
	proposals = getProposalsFilterVoterDepositer(t, port, addr, addr)
	require.Equal(t, proposalID2, (proposals[0]).GetProposalID())

	// Test query votes on Proposal 2
	votes := getVotes(t, port, proposalID2)
	require.Len(t, votes, 1)
	require.Equal(t, addr, votes[0].Voter)

	// Test query votes on Proposal 3
	votes = getVotes(t, port, proposalID3)
	require.Len(t, votes, 2)
	require.True(t, addr.String() == votes[0].Voter.String() || addr.String() == votes[1].Voter.String())
	require.True(t, addr2.String() == votes[0].Voter.String() || addr2.String() == votes[1].Voter.String())
}

//_____________________________________________________________________________
// get the account to get the sequence
func getAccount(t *testing.T, port string, addr sdk.AccAddress) auth.Account {
	res, body := Request(t, port, "GET", fmt.Sprintf("/accounts/%s", addr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var acc auth.Account
	err := cdc.UnmarshalJSON([]byte(body), &acc)
	require.Nil(t, err)
	return acc
}

func doSend(t *testing.T, port, seed, name, password string, addr sdk.AccAddress) (receiveAddr sdk.AccAddress, resultTx ctypes.ResultBroadcastTxCommit) {

	// create receive address
	kb := client.MockKeyBase()
	receiveInfo, _, err := kb.CreateMnemonic("receive_address", cryptoKeys.English, "1234567890", cryptoKeys.SigningAlgo("secp256k1"))
	require.Nil(t, err)
	receiveAddr = sdk.AccAddress(receiveInfo.GetPubKey().Address())

	acc := getAccount(t, port, addr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	// send
	coinbz, err := cdc.MarshalJSON(sdk.NewInt64Coin("steak", 1))
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
	res, body := Request(t, port, "POST", fmt.Sprintf("/accounts/%s/send", receiveAddr), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	return receiveAddr, resultTx
}

func doIBCTransfer(t *testing.T, port, seed, name, password string, addr sdk.AccAddress) (resultTx ctypes.ResultBroadcastTxCommit) {
	// create receive address
	kb := client.MockKeyBase()
	receiveInfo, _, err := kb.CreateMnemonic("receive_address", cryptoKeys.English, "1234567890", cryptoKeys.SigningAlgo("secp256k1"))
	require.Nil(t, err)
	receiveAddr := sdk.AccAddress(receiveInfo.GetPubKey().Address())

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
		"src_chain_id": "%s",
		"amount":[
			{
				"denom": "%s",
				"amount": "1"
			}
		]
	}`, name, password, accnum, sequence, chainID, "steak"))

	res, body := Request(t, port, "POST", fmt.Sprintf("/ibc/testchain/%s/send", receiveAddr), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	return resultTx
}

func getSigningInfo(t *testing.T, port string, validatorPubKey string) slashing.ValidatorSigningInfo {
	res, body := Request(t, port, "GET", fmt.Sprintf("/slashing/signing_info/%s", validatorPubKey), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var signingInfo slashing.ValidatorSigningInfo
	err := cdc.UnmarshalJSON([]byte(body), &signingInfo)
	require.Nil(t, err)
	return signingInfo
}

// ============= Stake Module ================

func getDelegation(t *testing.T, port string, delegatorAddr, validatorAddr sdk.AccAddress) rest.DelegationWithoutRat {

	// get the account to get the sequence
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/delegations/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var bond rest.DelegationWithoutRat
	err := cdc.UnmarshalJSON([]byte(body), &bond)
	require.Nil(t, err)
	return bond
}

func getUndelegations(t *testing.T, port string, delegatorAddr, validatorAddr sdk.AccAddress) []stake.UnbondingDelegation {

	// get the account to get the sequence
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/unbonding_delegations/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var unbondings []stake.UnbondingDelegation
	err := cdc.UnmarshalJSON([]byte(body), &unbondings)
	require.Nil(t, err)
	return unbondings
}

func getDelegationSummary(t *testing.T, port string, delegatorAddr sdk.AccAddress) rest.DelegationSummary {

	// get the account to get the sequence
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var summary rest.DelegationSummary
	err := cdc.UnmarshalJSON([]byte(body), &summary)
	require.Nil(t, err)
	return summary
}

func getBondingTxs(t *testing.T, port string, delegatorAddr sdk.AccAddress, query string) []tx.Info {

	// get the account to get the sequence
	var res *http.Response
	var body string
	if len(query) > 0 {
		res, body = Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/txs?type=%s", delegatorAddr, query), nil)
	} else {
		res, body = Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/txs", delegatorAddr), nil)
	}
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var txs []tx.Info
	err := cdc.UnmarshalJSON([]byte(body), &txs)
	require.Nil(t, err)
	return txs
}

func doDelegate(t *testing.T, port, seed, name, password string, delegatorAddr, validatorAddr sdk.AccAddress) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, port, delegatorAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

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
				"delegation": { "denom": "%s", "amount": "60" }
			}
		],
		"begin_unbondings": [],
		"complete_unbondings": [],
		"begin_redelegates": [],
		"complete_redelegates": []
	}`, name, password, accnum, sequence, chainID, delegatorAddr, validatorAddr, "steak"))
	res, body := Request(t, port, "POST", fmt.Sprintf("/stake/delegators/%s/delegations", delegatorAddr), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func doBeginUnbonding(t *testing.T, port, seed, name, password string,
	delegatorAddr, validatorAddr sdk.AccAddress) (resultTx ctypes.ResultBroadcastTxCommit) {

	// get the account to get the sequence
	acc := getAccount(t, port, delegatorAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)

	// send
	jsonStr := []byte(fmt.Sprintf(`{
		"name": "%s",
		"password": "%s",
		"account_number": "%d",
		"sequence": "%d",
		"gas": "20000",
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
	}`, name, password, accnum, sequence, chainID, delegatorAddr, validatorAddr))
	res, body := Request(t, port, "POST", fmt.Sprintf("/stake/delegators/%s/delegations", delegatorAddr), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func doBeginRedelegation(t *testing.T, port, seed, name, password string,
	delegatorAddr, validatorSrcAddr, validatorDstAddr sdk.AccAddress) (resultTx ctypes.ResultBroadcastTxCommit) {

	// get the account to get the sequence
	acc := getAccount(t, port, delegatorAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

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
	}`, name, password, accnum, sequence, chainID, delegatorAddr, validatorSrcAddr, validatorDstAddr))
	res, body := Request(t, port, "POST", fmt.Sprintf("/stake/delegators/%s/delegations", delegatorAddr), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func getValidators(t *testing.T, port string) []stake.BechValidator {
	// get the account to get the sequence
	res, body := Request(t, port, "GET", "/stake/validators", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var validators []stake.BechValidator
	err := cdc.UnmarshalJSON([]byte(body), &validators)
	require.Nil(t, err)
	return validators
}

func getValidator(t *testing.T, port string, validatorAddr sdk.AccAddress) stake.BechValidator {
	// get the account to get the sequence
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/validators/%s", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var validator stake.BechValidator
	err := cdc.UnmarshalJSON([]byte(body), &validator)
	require.Nil(t, err)
	return validator
}

// ============= Governance Module ================

func getProposal(t *testing.T, port string, proposalID int64) gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var proposal gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposal)
	require.Nil(t, err)
	return proposal
}

func getDeposit(t *testing.T, port string, proposalID int64, depositerAddr sdk.AccAddress) gov.Deposit {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/deposits/%s", proposalID, depositerAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var deposit gov.Deposit
	err := cdc.UnmarshalJSON([]byte(body), &deposit)
	require.Nil(t, err)
	return deposit
}

func getVote(t *testing.T, port string, proposalID int64, voterAddr sdk.AccAddress) gov.Vote {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/votes/%s", proposalID, voterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var vote gov.Vote
	err := cdc.UnmarshalJSON([]byte(body), &vote)
	require.Nil(t, err)
	return vote
}

func getVotes(t *testing.T, port string, proposalID int64) []gov.Vote {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/votes", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var votes []gov.Vote
	err := cdc.UnmarshalJSON([]byte(body), &votes)
	require.Nil(t, err)
	return votes
}

func getProposalsAll(t *testing.T, port string) []gov.Proposal {
	res, body := Request(t, port, "GET", "/gov/proposals", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

func getProposalsFilterDepositer(t *testing.T, port string, depositerAddr sdk.AccAddress) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?depositer=%s", depositerAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

func getProposalsFilterVoter(t *testing.T, port string, voterAddr sdk.AccAddress) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?voter=%s", voterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

func getProposalsFilterVoterDepositer(t *testing.T, port string, voterAddr, depositerAddr sdk.AccAddress) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?depositer=%s&voter=%s", depositerAddr, voterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

func getProposalsFilterStatus(t *testing.T, port string, status gov.ProposalStatus) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?status=%s", status), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

func doSubmitProposal(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)

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
	}`, proposerAddr, name, password, chainID, accnum, sequence))
	res, body := Request(t, port, "POST", "/gov/proposals", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

func doDeposit(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress, proposalID int64) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)

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
	}`, proposerAddr, name, password, chainID, accnum, sequence))
	res, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/deposits", proposalID), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

func doVote(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress, proposalID int64) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)

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
	}`, proposerAddr, name, password, chainID, accnum, sequence))
	res, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/votes", proposalID), jsonStr)
	fmt.Println(res)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}
