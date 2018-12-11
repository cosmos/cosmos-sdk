package lcd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	p2p "github.com/tendermint/tendermint/p2p"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	client "github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptoKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
	tests "github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	version "github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
)

func init() {
	mintkey.BcryptSecurityParameter = 1
	version.Version = os.Getenv("VERSION")
}

func TestKeys(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// get seed
	// TODO Do we really need this endpoint?
	res, body := Request(t, port, "GET", "/keys/seed", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	reg, err := regexp.Compile(`([a-z]+ ){12}`)
	require.Nil(t, err)
	match := reg.MatchString(seed)
	require.True(t, match, "Returned seed has wrong format", seed)

	// recover key
	recoverName := "test_recovername"
	recoverPassword := "1234567890"
	doRecoverKey(t, port, recoverName, recoverPassword, seed)

	newName := "test_newname"
	newPassword := "0987654321"
	// add key
	jsonStr := []byte(fmt.Sprintf(`{"name":"%s", "password":"%s", "seed":"%s"}`, newName, newPassword, seed))
	res, body = Request(t, port, "POST", "/keys", jsonStr)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.KeyOutput
	err = codec.Cdc.UnmarshalJSON([]byte(body), &resp)
	require.Nil(t, err, body)

	addr2Bech32 := resp.Address
	_, err = sdk.AccAddressFromBech32(addr2Bech32)
	require.NoError(t, err, "Failed to return a correct bech32 address")

	// test if created account is the correct account
	expectedInfo, _ := GetKeyBase(t).CreateKey(newName, seed, newPassword)
	expectedAccount := sdk.AccAddress(expectedInfo.GetPubKey().Address().Bytes())
	require.Equal(t, expectedAccount.String(), addr2Bech32)

	// existing keys
	res, body = Request(t, port, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m [3]keys.KeyOutput
	err = cdc.UnmarshalJSON([]byte(body), &m)
	require.Nil(t, err)

	addrBech32 := addr.String()

	require.Equal(t, name, m[0].Name, "Did not serve keys name correctly")
	require.Equal(t, addrBech32, m[0].Address, "Did not serve keys Address correctly")
	require.Equal(t, newName, m[1].Name, "Did not serve keys name correctly")
	require.Equal(t, addr2Bech32, m[1].Address, "Did not serve keys Address correctly")

	// select key
	keyEndpoint := fmt.Sprintf("/keys/%s", newName)
	res, body = Request(t, port, "GET", keyEndpoint, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m2 keys.KeyOutput
	err = cdc.UnmarshalJSON([]byte(body), &m2)
	require.Nil(t, err)

	require.Equal(t, newName, m2.Name, "Did not serve keys name correctly")
	require.Equal(t, addr2Bech32, m2.Address, "Did not serve keys Address correctly")

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
	// skip the test if the VERSION environment variable has not been set
	if version.Version == "" {
		t.SkipNow()
	}

	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()

	// node info
	res, body := Request(t, port, "GET", "/version", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	reg, err := regexp.Compile(`\d+\.\d+\.\d+.*`)
	require.Nil(t, err)
	match := reg.MatchString(body)
	require.True(t, match, body, fmt.Sprintf("%s", body))

	// node info
	res, body = Request(t, port, "GET", "/node_version", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	reg, err = regexp.Compile(`\d+\.\d+\.\d+.*`)
	require.Nil(t, err)
	match = reg.MatchString(body)
	require.True(t, match, body)
}

func TestNodeStatus(t *testing.T) {
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()

	// node info
	res, body := Request(t, port, "GET", "/node_info", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var nodeInfo p2p.DefaultNodeInfo
	err := cdc.UnmarshalJSON([]byte(body), &nodeInfo)
	require.Nil(t, err, "Couldn't parse node info")

	require.NotEqual(t, p2p.DefaultNodeInfo{}, nodeInfo, "res: %v", res)

	// syncing
	res, body = Request(t, port, "GET", "/syncing", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// we expect that there is no other node running so the syncing state is "false"
	require.Equal(t, "false", body)
}

func TestBlock(t *testing.T) {
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()

	var resultBlock ctypes.ResultBlock

	res, body := Request(t, port, "GET", "/blocks/latest", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultBlock)
	require.Nil(t, err, "Couldn't parse block")

	require.NotEqual(t, ctypes.ResultBlock{}, resultBlock)

	// --

	res, body = Request(t, port, "GET", "/blocks/2", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = codec.Cdc.UnmarshalJSON([]byte(body), &resultBlock)
	require.Nil(t, err, "Couldn't parse block")

	require.NotEqual(t, ctypes.ResultBlock{}, resultBlock)

	// --

	res, body = Request(t, port, "GET", "/blocks/1000000000", nil)
	require.Equal(t, http.StatusNotFound, res.StatusCode, body)
}

func TestValidators(t *testing.T) {
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()

	var resultVals rpc.ResultValidatorsOutput

	res, body := Request(t, port, "GET", "/validatorsets/latest", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultVals)
	require.Nil(t, err, "Couldn't parse validatorset")

	require.NotEqual(t, rpc.ResultValidatorsOutput{}, resultVals)

	require.Contains(t, resultVals.Validators[0].Address.String(), "cosmosvaloper")
	require.Contains(t, resultVals.Validators[0].PubKey, "cosmosvalconspub")

	// --

	res, body = Request(t, port, "GET", "/validatorsets/2", nil)
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
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	bz, err := hex.DecodeString("8FA6AB57AD6870F6B5B2E57735F38F2F30E73CB6")
	require.NoError(t, err)
	someFakeAddr := sdk.AccAddress(bz)

	// query empty
	res, body := Request(t, port, "GET", fmt.Sprintf("/auth/accounts/%s", someFakeAddr), nil)
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

	require.Equal(t, stakeTypes.DefaultBondDenom, mycoins.Denom)
	require.Equal(t, initialBalance[0].Amount.SubRaw(1), mycoins.Amount)

	// query receiver
	acc = getAccount(t, port, receiveAddr)
	coins = acc.GetCoins()
	mycoins = coins[0]

	require.Equal(t, stakeTypes.DefaultBondDenom, mycoins.Denom)
	require.Equal(t, int64(1), mycoins.Amount.Int64())

	// test failure with too little gas
	res, body, _ = doSendWithGas(t, port, seed, name, password, addr, "100", 0, false, false)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode, body)

	// test failure with negative gas
	res, body, _ = doSendWithGas(t, port, seed, name, password, addr, "-200", 0, false, false)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, body)

	// test failure with 0 gas
	res, body, _ = doSendWithGas(t, port, seed, name, password, addr, "0", 0, false, false)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode, body)

	// test failure with wrong adjustment
	res, body, _ = doSendWithGas(t, port, seed, name, password, addr, "simulate", 0.1, false, false)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode, body)

	// run simulation and test success with estimated gas
	res, body, _ = doSendWithGas(t, port, seed, name, password, addr, "", 0, true, false)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var responseBody struct {
		GasEstimate int64 `json:"gas_estimate"`
	}
	require.Nil(t, json.Unmarshal([]byte(body), &responseBody))
	res, body, _ = doSendWithGas(t, port, seed, name, password, addr, fmt.Sprintf("%v", responseBody.GasEstimate), 0, false, false)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
}

func DisabledTestIBCTransfer(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
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

	require.Equal(t, stakeTypes.DefaultBondDenom, mycoins.Denom)
	require.Equal(t, initialBalance[0].Amount.SubRaw(1), mycoins.Amount)

	// TODO: query ibc egress packet state
}

func TestCoinSendGenerateSignAndBroadcast(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()
	acc := getAccount(t, port, addr)

	// generate TX
	res, body, _ := doSendWithGas(t, port, seed, name, password, addr, "simulate", 0, false, true)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var msg auth.StdTx
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &msg))
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, msg.Msgs[0].Route(), "bank")
	require.Equal(t, msg.Msgs[0].GetSigners(), []sdk.AccAddress{addr})
	require.Equal(t, 0, len(msg.Signatures))
	gasEstimate := msg.Fee.Gas

	// sign tx
	var signedMsg auth.StdTx
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	payload := authrest.SignBody{
		Tx:               msg,
		LocalAccountName: name,
		Password:         password,
		ChainID:          viper.GetString(client.FlagChainID),
		AccountNumber:    accnum,
		Sequence:         sequence,
	}
	json, err := cdc.MarshalJSON(payload)
	require.Nil(t, err)
	res, body = Request(t, port, "POST", "/tx/sign", json)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &signedMsg))
	require.Equal(t, len(msg.Msgs), len(signedMsg.Msgs))
	require.Equal(t, msg.Msgs[0].Type(), signedMsg.Msgs[0].Type())
	require.Equal(t, msg.Msgs[0].GetSigners(), signedMsg.Msgs[0].GetSigners())
	require.Equal(t, 1, len(signedMsg.Signatures))

	// broadcast tx
	broadcastPayload := struct {
		Tx     auth.StdTx `json:"tx"`
		Return string     `json:"return"`
	}{Tx: signedMsg, Return: "block"}
	json, err = cdc.MarshalJSON(broadcastPayload)
	require.Nil(t, err)
	res, body = Request(t, port, "POST", "/tx/broadcast", json)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// check if tx was committed
	var resultTx ctypes.ResultBroadcastTxCommit
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &resultTx))
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)
	require.Equal(t, gasEstimate, uint64(resultTx.DeliverTx.GasWanted))
	require.Equal(t, gasEstimate, uint64(resultTx.DeliverTx.GasUsed))
}

