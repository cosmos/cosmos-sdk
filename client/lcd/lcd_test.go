package lcd

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/bank"
	dclcommon "github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	distrrest "github.com/cosmos/cosmos-sdk/x/distribution/client/rest"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

const (
	name1 = "test1"
	name2 = "test2"
	name3 = "test3"
	memo  = "LCD test tx"
	pw    = app.DefaultKeyPass
	altPw = "12345678901"
)

var fees = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}

func init() {
	mintkey.BcryptSecurityParameter = 1
	version.Version = os.Getenv("VERSION")
}

func TestSeedsAreDifferent(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, _ := CreateAddr(t, name1, pw, kb)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
	defer cleanup()

	mnemonic1 := getKeysSeed(t, port)
	mnemonic2 := getKeysSeed(t, port)

	require.NotEqual(t, mnemonic1, mnemonic2)
}

func TestKeyRecover(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{}, true)
	defer cleanup()

	myName1 := "TestKeyRecover_1"
	myName2 := "TestKeyRecover_2"

	mnemonic := getKeysSeed(t, port)
	expectedInfo, _ := kb.CreateAccount(myName1, mnemonic, "", pw, 0, 0)
	expectedAddress := expectedInfo.GetAddress().String()
	expectedPubKey := sdk.MustBech32ifyAccPub(expectedInfo.GetPubKey())

	// recover key
	doRecoverKey(t, port, myName2, pw, mnemonic, 0, 0)

	keys := getKeys(t, port)

	require.Equal(t, expectedAddress, keys[0].Address)
	require.Equal(t, expectedPubKey, keys[0].PubKey)
}

func TestKeyRecoverHDPath(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{}, true)
	defer cleanup()

	mnemonic := getKeysSeed(t, port)

	for account := uint32(0); account < 50; account += 13 {
		for index := uint32(0); index < 50; index += 15 {
			name1Idx := fmt.Sprintf("name1_%d_%d", account, index)
			name2Idx := fmt.Sprintf("name2_%d_%d", account, index)

			expectedInfo, _ := kb.CreateAccount(name1Idx, mnemonic, "", pw, account, index)
			expectedAddress := expectedInfo.GetAddress().String()
			expectedPubKey := sdk.MustBech32ifyAccPub(expectedInfo.GetPubKey())

			// recover key
			doRecoverKey(t, port, name2Idx, pw, mnemonic, account, index)

			keysName2Idx := getKey(t, port, name2Idx)

			require.Equal(t, expectedAddress, keysName2Idx.Address)
			require.Equal(t, expectedPubKey, keysName2Idx.PubKey)
		}
	}
}

func TestKeys(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr1, _ := CreateAddr(t, name1, pw, kb)
	addr1Bech32 := addr1.String()

	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr1}, true)
	defer cleanup()

	// get new seed & recover key
	mnemonic2 := getKeysSeed(t, port)
	doRecoverKey(t, port, name2, pw, mnemonic2, 0, 0)

	// add key
	mnemonic3 := mnemonic2
	resp := doKeysPost(t, port, name3, pw, mnemonic3, 0, 0)

	addr3Bech32 := resp.Address
	_, err = sdk.AccAddressFromBech32(addr3Bech32)
	require.NoError(t, err, "Failed to return a correct bech32 address")

	// test if created account is the correct account
	expectedInfo3, _ := kb.CreateAccount(name3, mnemonic3, "", pw, 0, 0)
	expectedAddress3 := sdk.AccAddress(expectedInfo3.GetPubKey().Address()).String()
	require.Equal(t, expectedAddress3, addr3Bech32)

	// existing keys
	require.Equal(t, name1, getKey(t, port, name1).Name, "Did not serve keys name correctly")
	require.Equal(t, addr1Bech32, getKey(t, port, name1).Address, "Did not serve keys Address correctly")
	require.Equal(t, name2, getKey(t, port, name2).Name, "Did not serve keys name correctly")
	require.Equal(t, addr3Bech32, getKey(t, port, name2).Address, "Did not serve keys Address correctly")
	require.Equal(t, name3, getKey(t, port, name3).Name, "Did not serve keys name correctly")
	require.Equal(t, addr3Bech32, getKey(t, port, name3).Address, "Did not serve keys Address correctly")

	// select key
	key := getKey(t, port, name3)
	require.Equal(t, name3, key.Name, "Did not serve keys name correctly")
	require.Equal(t, addr3Bech32, key.Address, "Did not serve keys Address correctly")

	// update key
	updateKey(t, port, name3, pw, altPw, false)

	// here it should say unauthorized as we changed the password before
	updateKey(t, port, name3, pw, altPw, true)

	// delete key
	deleteKey(t, port, name3, altPw)
}

