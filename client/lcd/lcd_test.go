package lcd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	client "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"

	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
	tests "github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
)

const (
	name1 = "test1"
	name2 = "test2"
	name3 = "test3"
	memo  = "LCD test tx"
	pw    = app.DefaultKeyPass
	altPw = "12345678901"
)

var fees = sdk.Coins{sdk.NewInt64Coin(stakeTypes.DefaultBondDenom, 5)}

func init() {
	mintkey.BcryptSecurityParameter = 1
	version.Version = os.Getenv("VERSION")
}

func TestKeys(t *testing.T) {
	addr, _ := CreateAddr(t, name1, pw, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// get new seed
	seed := getKeysSeed(t, port)

	// recover key
	doRecoverKey(t, port, name2, pw, seed)

	// add key
	resp := doKeysPost(t, port, name3, pw, seed)

	addrBech32 := addr.String()
	addr2Bech32 := resp.Address
	_, err := sdk.AccAddressFromBech32(addr2Bech32)
	require.NoError(t, err, "Failed to return a correct bech32 address")

	// test if created account is the correct account
	expectedInfo, _ := GetKeyBase(t).CreateKey(name3, seed, pw)
	expectedAccount := sdk.AccAddress(expectedInfo.GetPubKey().Address().Bytes())
	require.Equal(t, expectedAccount.String(), addr2Bech32)

	// existing keys
	keys := getKeys(t, port)
	require.Equal(t, name1, keys[0].Name, "Did not serve keys name correctly")
	require.Equal(t, addrBech32, keys[0].Address, "Did not serve keys Address correctly")
	require.Equal(t, name2, keys[1].Name, "Did not serve keys name correctly")
	require.Equal(t, addr2Bech32, keys[1].Address, "Did not serve keys Address correctly")

	// select key
	key := getKey(t, port, name3)
	require.Equal(t, name3, key.Name, "Did not serve keys name correctly")
	require.Equal(t, addr2Bech32, key.Address, "Did not serve keys Address correctly")

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
	getNodeInfo(t, port)
	getSyncStatus(t, port, false)
}

func TestBlock(t *testing.T) {
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()
	getBlock(t, port, -1, false)
	getBlock(t, port, 2, false)
	getBlock(t, port, 100000000, true)
}

func TestValidators(t *testing.T) {
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()
	resultVals := getValidatorSets(t, port, -1, false)
	require.Contains(t, resultVals.Validators[0].Address.String(), "cosmosvaloper")
	require.Contains(t, resultVals.Validators[0].PubKey, "cosmosvalconspub")
	getValidatorSets(t, port, 2, false)
	getValidatorSets(t, port, 10000000, true)
}

func TestCoinSend(t *testing.T) {
	addr, seed := CreateAddr(t, name1, pw, GetKeyBase(t))
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
	receiveAddr, resultTx := doTransfer(t, port, seed, name1, memo, pw, addr, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	expectedBalance := initialBalance[0].Minus(fees[0])

	require.Equal(t, stakeTypes.DefaultBondDenom, coins[0].Denom)
	require.Equal(t, expectedBalance.Amount.SubRaw(1), coins[0].Amount)
	expectedBalance = coins[0]

	// query receiver
	acc2 := getAccount(t, port, receiveAddr)
	coins2 := acc2.GetCoins()
	require.Equal(t, stakeTypes.DefaultBondDenom, coins2[0].Denom)
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
	res, body, _ = doTransferWithGas(t, port, seed, name1, memo, pw, addr, "10000", 1.0, true, false, fees)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var responseBody struct {
		GasEstimate int64 `json:"gas_estimate"`
	}
	require.Nil(t, json.Unmarshal([]byte(body), &responseBody))

	acc = getAccount(t, port, addr)
	require.Equal(t, expectedBalance.Amount, acc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom))

	res, body, _ = doTransferWithGas(t, port, seed, name1, memo, pw, addr,
		fmt.Sprintf("%d", responseBody.GasEstimate), 1.0, false, false, fees)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	tests.WaitForHeight(resultTx.Height+1, port)
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	acc = getAccount(t, port, addr)
	expectedBalance = expectedBalance.Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.SubRaw(1), acc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom))
}