func TestTxs(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	var emptyTxs []tx.Info
	txs := getTransactions(t, port)
	require.Equal(t, emptyTxs, txs)

	// query empty
	txs = getTransactions(t, port, fmt.Sprintf("sender=%s", addr.String()))
	require.Equal(t, emptyTxs, txs)

	// also tests url decoding
	txs = getTransactions(t, port, fmt.Sprintf("sender=%s", addr.String()))
	require.Equal(t, emptyTxs, txs)

	txs = getTransactions(t, port, fmt.Sprintf("action=submit%%20proposal&proposer=%s", addr.String()))
	require.Equal(t, emptyTxs, txs)

	// create tx
	receiveAddr, resultTx := doSend(t, port, seed, name, password, addr)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx is queryable
	txs = getTransactions(t, port, fmt.Sprintf("tx.hash=%s", resultTx.Hash))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Hash, txs[0].Hash)

	// query sender
	txs = getTransactions(t, port, fmt.Sprintf("sender=%s", addr.String()))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	// query recipient
	txs = getTransactions(t, port, fmt.Sprintf("recipient=%s", receiveAddr.String()))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)
}

func TestPoolParamsQuery(t *testing.T) {
	_, password := "test", "1234567890"
	addr, _ := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	defaultParams := stake.DefaultParams()

	res, body := Request(t, port, "GET", "/stake/parameters", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var params stake.Params
	err := cdc.UnmarshalJSON([]byte(body), &params)
	require.Nil(t, err)
	require.True(t, defaultParams.Equal(params))

	res, body = Request(t, port, "GET", "/stake/pool", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NotNil(t, body)

	initialPool := stake.InitialPool()
	initialPool.LooseTokens = initialPool.LooseTokens.Add(sdk.NewDec(100))
	initialPool.BondedTokens = initialPool.BondedTokens.Add(sdk.NewDec(100))     // Delegate tx on GaiaAppGenState
	initialPool.LooseTokens = initialPool.LooseTokens.Add(sdk.NewDec(int64(50))) // freeFermionsAcc = 50 on GaiaAppGenState

	var pool stake.Pool
	err = cdc.UnmarshalJSON([]byte(body), &pool)
	require.Nil(t, err)
	require.Equal(t, initialPool.BondedTokens, pool.BondedTokens)
	require.Equal(t, initialPool.LooseTokens, pool.LooseTokens)
}

func TestValidatorsQuery(t *testing.T) {
	cleanup, valPubKeys, operAddrs, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()

	require.Equal(t, 1, len(valPubKeys))
	require.Equal(t, 1, len(operAddrs))

	validators := getValidators(t, port)
	require.Equal(t, 1, len(validators), fmt.Sprintf("%+v", validators))

	// make sure all the validators were found (order unknown because sorted by operator addr)
	foundVal := false

	if validators[0].ConsPubKey == valPubKeys[0] {
		foundVal = true
	}

	require.True(t, foundVal, "pk %v, operator %v", operAddrs[0], validators[0].OperatorAddr)
}

func TestValidatorQuery(t *testing.T) {
	cleanup, valPubKeys, operAddrs, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()
	require.Equal(t, 1, len(valPubKeys))
	require.Equal(t, 1, len(operAddrs))

	validator := getValidator(t, port, operAddrs[0])
	require.Equal(t, validator.OperatorAddr, operAddrs[0], "The returned validator does not hold the correct data")
}

func TestBonding(t *testing.T) {
	name, password, denom := "test", "1234567890", stakeTypes.DefaultBondDenom
	addr, seed := CreateAddr(t, name, password, GetKeyBase(t))

	cleanup, valPubKeys, operAddrs, port := InitializeTestLCD(t, 2, []sdk.AccAddress{addr})
	defer cleanup()

	require.Equal(t, 2, len(valPubKeys))
	require.Equal(t, 2, len(operAddrs))

	amt := sdk.NewDec(60)
	validator := getValidator(t, port, operAddrs[0])

	// create bond TX
	resultTx := doDelegate(t, port, seed, name, password, addr, operAddrs[0], 60)
	tests.WaitForHeight(resultTx.Height+1, port)

	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query tx
	txs := getTransactions(t, port,
		fmt.Sprintf("action=delegate&delegator=%s", addr),
		fmt.Sprintf("destination-validator=%s", operAddrs[0]),
	)
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	acc := getAccount(t, port, addr)
	coins := acc.GetCoins()

	require.Equal(t, int64(40), coins.AmountOf(denom).Int64())

	// query delegation
	bond := getDelegation(t, port, addr, operAddrs[0])
	require.Equal(t, amt, bond.Shares)

	delegatorDels := getDelegatorDelegations(t, port, addr)
	require.Len(t, delegatorDels, 1)
	require.Equal(t, amt, delegatorDels[0].Shares)

	// query all delegations to validator
	bonds := getValidatorDelegations(t, port, operAddrs[0])
	require.Len(t, bonds, 2)

	bondedValidators := getDelegatorValidators(t, port, addr)
	require.Len(t, bondedValidators, 1)
	require.Equal(t, operAddrs[0], bondedValidators[0].OperatorAddr)
	require.Equal(t, validator.DelegatorShares.Add(amt).String(), bondedValidators[0].DelegatorShares.String())

	bondedValidator := getDelegatorValidator(t, port, addr, operAddrs[0])
	require.Equal(t, operAddrs[0], bondedValidator.OperatorAddr)

	// testing unbonding
	resultTx = doBeginUnbonding(t, port, seed, name, password, addr, operAddrs[0], 30)
	tests.WaitForHeight(resultTx.Height+1, port)

	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// sender should have not received any coins as the unbonding has only just begun
	acc = getAccount(t, port, addr)
	coins = acc.GetCoins()
	require.Equal(t, int64(40), coins.AmountOf(stakeTypes.DefaultBondDenom).Int64())

	// query tx
	txs = getTransactions(t, port,
		fmt.Sprintf("action=begin_unbonding&delegator=%s", addr),
		fmt.Sprintf("source-validator=%s", operAddrs[0]),
	)
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	unbonding := getUndelegation(t, port, addr, operAddrs[0])
	require.Equal(t, "30", unbonding.Balance.Amount.String())

	// test redelegation
	resultTx = doBeginRedelegation(t, port, seed, name, password, addr, operAddrs[0], operAddrs[1], 30)
	tests.WaitForHeight(resultTx.Height+1, port)

	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query tx
	txs = getTransactions(t, port,
		fmt.Sprintf("action=begin_redelegate&delegator=%s", addr),
		fmt.Sprintf("source-validator=%s", operAddrs[0]),
		fmt.Sprintf("destination-validator=%s", operAddrs[1]),
	)
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	// query delegations, unbondings and redelegations from validator and delegator
	delegatorDels = getDelegatorDelegations(t, port, addr)
	require.Len(t, delegatorDels, 1)
	require.Equal(t, "30.0000000000", delegatorDels[0].GetShares().String())

	delegatorUbds := getDelegatorUnbondingDelegations(t, port, addr)
	require.Len(t, delegatorUbds, 1)
	require.Equal(t, "30", delegatorUbds[0].Balance.Amount.String())

	delegatorReds := getDelegatorRedelegations(t, port, addr)
	require.Len(t, delegatorReds, 1)
	require.Equal(t, "30", delegatorReds[0].Balance.Amount.String())

	validatorUbds := getValidatorUnbondingDelegations(t, port, operAddrs[0])
	require.Len(t, validatorUbds, 1)
	require.Equal(t, "30", validatorUbds[0].Balance.Amount.String())

	validatorReds := getValidatorRedelegations(t, port, operAddrs[0])
	require.Len(t, validatorReds, 1)
	require.Equal(t, "30", validatorReds[0].Balance.Amount.String())

	// TODO Undonding status not currently implemented
	// require.Equal(t, sdk.Unbonding, bondedValidators[0].Status)

	// query txs
	txs = getBondingTxs(t, port, addr, "")
	require.Len(t, txs, 3, "All Txs found")

	txs = getBondingTxs(t, port, addr, "bond")
	require.Len(t, txs, 1, "All bonding txs found")

	txs = getBondingTxs(t, port, addr, "unbond")
	require.Len(t, txs, 1, "All unbonding txs found")

	txs = getBondingTxs(t, port, addr, "redelegate")
	require.Len(t, txs, 1, "All redelegation txs found")
}

func TestSubmitProposal(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name, password, addr, 5)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID)

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	// query tx
	txs := getTransactions(t, port, fmt.Sprintf("action=submit_proposal&proposer=%s", addr))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)
}