func TestVersion(t *testing.T) {
	// skip the test if the VERSION environment variable has not been set
	if version.Version == "" {
		t.SkipNow()
	}

	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{}, true)
	defer cleanup()

	// node info
	res, body := Request(t, port, "GET", "/version", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	reg, err := regexp.Compile(`\d+\.\d+\.\d+.*`)
	require.Nil(t, err)
	match := reg.MatchString(body)
	require.True(t, match, body, body)

	// node info
	res, body = Request(t, port, "GET", "/node_version", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	reg, err = regexp.Compile(`\d+\.\d+\.\d+.*`)
	require.Nil(t, err)
	match = reg.MatchString(body)
	require.True(t, match, body)
}

func TestNodeStatus(t *testing.T) {
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{}, true)
	defer cleanup()
	getNodeInfo(t, port)
	getSyncStatus(t, port, false)
}

func TestBlock(t *testing.T) {
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{}, true)
	defer cleanup()
	getBlock(t, port, -1, false)
	getBlock(t, port, 2, false)
	getBlock(t, port, 100000000, true)
}

func TestValidators(t *testing.T) {
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{}, true)
	defer cleanup()
	resultVals := getValidatorSets(t, port, -1, false)
	require.Contains(t, resultVals.Validators[0].Address.String(), "cosmosvalcons")
	require.Contains(t, resultVals.Validators[0].PubKey, "cosmosvalconspub")
	getValidatorSets(t, port, 2, false)
	getValidatorSets(t, port, 10000000, true)
}

func TestCoinSend(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, seed := CreateAddr(t, name1, pw, kb)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
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
	receiveAddr, resultTx := doTransfer(t, port, seed, name1, memo, pw, addr, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.Code)

	// query sender
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	expectedBalance := initialBalance[0].Minus(fees[0])

	require.Equal(t, sdk.DefaultBondDenom, coins[0].Denom)
	require.Equal(t, expectedBalance.Amount.SubRaw(1), coins[0].Amount)
	expectedBalance = coins[0]

	// query receiver
	acc2 := getAccount(t, port, receiveAddr)
	coins2 := acc2.GetCoins()
	require.Equal(t, sdk.DefaultBondDenom, coins2[0].Denom)
	require.Equal(t, int64(1), coins2[0].Amount.Int64())

	// test failure with too little gas
	res, body, _ = doTransferWithGas(t, port, seed, name1, memo, pw, addr, "100", 0, false, false, fees)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode, body)
	require.Nil(t, err)

	// test failure with negative gas
	res, body, _ = doTransferWithGas(t, port, seed, name1, memo, pw, addr, "-200", 0, false, false, fees)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, body)

	// test failure with negative adjustment
	res, body, _ = doTransferWithGas(t, port, seed, name1, memo, pw, addr, "10000", -0.1, true, false, fees)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, body)

	// test failure with 0 gas
	res, body, _ = doTransferWithGas(t, port, seed, name1, memo, pw, addr, "0", 0, false, false, fees)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode, body)

	// test failure with wrong adjustment
	res, body, _ = doTransferWithGas(t, port, seed, name1, memo, pw, addr, client.GasFlagAuto, 0.1, false, false, fees)

	require.Equal(t, http.StatusInternalServerError, res.StatusCode, body)

	// run simulation and test success with estimated gas
	res, body, _ = doTransferWithGas(
		t, port, seed, name1, memo, pw, addr, "10000", 1.0, true, false, fees,
	)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var gasEstResp rest.GasEstimateResponse
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &gasEstResp))
	require.NotZero(t, gasEstResp.GasEstimate)

	acc = getAccount(t, port, addr)
	require.Equal(t, expectedBalance.Amount, acc.GetCoins().AmountOf(sdk.DefaultBondDenom))

	// run successful tx
	gas := fmt.Sprintf("%d", gasEstResp.GasEstimate)
	res, body, _ = doTransferWithGas(t, port, seed, name1, memo, pw, addr, gas, 1.0, false, false, fees)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	tests.WaitForHeight(resultTx.Height+1, port)
	require.Equal(t, uint32(0), resultTx.Code)

	acc = getAccount(t, port, addr)
	expectedBalance = expectedBalance.Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.SubRaw(1), acc.GetCoins().AmountOf(sdk.DefaultBondDenom))
}