func TestCoinSendGenerateSignAndBroadcast(t *testing.T) {
	addr, seed := CreateAddr(t, name1, pw, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()
	acc := getAccount(t, port, addr)

	// generate TX
	res, body, _ := doTransferWithGas(t, port, seed, name1, memo, "", addr, client.GasFlagAuto, 1, false, true, fees)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var msg auth.StdTx
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &msg))
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, msg.Msgs[0].Route(), "bank")
	require.Equal(t, msg.Msgs[0].GetSigners(), []sdk.AccAddress{addr})
	require.Equal(t, 0, len(msg.Signatures))
	require.Equal(t, memo, msg.Memo)

	gasEstimate := int64(msg.Fee.Gas)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	// sign tx
	var signedMsg auth.StdTx

	payload := authrest.SignBody{
		Tx: msg,
		BaseReq: utils.NewBaseReq(
			name1, pw, "", viper.GetString(client.FlagChainID), "", "", accnum, sequence, nil, false, false,
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
	var resultTx ctypes.ResultBroadcastTxCommit
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &resultTx))
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)
	require.Equal(t, gasEstimate, resultTx.DeliverTx.GasWanted)
}

func TestTxs(t *testing.T) {
	addr, seed := CreateAddr(t, name1, pw, GetKeyBase(t))
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
	receiveAddr, resultTx := doTransfer(t, port, seed, name1, memo, pw, addr, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx is queryable
	tx := getTransaction(t, port, resultTx.Hash.String())
	require.Equal(t, resultTx.Hash, tx.Hash)

	// query sender
	txs = getTransactions(t, port, fmt.Sprintf("sender=%s", addr.String()))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)
	fmt.Println(txs[0])

	// query recipient
	txs = getTransactions(t, port, fmt.Sprintf("recipient=%s", receiveAddr.String()))
	require.Len(t, txs, 1)
	require.Equal(t, resultTx.Height, txs[0].Height)
}