func TestDeposit(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name, password, addr, 5)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID)

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	// create SubmitProposal TX
	resultTx = doDeposit(t, port, seed, name, password, addr, proposalID, 5)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query tx
	txs := getTransactions(t, port, fmt.Sprintf("action=deposit&depositor=%s", addr))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	// query proposal
	proposal = getProposal(t, port, proposalID)
	require.True(t, proposal.GetTotalDeposit().IsEqual(sdk.Coins{sdk.NewInt64Coin(stakeTypes.DefaultBondDenom, 10)}))

	// query deposit
	deposit := getDeposit(t, port, proposalID, addr)
	require.True(t, deposit.Amount.IsEqual(sdk.Coins{sdk.NewInt64Coin(stakeTypes.DefaultBondDenom, 10)}))
}

func TestVote(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, operAddrs, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name, password, addr, 5)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID)

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	// deposit
	resultTx = doDeposit(t, port, seed, name, password, addr, proposalID, 5)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query proposal
	proposal = getProposal(t, port, proposalID)
	require.Equal(t, gov.StatusVotingPeriod, proposal.GetStatus())

	// vote
	resultTx = doVote(t, port, seed, name, password, addr, proposalID)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query tx
	txs := getTransactions(t, port, fmt.Sprintf("action=vote&voter=%s", addr))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	vote := getVote(t, port, proposalID, addr)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, gov.OptionYes, vote.Option)

	tally := getTally(t, port, proposalID)
	require.Equal(t, sdk.ZeroDec(), tally.Yes, "tally should be 0 as the address is not bonded")

	// create bond TX
	resultTx = doDelegate(t, port, seed, name, password, addr, operAddrs[0], 60)
	tests.WaitForHeight(resultTx.Height+1, port)

	// vote
	resultTx = doVote(t, port, seed, name, password, addr, proposalID)
	tests.WaitForHeight(resultTx.Height+1, port)

	tally = getTally(t, port, proposalID)
	require.Equal(t, sdk.NewDec(60), tally.Yes, "tally should be equal to the amount delegated")
}