func TestCoinSendAccAuto(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, seed := CreateAddr(t, name1, pw, kb)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
	defer cleanup()

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// send a transfer tx without specifying account number and sequence
	res, body, _ := doTransferWithGasAccAuto(t, port, seed, name1, memo, pw, "200000", 1.0, false, false, fees)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// query sender
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	expectedBalance := initialBalance[0].Minus(fees[0])

	require.Equal(t, sdk.DefaultBondDenom, coins[0].Denom)
	require.Equal(t, expectedBalance.Amount.SubRaw(1), coins[0].Amount)
}

func TestCoinMultiSendGenerateOnly(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, seed := CreateAddr(t, name1, pw, kb)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
	defer cleanup()

	// generate only
	res, body, _ := doTransferWithGas(t, port, seed, "", memo, "", addr, "200000", 1, false, true, fees)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var stdTx auth.StdTx
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &stdTx))
	require.Equal(t, len(stdTx.Msgs), 1)
	require.Equal(t, stdTx.GetMsgs()[0].Route(), "bank")
	require.Equal(t, stdTx.GetMsgs()[0].GetSigners(), []sdk.AccAddress{addr})
	require.Equal(t, 0, len(stdTx.Signatures))
	require.Equal(t, memo, stdTx.Memo)
	require.NotZero(t, stdTx.Fee.Gas)
	require.IsType(t, stdTx.GetMsgs()[0], bank.MsgSend{})
	require.Equal(t, addr, stdTx.GetMsgs()[0].(bank.MsgSend).FromAddress)
}

func TestCoinSendGenerateSignAndBroadcast(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, seed := CreateAddr(t, name1, pw, kb)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
	defer cleanup()
	acc := getAccount(t, port, addr)

	// simulate tx
	res, body, _ := doTransferWithGas(
		t, port, seed, name1, memo, "", addr, client.GasFlagAuto, 1.0, true, false, fees,
	)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var gasEstResp rest.GasEstimateResponse
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &gasEstResp))
	require.NotZero(t, gasEstResp.GasEstimate)

	// generate tx
	gas := fmt.Sprintf("%d", gasEstResp.GasEstimate)
	res, body, _ = doTransferWithGas(t, port, seed, name1, memo, "", addr, gas, 1, false, true, fees)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var msg auth.StdTx
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &msg))
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, msg.Msgs[0].Route(), "bank")
	require.Equal(t, msg.Msgs[0].GetSigners(), []sdk.AccAddress{addr})
	require.Equal(t, 0, len(msg.Signatures))
	require.Equal(t, memo, msg.Memo)
	require.NotZero(t, msg.Fee.Gas)

	gasEstimate := int64(msg.Fee.Gas)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	// sign tx
	var signedMsg auth.StdTx

	payload := authrest.SignBody{
		Tx: msg,
		BaseReq: rest.NewBaseReq(
			name1, pw, "", viper.GetString(client.FlagChainID), "", "",
			accnum, sequence, nil, nil, false, false,
		),
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
	var resultTx sdk.TxResponse
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &resultTx))
	require.Equal(t, uint32(0), resultTx.Code)
	require.Equal(t, gasEstimate, resultTx.GasWanted)
}