func TestPoolParamsQuery(t *testing.T) {
	addr, _ := CreateAddr(t, name1, pw, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	defaultParams := stake.DefaultParams()

	params := getStakeParams(t, port)
	require.True(t, defaultParams.Equal(params))

	pool := getStakePool(t, port)

	initialPool := stake.InitialPool()
	initialPool.LooseTokens = initialPool.LooseTokens.Add(sdk.NewDec(100))
	initialPool.BondedTokens = initialPool.BondedTokens.Add(sdk.NewDec(100))     // Delegate tx on GaiaAppGenState
	initialPool.LooseTokens = initialPool.LooseTokens.Add(sdk.NewDec(int64(50))) // freeFermionsAcc = 50 on GaiaAppGenState

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
	addr, _ := CreateAddr(t, name1, pw, GetKeyBase(t))

	cleanup, valPubKeys, operAddrs, port := InitializeTestLCD(t, 2, []sdk.AccAddress{addr})
	defer cleanup()

	require.Equal(t, 2, len(valPubKeys))
	require.Equal(t, 2, len(operAddrs))

	amt := sdk.NewDec(60)
	validator := getValidator(t, port, operAddrs[0])

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create bond TX
	resultTx := doDelegate(t, port, name1, pw, addr, operAddrs[0], 60, fees)
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

	// verify balance
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	expectedBalance := initialBalance[0].Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.SubRaw(60), coins.AmountOf(stakeTypes.DefaultBondDenom))
	expectedBalance = coins[0]

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
	resultTx = doBeginUnbonding(t, port, name1, pw, addr, operAddrs[0], 30, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// sender should have not received any coins as the unbonding has only just begun
	acc = getAccount(t, port, addr)
	coins = acc.GetCoins()
	expectedBalance = expectedBalance.Minus(fees[0])
	require.True(t,
		expectedBalance.Amount.LT(coins.AmountOf(stakeTypes.DefaultBondDenom)) ||
			expectedBalance.Amount.Equal(coins.AmountOf(stakeTypes.DefaultBondDenom)),
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

	unbonding := getUndelegation(t, port, addr, operAddrs[0])
	require.Equal(t, int64(30), unbonding.Balance.Amount.Int64())

	// test redelegation
	resultTx = doBeginRedelegation(t, port, name1, pw, addr, operAddrs[0], operAddrs[1], 30, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// verify balance after paying fees
	acc = getAccount(t, port, addr)
	expectedBalance = expectedBalance.Minus(fees[0])
	require.True(t,
		expectedBalance.Amount.LT(coins.AmountOf(stakeTypes.DefaultBondDenom)) ||
			expectedBalance.Amount.Equal(coins.AmountOf(stakeTypes.DefaultBondDenom)),
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
	delegatorDels = getDelegatorDelegations(t, port, addr)
	require.Len(t, delegatorDels, 1)
	require.Equal(t, "30.0000000000", delegatorDels[0].GetShares().String())

	redelegation := getRedelegations(t, port, addr, operAddrs[0], operAddrs[1])
	require.Len(t, redelegation, 1)
	require.Equal(t, "30", redelegation[0].Balance.Amount.String())

	delegatorUbds := getDelegatorUnbondingDelegations(t, port, addr)
	require.Len(t, delegatorUbds, 1)
	require.Equal(t, "30", delegatorUbds[0].Balance.Amount.String())

	delegatorReds := getRedelegations(t, port, addr, nil, nil)
	require.Len(t, delegatorReds, 1)
	require.Equal(t, "30", delegatorReds[0].Balance.Amount.String())

	validatorUbds := getValidatorUnbondingDelegations(t, port, operAddrs[0])
	require.Len(t, validatorUbds, 1)
	require.Equal(t, "30", validatorUbds[0].Balance.Amount.String())

	validatorReds := getRedelegations(t, port, nil, operAddrs[0], nil)
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
	addr, seed := CreateAddr(t, name1, pw, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name1, pw, addr, 5, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID)

	// verify balance
	acc = getAccount(t, port, addr)
	expectedBalance := initialBalance[0].Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.SubRaw(5), acc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom))

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	proposer := getProposer(t, port, proposalID)
	require.Equal(t, addr.String(), proposer.Proposer)
	require.Equal(t, proposalID, proposer.ProposalID)
}

func TestDeposit(t *testing.T) {
	addr, seed := CreateAddr(t, name1, pw, GetKeyBase(t))
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name1, pw, addr, 5, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID)

	// verify balance
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	expectedBalance := initialBalance[0].Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.SubRaw(5), coins.AmountOf(stakeTypes.DefaultBondDenom))
	expectedBalance = coins[0]

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	// create SubmitProposal TX
	resultTx = doDeposit(t, port, seed, name1, pw, addr, proposalID, 5, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// verify balance after deposit and fee
	acc = getAccount(t, port, addr)
	expectedBalance = expectedBalance.Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.SubRaw(5), acc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom))

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
	addr, seed := CreateAddr(t, name1, pw, GetKeyBase(t))
	cleanup, _, operAddrs, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	acc := getAccount(t, port, addr)
	initialBalance := acc.GetCoins()

	// create SubmitProposal TX
	resultTx := doSubmitProposal(t, port, seed, name1, pw, addr, 10, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	var proposalID uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID)

	// verify balance
	acc = getAccount(t, port, addr)
	coins := acc.GetCoins()
	expectedBalance := initialBalance[0].Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.SubRaw(10), coins.AmountOf(stakeTypes.DefaultBondDenom))
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
	require.Equal(t, expectedBalance.Amount, coins.AmountOf(stakeTypes.DefaultBondDenom))
	expectedBalance = coins[0]

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
	resultTx = doDelegate(t, port, name1, pw, addr, operAddrs[0], 60, fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// verify balance
	acc = getAccount(t, port, addr)
	coins = acc.GetCoins()
	expectedBalance = expectedBalance.Minus(fees[0])
	require.Equal(t, expectedBalance.Amount.SubRaw(60), coins.AmountOf(stakeTypes.DefaultBondDenom))
	expectedBalance = coins[0]

	tally = getTally(t, port, proposalID)
	require.Equal(t, sdk.NewDec(60), tally.Yes, "tally should be equal to the amount delegated")

	// change vote option
	resultTx = doVote(t, port, seed, name1, pw, addr, proposalID, "No", fees)
	tests.WaitForHeight(resultTx.Height+1, port)

	// verify balance
	acc = getAccount(t, port, addr)
	expectedBalance = expectedBalance.Minus(fees[0])
	require.Equal(t, expectedBalance.Amount, acc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom))

	tally = getTally(t, port, proposalID)
	require.Equal(t, sdk.ZeroDec(), tally.Yes, "tally should be 0 the user changed the option")
	require.Equal(t, sdk.NewDec(60), tally.No, "tally should be equal to the amount delegated")
}

func TestUnjail(t *testing.T) {
	addr, _ := CreateAddr(t, name1, pw, GetKeyBase(t))
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
	resultTx := doSubmitProposal(t, port, seeds[0], names[0], passwords[0], addrs[0], halfMinDeposit, fees)
	var proposalID1 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID1)
	tests.WaitForHeight(resultTx.Height+1, port)

	resultTx = doSubmitProposal(t, port, seeds[0], names[0], passwords[0], addrs[0], halfMinDeposit, fees)
	var proposalID2 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID2)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 proposes (and deposits) proposals #3
	resultTx = doSubmitProposal(t, port, seeds[1], names[1], passwords[1], addrs[1], halfMinDeposit, fees)
	var proposalID3 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(resultTx.DeliverTx.GetData(), &proposalID3)
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
	resultTx = doDeposit(t, port, seeds[0], names[0], passwords[0], addrs[0], proposalID1, 1, fees)
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
	cleanup, _, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{})
	defer cleanup()

	res, body := Request(t, port, "GET", "/slashing/parameters", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var params slashing.Params
	err := cdc.UnmarshalJSON([]byte(body), &params)
	require.NoError(t, err)
}