func TestUnjail(t *testing.T) {
	_, password := "test", "1234567890"
	addr, _ := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, valPubKeys, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// XXX: any less than this and it fails
	tests.WaitForHeight(3, port)
	pkString, _ := sdk.Bech32ifyConsPub(valPubKeys[0])
	signingInfo := getSigningInfo(t, port, pkString)
	tests.WaitForHeight(4, port)
	require.Equal(t, true, signingInfo.IndexOffset > 0)
	require.Equal(t, time.Unix(0, 0).UTC(), signingInfo.JailedUntil)
	require.Equal(t, true, signingInfo.MissedBlocksCounter == 0)
}

func TestProposalsQuery(t *testing.T) {
	addrs, seeds, names, passwords := CreateAddrs(t, GetKeyBase(t), 2)

	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addrs[0], addrs[1]})
	defer cleanup()

	depositParam := getDepositParam(t, port)
	halfMinDeposit := depositParam.MinDeposit.AmountOf(stakeTypes.DefaultBondDenom).Int64() / 2
	getVotingParam(t, port)
	getTallyingParam(t, port)

	// Addr1 proposes (and deposits) proposals #1 and #2
	resultTx := doSubmitProposal(t, port, seeds[0], names[0], passwords[0], addrs[0], halfMinDeposit)
	var proposalID1 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID1)
	tests.WaitForHeight(resultTx.Height+1, port)

	resultTx = doSubmitProposal(t, port, seeds[0], names[0], passwords[0], addrs[0], halfMinDeposit)
	var proposalID2 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID2)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 proposes (and deposits) proposals #3
	resultTx = doSubmitProposal(t, port, seeds[1], names[1], passwords[1], addrs[1], halfMinDeposit)
	var proposalID3 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID3)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 deposits on proposals #2 & #3
	resultTx = doDeposit(t, port, seeds[1], names[1], passwords[1], addrs[1], proposalID2, halfMinDeposit)
	tests.WaitForHeight(resultTx.Height+1, port)

	resultTx = doDeposit(t, port, seeds[1], names[1], passwords[1], addrs[1], proposalID3, halfMinDeposit)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check deposits match proposal and individual deposits
	deposits := getDeposits(t, port, proposalID1)
	require.Len(t, deposits, 1)
	deposit := getDeposit(t, port, proposalID1, addrs[0])
	require.Equal(t, deposit, deposits[0])

	deposits = getDeposits(t, port, proposalID2)
	require.Len(t, deposits, 2)
	deposit = getDeposit(t, port, proposalID2, addrs[0])
	require.True(t, deposit.Equals(deposits[0]))
	deposit = getDeposit(t, port, proposalID2, addrs[1])
	require.True(t, deposit.Equals(deposits[1]))

	deposits = getDeposits(t, port, proposalID3)
	require.Len(t, deposits, 1)
	deposit = getDeposit(t, port, proposalID3, addrs[1])
	require.Equal(t, deposit, deposits[0])

	// increasing the amount of the deposit should update the existing one
	resultTx = doDeposit(t, port, seeds[0], names[0], passwords[0], addrs[0], proposalID1, 1)
	tests.WaitForHeight(resultTx.Height+1, port)

	deposits = getDeposits(t, port, proposalID1)
	require.Len(t, deposits, 1)

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
	resultTx = doVote(t, port, seeds[0], names[0], passwords[0], addrs[0], proposalID2)
	tests.WaitForHeight(resultTx.Height+1, port)
	resultTx = doVote(t, port, seeds[0], names[0], passwords[0], addrs[0], proposalID3)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 votes on proposal #3
	resultTx = doVote(t, port, seeds[1], names[1], passwords[1], addrs[1], proposalID3)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Test query all proposals
	proposals = getProposalsAll(t, port)
	require.Equal(t, proposalID1, (proposals[0]).GetProposalID())
	require.Equal(t, proposalID2, (proposals[1]).GetProposalID())
	require.Equal(t, proposalID3, (proposals[2]).GetProposalID())

	// Test query deposited by addr1
	proposals = getProposalsFilterDepositor(t, port, addrs[0])
	require.Equal(t, proposalID1, (proposals[0]).GetProposalID())

	// Test query deposited by addr2
	proposals = getProposalsFilterDepositor(t, port, addrs[1])
	require.Equal(t, proposalID2, (proposals[0]).GetProposalID())
	require.Equal(t, proposalID3, (proposals[1]).GetProposalID())

	// Test query voted by addr1
	proposals = getProposalsFilterVoter(t, port, addrs[0])
	require.Equal(t, proposalID2, (proposals[0]).GetProposalID())
	require.Equal(t, proposalID3, (proposals[1]).GetProposalID())

	// Test query voted by addr2
	proposals = getProposalsFilterVoter(t, port, addrs[1])
	require.Equal(t, proposalID3, (proposals[0]).GetProposalID())

	// Test query voted and deposited by addr1
	proposals = getProposalsFilterVoterDepositor(t, port, addrs[0], addrs[0])
	require.Equal(t, proposalID2, (proposals[0]).GetProposalID())

	// Test query votes on Proposal 2
	votes := getVotes(t, port, proposalID2)
	require.Len(t, votes, 1)
	require.Equal(t, addrs[0], votes[0].Voter)

	// Test query votes on Proposal 3
	votes = getVotes(t, port, proposalID3)
	require.Len(t, votes, 2)
	require.True(t, addrs[0].String() == votes[0].Voter.String() || addrs[0].String() == votes[1].Voter.String())
	require.True(t, addrs[1].String() == votes[0].Voter.String() || addrs[1].String() == votes[1].Voter.String())
}