func TestEncodeTx(t *testing.T) {
	// Setup
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, seed := CreateAddr(t, name1, pw, kb)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
	defer cleanup()

	// Make a transaction to test with
	res, body, _ := doTransferWithGas(t, port, seed, name1, memo, "", addr, "2", 1, false, true, fees)
	var tx auth.StdTx
	cdc.UnmarshalJSON([]byte(body), &tx)

	// Build the request
	encodeReq := struct {
		Tx auth.StdTx `json:"tx"`
	}{Tx: tx}
	encodedJSON, _ := cdc.MarshalJSON(encodeReq)
	res, body = Request(t, port, "POST", "/tx/encode", encodedJSON)

	// Make sure it came back ok, and that we can decode it back to the transaction
	// 200 response
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	encodeResp := struct {
		Tx string `json:"tx"`
	}{}

	// No error decoding the JSON
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &encodeResp))

	// Check that the base64 decodes
	decodedBytes, err := base64.StdEncoding.DecodeString(encodeResp.Tx)
	require.Nil(t, err)

	// Check that the transaction decodes as expected
	var decodedTx auth.StdTx
	require.Nil(t, cdc.UnmarshalBinaryLengthPrefixed(decodedBytes, &decodedTx))
	require.Equal(t, memo, decodedTx.Memo)
}

