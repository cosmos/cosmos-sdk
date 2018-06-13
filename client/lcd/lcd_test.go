package lcd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	cryptoKeys "github.com/tendermint/go-crypto/keys"
	p2p "github.com/tendermint/tendermint/p2p"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	client "github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	rpc "github.com/cosmos/cosmos-sdk/client/rpc"
	tests "github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"
	stakerest "github.com/cosmos/cosmos-sdk/x/stake/client/rest"
)

func TestKeys(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 2, []sdk.Address{addr})
	defer cleanup()

	// get seed
	res, body := Request(t, port, "GET", "/keys/seed", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	newSeed := body
	reg, err := regexp.Compile(`([a-z]+ ){12}`)
	require.Nil(t, err)
	match := reg.MatchString(seed)
	assert.True(t, match, "Returned seed has wrong format", seed)

	newName := "test_newname"
	newPassword := "0987654321"

	// add key
	var jsonStr = []byte(fmt.Sprintf(`{"name":"test_fail", "password":"%s"}`, password))
	res, body = Request(t, port, "POST", "/keys", jsonStr)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Account creation should require a seed")

	jsonStr = []byte(fmt.Sprintf(`{"name":"%s", "password":"%s", "seed": "%s"}`, newName, newPassword, newSeed))
	res, body = Request(t, port, "POST", "/keys", jsonStr)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	addr2 := body
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

	err = json.Unmarshal([]byte(body), &resultBlock)
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
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestCoinSend(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 2, []sdk.Address{addr})
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

	// check if tx was commited
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	mycoins := coins[0]
	assert.Equal(t, "steak", mycoins.Denom)
	assert.Equal(t, initialBalance[0].Amount-1, mycoins.Amount)

	// query receiver
	acc = getAccount(t, port, receiveAddr)
	coins = acc.GetCoins()
	mycoins = coins[0]
	assert.Equal(t, "steak", mycoins.Denom)
	assert.Equal(t, int64(1), mycoins.Amount)
}

func TestIBCTransfer(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 2, []sdk.Address{addr})
	defer cleanup()

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create TX
	resultTx := doIBCTransfer(t, port, seed, name, password, addr)

	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was commited
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	mycoins := coins[0]
	assert.Equal(t, "steak", mycoins.Denom)
	assert.Equal(t, initialBalance[0].Amount-1, mycoins.Amount)

	// TODO: query ibc egress packet state
}

func TestTxs(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKB(t))
	cleanup, _, port := InitializeTestLCD(t, 2, []sdk.Address{addr})
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

	// query sender
	addrBech := sdk.MustBech32ifyAcc(addr)
	res, body = Request(t, port, "GET", fmt.Sprintf("/txs?tag=sender_bech32='%s'", addrBech), nil)
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
	require.Equal(t, 2, len(pks))
	defer cleanup()

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
	cleanup, pks, port := InitializeTestLCD(t, 2, []sdk.Address{addr})
	defer cleanup()

	validator1Owner := pks[0].Address()

	// create bond TX
	resultTx := doBond(t, port, seed, name, password, addr, validator1Owner)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was commited
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc := getAccount(t, port, addr)
	coins := acc.GetCoins()
	assert.Equal(t, int64(40), coins.AmountOf(denom))

	// query validator
	bond := getDelegation(t, port, addr, validator1Owner)
	assert.Equal(t, "60/1", bond.Shares.String())

	//////////////////////
	// testing unbonding

	// create unbond TX
	resultTx = doUnbond(t, port, seed, name, password, addr, validator1Owner)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query validator
	bond = getDelegation(t, port, addr, validator1Owner)
	assert.Equal(t, "30/1", bond.Shares.String())

	// check if tx was commited
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// TODO fix shares fn in staking
	// query sender
	//acc := getAccount(t, sendAddr)
	//coins := acc.GetCoins()
	//assert.Equal(t, int64(98), coins.AmountOf(coinDenom))

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
	receiveInfo, _, err := kb.Create("receive_address", "1234567890", cryptoKeys.CryptoAlgo("ed25519"))
	require.Nil(t, err)
	receiveAddr = receiveInfo.PubKey.Address()
	receiveAddrBech := sdk.MustBech32ifyAcc(receiveAddr)

	acc := getAccount(t, port, addr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	// send
	jsonStr := []byte(fmt.Sprintf(`{
		"name":"%s", 
		"password":"%s",
		"account_number":%d, 
		"sequence":%d, 
		"gas": 10000,
		"amount":[
			{ 
				"denom": "%s", 
				"amount": 1 
			}
		] 
	}`, name, password, accnum, sequence, "steak"))
	res, body := Request(t, port, "POST", "/accounts/"+receiveAddrBech+"/send", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	return receiveAddr, resultTx
}

func doIBCTransfer(t *testing.T, port, seed, name, password string, addr sdk.Address) (resultTx ctypes.ResultBroadcastTxCommit) {
	// create receive address
	kb := client.MockKeyBase()
	receiveInfo, _, err := kb.Create("receive_address", "1234567890", cryptoKeys.CryptoAlgo("ed25519"))
	require.Nil(t, err)
	receiveAddr := receiveInfo.PubKey.Address()
	receiveAddrBech := sdk.MustBech32ifyAcc(receiveAddr)

	// get the account to get the sequence
	acc := getAccount(t, port, addr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	// send
	jsonStr := []byte(fmt.Sprintf(`{ 
		"name":"%s", 
		"password": "%s", 
		"account_number":%d,
		"sequence": %d, 
		"gas": 100000,
		"amount":[
			{ 
				"denom": "%s", 
				"amount": 1 
			}
		] 
	}`, name, password, accnum, sequence, "steak"))
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
	res, body := Request(t, port, "GET", "/stake/"+delegatorAddrBech+"/bonding_status/"+validatorAddrBech, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var bond stake.Delegation
	err := cdc.UnmarshalJSON([]byte(body), &bond)
	require.Nil(t, err)
	return bond
}

func doBond(t *testing.T, port, seed, name, password string, delegatorAddr, validatorAddr sdk.Address) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, port, delegatorAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	delegatorAddrBech := sdk.MustBech32ifyAcc(delegatorAddr)
	validatorAddrBech := sdk.MustBech32ifyVal(validatorAddr)

	// send
	jsonStr := []byte(fmt.Sprintf(`{
		"name": "%s",
		"password": "%s",
		"account_number": %d,
		"sequence": %d,
		"gas": 10000,
		"delegate": [
			{
				"delegator_addr": "%s",
				"validator_addr": "%s",
				"bond": { "denom": "%s", "amount": 60 }
			}
		],
		"unbond": []
	}`, name, password, accnum, sequence, delegatorAddrBech, validatorAddrBech, "steak"))
	res, body := Request(t, port, "POST", "/stake/delegations", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func doUnbond(t *testing.T, port, seed, name, password string, delegatorAddr, validatorAddr sdk.Address) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, port, delegatorAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	delegatorAddrBech := sdk.MustBech32ifyAcc(delegatorAddr)
	validatorAddrBech := sdk.MustBech32ifyVal(validatorAddr)

	// send
	jsonStr := []byte(fmt.Sprintf(`{
		"name": "%s",
		"password": "%s",
		"account_number": %d,
		"sequence": %d,
		"gas": 10000,
		"delegate": [],
		"unbond": [
			{
				"delegator_addr": "%s",
				"validator_addr": "%s",
				"shares": "30"
			}
		]
	}`, name, password, accnum, sequence, delegatorAddrBech, validatorAddrBech))
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