//_____________________________________________________________________________
// get the account to get the sequence
func getAccount(t *testing.T, port string, addr sdk.AccAddress) auth.Account {
	res, body := Request(t, port, "GET", fmt.Sprintf("/auth/accounts/%s", addr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var acc auth.Account
	err := cdc.UnmarshalJSON([]byte(body), &acc)
	require.Nil(t, err)
	return acc
}

func doSendWithGas(t *testing.T, port, seed, name, password string, addr sdk.AccAddress, gas string,
	gasAdjustment float64, simulate, generateOnly bool) (
	res *http.Response, body string, receiveAddr sdk.AccAddress) {

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
	coinbz, err := cdc.MarshalJSON(sdk.NewInt64Coin(stakeTypes.DefaultBondDenom, 1))
	if err != nil {
		panic(err)
	}

	gasStr := ""
	if len(gas) != 0 {
		gasStr = fmt.Sprintf(`
		"gas":%q,
		`, gas)
	}
	gasAdjustmentStr := ""
	if gasAdjustment > 0 {
		gasAdjustmentStr = fmt.Sprintf(`
		"gas_adjustment":"%v",
		`, gasAdjustment)
	}
	jsonStr := []byte(fmt.Sprintf(`{
		"amount":[%s],
		"base_req": {
			%v%v
			"name": "%s",
			"password": "%s",
			"chain_id": "%s",
			"account_number":"%d",
			"sequence": "%d",
			"simulate": %v,
			"generate_only": %v
		}
	}`, coinbz, gasStr, gasAdjustmentStr, name, password, chainID, accnum, sequence, simulate, generateOnly))

	res, body = Request(t, port, "POST", fmt.Sprintf("/bank/accounts/%s/transfers", receiveAddr), jsonStr)
	return
}

func doRecoverKey(t *testing.T, port, recoverName, recoverPassword, seed string) {
	jsonStr := []byte(fmt.Sprintf(`{"password":"%s", "seed":"%s"}`, recoverPassword, seed))
	res, body := Request(t, port, "POST", fmt.Sprintf("/keys/%s/recover", recoverName), jsonStr)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.KeyOutput
	err := codec.Cdc.UnmarshalJSON([]byte(body), &resp)
	require.Nil(t, err, body)

	addr1Bech32 := resp.Address
	_, err = sdk.AccAddressFromBech32(addr1Bech32)
	require.NoError(t, err, "Failed to return a correct bech32 address")
}

func doSend(t *testing.T, port, seed, name, password string, addr sdk.AccAddress) (receiveAddr sdk.AccAddress, resultTx ctypes.ResultBroadcastTxCommit) {
	res, body, receiveAddr := doSendWithGas(t, port, seed, name, password, addr, "", 0, false, false)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	return receiveAddr, resultTx
}

func getTransactions(t *testing.T, port string, tags ...string) []tx.Info {
	var txs []tx.Info
	if len(tags) == 0 {
		return txs
	}
	queryStr := strings.Join(tags, "&")
	res, body := Request(t, port, "GET", fmt.Sprintf("/txs?%s", queryStr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &txs)
	require.NoError(t, err)
	return txs
}

// ============= IBC Module ================

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
		"amount":[
			{
				"denom": "%s",
				"amount": "1"
			}
		],
		"base_req": {
			"name": "%s",
			"password": "%s",
			"chain_id": "%s",
			"account_number":"%d",
			"sequence":"%d"
		}
	}`, stakeTypes.DefaultBondDenom, name, password, chainID, accnum, sequence))

	res, body := Request(t, port, "POST", fmt.Sprintf("/ibc/testchain/%s/send", receiveAddr), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	return resultTx
}

// ============= Slashing Module ================

func getSigningInfo(t *testing.T, port string, validatorPubKey string) slashing.ValidatorSigningInfo {
	res, body := Request(t, port, "GET", fmt.Sprintf("/slashing/validators/%s/signing_info", validatorPubKey), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var signingInfo slashing.ValidatorSigningInfo
	err := cdc.UnmarshalJSON([]byte(body), &signingInfo)
	require.Nil(t, err)

	return signingInfo
}

func doUnjail(t *testing.T, port, seed, name, password string,
	valAddr sdk.ValAddress) (resultTx ctypes.ResultBroadcastTxCommit) {
	chainID := viper.GetString(client.FlagChainID)

	jsonStr := []byte(fmt.Sprintf(`{
		"base_req": {
			"name": "%s",
			"password": "%s",
			"chain_id": "%s",
			"account_number":"1",
			"sequence":"1"
		}
	}`, name, password, chainID))

	res, body := Request(t, port, "POST", fmt.Sprintf("/slashing/validators/%s/unjail", valAddr.String()), jsonStr)
	// TODO : fails with "401 must use own validator address"
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

// ============= Stake Module ================

func getDelegation(t *testing.T, port string, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) stake.Delegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/delegations/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bond stake.Delegation
	err := cdc.UnmarshalJSON([]byte(body), &bond)
	require.Nil(t, err)

	return bond
}

func getUndelegation(t *testing.T, port string, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) stake.UnbondingDelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/unbonding_delegations/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var unbond stake.UnbondingDelegation
	err := cdc.UnmarshalJSON([]byte(body), &unbond)
	require.Nil(t, err)

	return unbond
}

func getDelegatorDelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress) []stake.Delegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/delegations", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var dels []stake.Delegation

	err := cdc.UnmarshalJSON([]byte(body), &dels)
	require.Nil(t, err)

	return dels
}

func getDelegatorUnbondingDelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress) []stake.UnbondingDelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/unbonding_delegations", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var ubds []stake.UnbondingDelegation

	err := cdc.UnmarshalJSON([]byte(body), &ubds)
	require.Nil(t, err)

	return ubds
}

func getDelegatorRedelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress) []stake.Redelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/redelegations", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var reds []stake.Redelegation

	err := cdc.UnmarshalJSON([]byte(body), &reds)
	require.Nil(t, err)

	return reds
}

func getBondingTxs(t *testing.T, port string, delegatorAddr sdk.AccAddress, query string) []tx.Info {
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

func getDelegatorValidators(t *testing.T, port string, delegatorAddr sdk.AccAddress) []stake.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/validators", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bondedValidators []stake.Validator

	err := cdc.UnmarshalJSON([]byte(body), &bondedValidators)
	require.Nil(t, err)

	return bondedValidators
}

func getDelegatorValidator(t *testing.T, port string, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) stake.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/validators/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bondedValidator stake.Validator
	err := cdc.UnmarshalJSON([]byte(body), &bondedValidator)
	require.Nil(t, err)

	return bondedValidator
}

func doDelegate(t *testing.T, port, seed, name, password string,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount int64) (resultTx ctypes.ResultBroadcastTxCommit) {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)

	jsonStr := []byte(fmt.Sprintf(`{
		"delegations": [
			{
				"delegator_addr": "%s",
				"validator_addr": "%s",
				"delegation": { "denom": "%s", "amount": "%d" }
			}
		],
		"begin_unbondings": [],
		"begin_redelegates": [],
		"base_req": {
			"name": "%s",
			"password": "%s",
			"chain_id": "%s",
			"account_number":"%d",
			"sequence":"%d"
		}
	}`, delAddr, valAddr, stakeTypes.DefaultBondDenom, amount, name, password, chainID, accnum, sequence))

	res, body := Request(t, port, "POST", fmt.Sprintf("/stake/delegators/%s/delegations", delAddr), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func doBeginUnbonding(t *testing.T, port, seed, name, password string,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount int64) (resultTx ctypes.ResultBroadcastTxCommit) {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)

	jsonStr := []byte(fmt.Sprintf(`{
		"delegations": [],
		"begin_unbondings": [
			{
				"delegator_addr": "%s",
				"validator_addr": "%s",
				"shares": "%d"
			}
		],
		"begin_redelegates": [],
		"base_req": {
			"name": "%s",
			"password": "%s",
			"chain_id": "%s",
			"account_number":"%d",
			"sequence":"%d"
		}
	}`, delAddr, valAddr, amount, name, password, chainID, accnum, sequence))

	res, body := Request(t, port, "POST", fmt.Sprintf("/stake/delegators/%s/delegations", delAddr), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func doBeginRedelegation(t *testing.T, port, seed, name, password string,
	delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress, amount int64) (resultTx ctypes.ResultBroadcastTxCommit) {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)

	jsonStr := []byte(fmt.Sprintf(`{
		"delegations": [],
		"begin_unbondings": [],
		"begin_redelegates": [
			{
				"delegator_addr": "%s",
				"validator_src_addr": "%s",
				"validator_dst_addr": "%s",
				"shares": "%d"
			}
		],
		"base_req": {
			"name": "%s",
			"password": "%s",
			"chain_id": "%s",
			"account_number":"%d",
			"sequence":"%d"
		}
	}`, delAddr, valSrcAddr, valDstAddr, amount, name, password, chainID, accnum, sequence))

	res, body := Request(t, port, "POST", fmt.Sprintf("/stake/delegators/%s/delegations", delAddr), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func getValidators(t *testing.T, port string) []stake.Validator {
	res, body := Request(t, port, "GET", "/stake/validators", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var validators []stake.Validator
	err := cdc.UnmarshalJSON([]byte(body), &validators)
	require.Nil(t, err)

	return validators
}

func getValidator(t *testing.T, port string, validatorAddr sdk.ValAddress) stake.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/validators/%s", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var validator stake.Validator
	err := cdc.UnmarshalJSON([]byte(body), &validator)
	require.Nil(t, err)

	return validator
}

func getValidatorDelegations(t *testing.T, port string, validatorAddr sdk.ValAddress) []stake.Delegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/validators/%s/delegations", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var delegations []stake.Delegation
	err := cdc.UnmarshalJSON([]byte(body), &delegations)
	require.Nil(t, err)

	return delegations
}

func getValidatorUnbondingDelegations(t *testing.T, port string, validatorAddr sdk.ValAddress) []stake.UnbondingDelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/validators/%s/unbonding_delegations", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var ubds []stake.UnbondingDelegation
	err := cdc.UnmarshalJSON([]byte(body), &ubds)
	require.Nil(t, err)

	return ubds
}

func getValidatorRedelegations(t *testing.T, port string, validatorAddr sdk.ValAddress) []stake.Redelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/validators/%s/redelegations", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var reds []stake.Redelegation
	err := cdc.UnmarshalJSON([]byte(body), &reds)
	require.Nil(t, err)

	return reds
}

// ============= Governance Module ================

func getDepositParam(t *testing.T, port string) gov.DepositParams {
	res, body := Request(t, port, "GET", "/gov/parameters/deposit", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var depositParams gov.DepositParams
	err := cdc.UnmarshalJSON([]byte(body), &depositParams)
	require.Nil(t, err)
	return depositParams
}

func getVotingParam(t *testing.T, port string) gov.VotingParams {
	res, body := Request(t, port, "GET", "/gov/parameters/voting", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var votingParams gov.VotingParams
	err := cdc.UnmarshalJSON([]byte(body), &votingParams)
	require.Nil(t, err)
	return votingParams
}

func getTallyingParam(t *testing.T, port string) gov.TallyParams {
	res, body := Request(t, port, "GET", "/gov/parameters/tallying", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var tallyParams gov.TallyParams
	err := cdc.UnmarshalJSON([]byte(body), &tallyParams)
	require.Nil(t, err)
	return tallyParams
}

func getProposal(t *testing.T, port string, proposalID uint64) gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var proposal gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposal)
	require.Nil(t, err)
	return proposal
}

func getDeposits(t *testing.T, port string, proposalID uint64) []gov.Deposit {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/deposits", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var deposits []gov.Deposit
	err := cdc.UnmarshalJSON([]byte(body), &deposits)
	require.Nil(t, err)
	return deposits
}

func getDeposit(t *testing.T, port string, proposalID uint64, depositorAddr sdk.AccAddress) gov.Deposit {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/deposits/%s", proposalID, depositorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var deposit gov.Deposit
	err := cdc.UnmarshalJSON([]byte(body), &deposit)
	require.Nil(t, err)
	return deposit
}

func getVote(t *testing.T, port string, proposalID uint64, voterAddr sdk.AccAddress) gov.Vote {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/votes/%s", proposalID, voterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var vote gov.Vote
	err := cdc.UnmarshalJSON([]byte(body), &vote)
	require.Nil(t, err)
	return vote
}

func getVotes(t *testing.T, port string, proposalID uint64) []gov.Vote {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/votes", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var votes []gov.Vote
	err := cdc.UnmarshalJSON([]byte(body), &votes)
	require.Nil(t, err)
	return votes
}

func getTally(t *testing.T, port string, proposalID uint64) gov.TallyResult {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/tally", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var tally gov.TallyResult
	err := cdc.UnmarshalJSON([]byte(body), &tally)
	require.Nil(t, err)
	return tally
}

func getProposalsAll(t *testing.T, port string) []gov.Proposal {
	res, body := Request(t, port, "GET", "/gov/proposals", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

func getProposalsFilterDepositor(t *testing.T, port string, depositorAddr sdk.AccAddress) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?depositor=%s", depositorAddr), nil)
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

func getProposalsFilterVoterDepositor(t *testing.T, port string, voterAddr, depositorAddr sdk.AccAddress) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?depositor=%s&voter=%s", depositorAddr, voterAddr), nil)
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

func doSubmitProposal(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress, amount int64) (resultTx ctypes.ResultBroadcastTxCommit) {

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
		"initial_deposit": [{ "denom": "%s", "amount": "%d" }],
		"base_req": {
			"name": "%s",
			"password": "%s",
			"chain_id": "%s",
			"account_number":"%d",
			"sequence":"%d"
		}
	}`, proposerAddr, stakeTypes.DefaultBondDenom, amount, name, password, chainID, accnum, sequence))
	res, body := Request(t, port, "POST", "/gov/proposals", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

func doDeposit(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress, proposalID uint64, amount int64) (resultTx ctypes.ResultBroadcastTxCommit) {

	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)

	// deposit on proposal
	jsonStr := []byte(fmt.Sprintf(`{
		"depositor": "%s",
		"amount": [{ "denom": "%s", "amount": "%d" }],
		"base_req": {
			"name": "%s",
			"password": "%s",
			"chain_id": "%s",
			"account_number":"%d",
			"sequence": "%d"
		}
	}`, proposerAddr, stakeTypes.DefaultBondDenom, amount, name, password, chainID, accnum, sequence))
	res, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/deposits", proposalID), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

func doVote(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress, proposalID uint64) (resultTx ctypes.ResultBroadcastTxCommit) {
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
			"sequence": "%d"
		}
	}`, proposerAddr, name, password, chainID, accnum, sequence))
	res, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/votes", proposalID), jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}