func TestTxs(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, seed := CreateAddr(t, name1, pw, kb)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
	defer cleanup()

	var emptyTxs []sdk.TxResponse
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
	receiveAddr, resultTx := doTransfer(t, port, seed, name1, memo, pw, addr, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx is queryable
	tx := getTransaction(t, port, resultTx.TxHash)
	require.Equal(t, resultTx.TxHash, tx.TxHash)

	// query sender
	txs = getTransactions(t, port, fmt.Sprintf("sender=%s", addr.String()))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	// query recipient
	txs = getTransactions(t, port, fmt.Sprintf("recipient=%s", receiveAddr.String()))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	// query transaction that doesn't exist
	validTxHash := "9ADBECAAD8DACBEC3F4F535704E7CF715C765BDCEDBEF086AFEAD31BA664FB0B"
	res, body := getTransactionRequest(t, port, validTxHash)
	require.True(t, strings.Contains(body, validTxHash))
	require.Equal(t, http.StatusNotFound, res.StatusCode)

	// bad query string
	res, body = getTransactionRequest(t, port, "badtxhash")
	require.True(t, strings.Contains(body, "encoding/hex"))
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestPoolParamsQuery(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, _ := CreateAddr(t, name1, pw, kb)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
	defer cleanup()

	defaultParams := staking.DefaultParams()

	params := getStakingParams(t, port)
	require.True(t, defaultParams.Equal(params))

	pool := getStakingPool(t, port)

	initialPool := staking.InitialPool()
	tokens := sdk.TokensFromTendermintPower(100)
	freeTokens := sdk.TokensFromTendermintPower(50)
	initialPool.NotBondedTokens = initialPool.NotBondedTokens.Add(tokens)
	initialPool.BondedTokens = initialPool.BondedTokens.Add(tokens)           // Delegate tx on GaiaAppGenState
	initialPool.NotBondedTokens = initialPool.NotBondedTokens.Add(freeTokens) // freeTokensPerAcc = 50 on GaiaAppGenState

	require.Equal(t, initialPool.BondedTokens, pool.BondedTokens)

	//TODO include this test once REST for distribution is online, need to include distribution tokens from inflation
	//     for this equality to make sense
	//require.Equal(t, initialPool.NotBondedTokens, pool.NotBondedTokens)
}

func TestValidatorsQuery(t *testing.T) {
	cleanup, valPubKeys, operAddrs, port := InitializeTestLCD(t, 1, []sdk.AccAddress{}, true)
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
	cleanup, valPubKeys, operAddrs, port := InitializeTestLCD(t, 1, []sdk.AccAddress{}, true)
	defer cleanup()
	require.Equal(t, 1, len(valPubKeys))
	require.Equal(t, 1, len(operAddrs))

	validator := getValidator(t, port, operAddrs[0])
	require.Equal(t, validator.OperatorAddr, operAddrs[0], "The returned validator does not hold the correct data")
}

func TestBonding(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, _ := CreateAddr(t, name1, pw, kb)

	cleanup, valPubKeys, operAddrs, port := InitializeTestLCD(t, 2, []sdk.AccAddress{addr}, false)
	tests.WaitForHeight(1, port)
	defer cleanup()

	require.Equal(t, 2, len(valPubKeys))
	require.Equal(t, 2, len(operAddrs))

	amt := sdk.TokensFromTendermintPower(60)
	amtDec := sdk.NewDecFromInt(amt)
	validator := getValidator(t, port, operAddrs[0])

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create bond TX
	delTokens := sdk.TokensFromTendermintPower(60)
	resultTx := doDelegate(t, port, name1, pw, addr, operAddrs[0], delTokens, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	require.Equal(t, uint32(0), resultTx.Code)

	// query tx
	txs := getTransactions(t, port,
		fmt.Sprintf("action=delegate&delegator=%s", addr),
		fmt.Sprintf("destination-validator=%s", operAddrs[0]),
	)
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	// verify balance
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	expectedBalance := initialBalance[0].Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(delTokens), coins.AmountOf(sdk.DefaultBondDenom))
	expectedBalance = coins[0]

	// query delegation
	bond := getDelegation(t, port, addr, operAddrs[0])
	require.Equal(t, amtDec, bond.Shares)

	delegatorDels := getDelegatorDelegations(t, port, addr)
	require.Len(t, delegatorDels, 1)
	require.Equal(t, amtDec, delegatorDels[0].Shares)

	// query all delegations to validator
	bonds := getValidatorDelegations(t, port, operAddrs[0])
	require.Len(t, bonds, 2)

	bondedValidators := getDelegatorValidators(t, port, addr)
	require.Len(t, bondedValidators, 1)
	require.Equal(t, operAddrs[0], bondedValidators[0].OperatorAddr)
	require.Equal(t, validator.DelegatorShares.Add(amtDec).String(), bondedValidators[0].DelegatorShares.String())

	bondedValidator := getDelegatorValidator(t, port, addr, operAddrs[0])
	require.Equal(t, operAddrs[0], bondedValidator.OperatorAddr)

	// testing unbonding
	unbondingTokens := sdk.TokensFromTendermintPower(30)
	resultTx = doUndelegate(t, port, name1, pw, addr, operAddrs[0], unbondingTokens, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	require.Equal(t, uint32(0), resultTx.Code)

	// sender should have not received any coins as the unbonding has only just begun
	acc = getAccount(t, port, addr)
	coins = acc.GetCoins()
	expectedBalance = expectedBalance.Minus(fees[0])
	require.True(t,
		expectedBalance.Amount.LT(coins.AmountOf(sdk.DefaultBondDenom)) ||
			expectedBalance.Amount.Equal(coins.AmountOf(sdk.DefaultBondDenom)),
		"should get tokens back from automatic withdrawal after an unbonding delegation",
	)
	expectedBalance = coins[0]

	// query tx
	txs = getTransactions(t, port,
		fmt.Sprintf("action=begin_unbonding&delegator=%s", addr),
		fmt.Sprintf("source-validator=%s", operAddrs[0]),
	)
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	ubd := getUnbondingDelegation(t, port, addr, operAddrs[0])
	require.Len(t, ubd.Entries, 1)
	require.Equal(t, delTokens.DivRaw(2), ubd.Entries[0].Balance)

	// test redelegation
	rdTokens := sdk.TokensFromTendermintPower(30)
	resultTx = doBeginRedelegation(t, port, name1, pw, addr, operAddrs[0], operAddrs[1], rdTokens, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	require.Equal(t, uint32(0), resultTx.Code)

	// verify balance after paying fees
	acc = getAccount(t, port, addr)
	expectedBalance = expectedBalance.Minus(fees[0])
	require.True(t,
		expectedBalance.Amount.LT(coins.AmountOf(sdk.DefaultBondDenom)) ||
			expectedBalance.Amount.Equal(coins.AmountOf(sdk.DefaultBondDenom)),
		"should get tokens back from automatic withdrawal after an unbonding delegation",
	)

	// query tx
	txs = getTransactions(t, port,
		fmt.Sprintf("action=begin_redelegate&delegator=%s", addr),
		fmt.Sprintf("source-validator=%s", operAddrs[0]),
		fmt.Sprintf("destination-validator=%s", operAddrs[1]),
	)
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	// query delegations, unbondings and redelegations from validator and delegator
	rdShares := sdk.NewDecFromInt(rdTokens)
	delegatorDels = getDelegatorDelegations(t, port, addr)
	require.Len(t, delegatorDels, 1)
	require.Equal(t, rdShares, delegatorDels[0].GetShares())

	redelegation := getRedelegations(t, port, addr, operAddrs[0], operAddrs[1])
	require.Len(t, redelegation, 1)
	require.Len(t, redelegation[0].Entries, 1)

	delegatorUbds := getDelegatorUnbondingDelegations(t, port, addr)
	require.Len(t, delegatorUbds, 1)
	require.Len(t, delegatorUbds[0].Entries, 1)
	require.Equal(t, rdTokens, delegatorUbds[0].Entries[0].Balance)

	delegatorReds := getRedelegations(t, port, addr, nil, nil)
	require.Len(t, delegatorReds, 1)
	require.Len(t, delegatorReds[0].Entries, 1)

	validatorUbds := getValidatorUnbondingDelegations(t, port, operAddrs[0])
	require.Len(t, validatorUbds, 1)
	require.Len(t, validatorUbds[0].Entries, 1)
	require.Equal(t, rdTokens, validatorUbds[0].Entries[0].Balance)

	validatorReds := getRedelegations(t, port, nil, operAddrs[0], nil)
	require.Len(t, validatorReds, 1)
	require.Len(t, validatorReds[0].Entries, 1)

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
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, seed := CreateAddr(t, name1, pw, kb)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
	defer cleanup()

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create SubmitProposal TX
	proposalTokens := sdk.TokensFromTendermintPower(5)
	resultTx := doSubmitProposal(t, port, seed, name1, pw, addr, proposalTokens, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.Code)

	var proposalID uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.Data, &proposalID)

	// verify balance
	acc = getAccount(t, port, addr)
	expectedBalance := initialBalance[0].Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(proposalTokens), acc.GetCoins().AmountOf(sdk.DefaultBondDenom))

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	proposer := getProposer(t, port, proposalID)
	require.Equal(t, addr.String(), proposer.Proposer)
	require.Equal(t, proposalID, proposer.ProposalID)
}

func TestDeposit(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, seed := CreateAddr(t, name1, pw, kb)
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
	defer cleanup()

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create SubmitProposal TX
	proposalTokens := sdk.TokensFromTendermintPower(5)
	resultTx := doSubmitProposal(t, port, seed, name1, pw, addr, proposalTokens, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.Code)

	var proposalID uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.Data, &proposalID)

	// verify balance
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	expectedBalance := initialBalance[0].Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(proposalTokens), coins.AmountOf(sdk.DefaultBondDenom))
	expectedBalance = coins[0]

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	// create SubmitProposal TX
	depositTokens := sdk.TokensFromTendermintPower(5)
	resultTx = doDeposit(t, port, seed, name1, pw, addr, proposalID, depositTokens, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// verify balance after deposit and fee
	acc = getAccount(t, port, addr)
	expectedBalance = expectedBalance.Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(depositTokens), acc.GetCoins().AmountOf(sdk.DefaultBondDenom))

	// query tx
	txs := getTransactions(t, port, fmt.Sprintf("action=deposit&depositor=%s", addr))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	// query proposal
	totalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromTendermintPower(10))}
	proposal = getProposal(t, port, proposalID)
	require.True(t, proposal.GetTotalDeposit().IsEqual(totalCoins))

	// query deposit
	deposit := getDeposit(t, port, proposalID, addr)
	require.True(t, deposit.Amount.IsEqual(totalCoins))
}

func TestVote(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, seed := CreateAddr(t, name1, pw, kb)
	cleanup, _, operAddrs, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
	defer cleanup()

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create SubmitProposal TX
	proposalTokens := sdk.TokensFromTendermintPower(10)
	resultTx := doSubmitProposal(t, port, seed, name1, pw, addr, proposalTokens, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.Code)

	var proposalID uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.Data, &proposalID)

	// verify balance
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	expectedBalance := initialBalance[0].Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(proposalTokens), coins.AmountOf(sdk.DefaultBondDenom))
	expectedBalance = coins[0]

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())
	require.Equal(t, gov.StatusVotingPeriod, proposal.GetStatus())

	// vote
	resultTx = doVote(t, port, seed, name1, pw, addr, proposalID, "Yes", fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// verify balance after vote and fee
	acc = getAccount(t, port, addr)
	coins = acc.GetCoins()
	expectedBalance = expectedBalance.Minus(fees[0])
	require.Equal(t, expectedBalance.Amount, coins.AmountOf(sdk.DefaultBondDenom))
	expectedBalance = coins[0]

	// query tx
	txs := getTransactions(t, port, fmt.Sprintf("action=vote&voter=%s", addr))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)

	vote := getVote(t, port, proposalID, addr)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, gov.OptionYes, vote.Option)

	tally := getTally(t, port, proposalID)
	require.Equal(t, sdk.ZeroInt(), tally.Yes, "tally should be 0 as the address is not bonded")

	// create bond TX
	delTokens := sdk.TokensFromTendermintPower(60)
	resultTx = doDelegate(t, port, name1, pw, addr, operAddrs[0], delTokens, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// verify balance
	acc = getAccount(t, port, addr)
	coins = acc.GetCoins()
	expectedBalance = expectedBalance.Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(delTokens), coins.AmountOf(sdk.DefaultBondDenom))
	expectedBalance = coins[0]

	tally = getTally(t, port, proposalID)
	require.Equal(t, delTokens, tally.Yes, "tally should be equal to the amount delegated")

	// change vote option
	resultTx = doVote(t, port, seed, name1, pw, addr, proposalID, "No", fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// verify balance
	acc = getAccount(t, port, addr)
	expectedBalance = expectedBalance.Minus(fees[0])
	require.Equal(t, expectedBalance.Amount, acc.GetCoins().AmountOf(sdk.DefaultBondDenom))

	tally = getTally(t, port, proposalID)
	require.Equal(t, sdk.ZeroInt(), tally.Yes, "tally should be 0 the user changed the option")
	require.Equal(t, delTokens, tally.No, "tally should be equal to the amount delegated")
}

func TestUnjail(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, _ := CreateAddr(t, name1, pw, kb)
	cleanup, valPubKeys, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
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
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addrs, seeds, names, passwords := CreateAddrs(t, kb, 2)

	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addrs[0], addrs[1]}, true)
	defer cleanup()

	depositParam := getDepositParam(t, port)
	halfMinDeposit := depositParam.MinDeposit.AmountOf(sdk.DefaultBondDenom).DivRaw(2)
	getVotingParam(t, port)
	getTallyingParam(t, port)

	// Addr1 proposes (and deposits) proposals #1 and #2
	resultTx := doSubmitProposal(t, port, seeds[0], names[0], passwords[0], addrs[0], halfMinDeposit, fees)
	var proposalID1 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.Data, &proposalID1)
	tests.WaitForHeight(resultTx.Height+1, port)

	resultTx = doSubmitProposal(t, port, seeds[0], names[0], passwords[0], addrs[0], halfMinDeposit, fees)
	var proposalID2 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.Data, &proposalID2)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 proposes (and deposits) proposals #3
	resultTx = doSubmitProposal(t, port, seeds[1], names[1], passwords[1], addrs[1], halfMinDeposit, fees)
	var proposalID3 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.Data, &proposalID3)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 deposits on proposals #2 & #3
	resultTx = doDeposit(t, port, seeds[1], names[1], passwords[1], addrs[1], proposalID2, halfMinDeposit, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	resultTx = doDeposit(t, port, seeds[1], names[1], passwords[1], addrs[1], proposalID3, halfMinDeposit, fees)
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
	depositTokens := sdk.TokensFromTendermintPower(1)
	resultTx = doDeposit(t, port, seeds[0], names[0], passwords[0], addrs[0], proposalID1, depositTokens, fees)
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
	resultTx = doVote(t, port, seeds[0], names[0], passwords[0], addrs[0], proposalID2, "Yes", fees)
	tests.WaitForHeight(resultTx.Height+1, port)
	resultTx = doVote(t, port, seeds[0], names[0], passwords[0], addrs[0], proposalID3, "Yes", fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 votes on proposal #3
	resultTx = doVote(t, port, seeds[1], names[1], passwords[1], addrs[1], proposalID3, "Yes", fees)
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

func TestSlashingGetParams(t *testing.T) {
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{}, true)
	defer cleanup()

	res, body := Request(t, port, "GET", "/slashing/parameters", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var params slashing.Params
	err := cdc.UnmarshalJSON([]byte(body), &params)
	require.NoError(t, err)
}

func TestDistributionGetParams(t *testing.T) {
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{}, true)
	defer cleanup()

	res, body := Request(t, port, "GET", "/distribution/parameters", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &dclcommon.PrettyParams{}))
}

func TestDistributionFlow(t *testing.T) {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(t, ""))
	require.NoError(t, err)
	addr, seed := CreateAddr(t, name1, pw, kb)
	cleanup, _, valAddrs, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true)
	defer cleanup()

	valAddr := valAddrs[0]
	operAddr := sdk.AccAddress(valAddr)

	var rewards sdk.DecCoins
	res, body := Request(t, port, "GET", fmt.Sprintf("/distribution/outstanding_rewards"), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &rewards))

	var valDistInfo distrrest.ValidatorDistInfo
	res, body = Request(t, port, "GET", "/distribution/validators/"+valAddr.String(), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &valDistInfo))
	require.Equal(t, valDistInfo.OperatorAddress.String(), sdk.AccAddress(valAddr).String())

	// Delegate some coins
	delTokens := sdk.TokensFromTendermintPower(60)
	resultTx := doDelegate(t, port, name1, pw, addr, valAddr, delTokens, fees)
	tests.WaitForHeight(resultTx.Height+1, port)
	require.Equal(t, uint32(0), resultTx.Code)

	// send some coins
	_, resultTx = doTransfer(t, port, seed, name1, memo, pw, addr, fees)
	tests.WaitForHeight(resultTx.Height+5, port)
	require.Equal(t, uint32(0), resultTx.Code)

	// Query outstanding rewards changed
	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/outstanding_rewards"), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &rewards))

	// Query validator distribution info
	res, body = Request(t, port, "GET", "/distribution/validators/"+valAddr.String(), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &valDistInfo))

	// Query validator's rewards
	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/validators/%s/rewards", valAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &rewards))

	// Query self-delegation
	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/delegators/%s/rewards/%s", operAddr, valAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &rewards))

	// Query delegation
	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/delegators/%s/rewards/%s", addr, valAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &rewards))

	// Query delegator's rewards total
	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/delegators/%s/rewards", operAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &rewards))

	// Query delegator's withdrawal address
	var withdrawAddr string
	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/delegators/%s/withdraw_address", operAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &withdrawAddr))
	require.Equal(t, operAddr.String(), withdrawAddr)

	// Withdraw delegator's rewards
	resultTx = doWithdrawDelegatorAllRewards(t, port, seed, name1, pw, addr, fees)
	require.Equal(t, uint32(0), resultTx.Code)
}
