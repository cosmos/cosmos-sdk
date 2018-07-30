package rest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tmlibs/common"
)

// import (
// 	"encoding/hex"
// 	"fmt"
// 	"net/http"
// 	"regexp"
// 	"testing"
//
// 	"github.com/spf13/viper"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
//
// 	cryptoKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
// 	abci "github.com/tendermint/tendermint/abci/types"
// 	"github.com/tendermint/tendermint/libs/common"
// 	p2p "github.com/tendermint/tendermint/p2p"
// 	ctypes "github.com/tendermint/tendermint/rpc/core/types"
//
// 	client "github.com/cosmos/cosmos-sdk/client"
// 	keys "github.com/cosmos/cosmos-sdk/client/keys"
// 	rpc "github.com/cosmos/cosmos-sdk/client/rpc"
// 	tests "github.com/cosmos/cosmos-sdk/tests"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/cosmos-sdk/wire"
// 	"github.com/cosmos/cosmos-sdk/x/slashing"
// )

func TestStakingTxs(t *testing.T) {
	name, password := "test", "1234567890"
	addr, seed := CreateAddr(t, "test", password, GetKeyBase(t))
	cleanup, _, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr})
	defer cleanup()

	// query empty
	res, body = Request(t, port, "GET", fmt.Sprintf("stake/delegators/'%s'/txs", "cosmosaccaddr1jawd35d9aq4u76sr3fjalmcqc8hqygs9gtnmv3"), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Equal(t, "[]", body)

	// query empty bonding txs
	res, body = Request(t, port, "GET", fmt.Sprintf("stake/delegators/'%s'/txs?type=bond", "cosmosaccaddr1jawd35d9aq4u76sr3fjalmcqc8hqygs9gtnmv3"), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Equal(t, "[]", body)

	// query empty unbonding txs
	res, body = Request(t, port, "GET", fmt.Sprintf("stake/delegators/'%s'/txs?type=unbond", "cosmosaccaddr1jawd35d9aq4u76sr3fjalmcqc8hqygs9gtnmv3"), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Equal(t, "[]", body)

	// query empty redelegation txs
	res, body = Request(t, port, "GET", fmt.Sprintf("stake/delegators/'%s'/txs?type=redelegate", "cosmosaccaddr1jawd35d9aq4u76sr3fjalmcqc8hqygs9gtnmv3"), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Equal(t, "[]", body)

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
	require.Equal(t, "60/1", bond.Shares.String())

	// query bonding tx
	res, body = Request(t, port, "GET", fmt.Sprintf("stake/delegators/'%s'/txs", "cosmosaccaddr1jawd35d9aq4u76sr3fjalmcqc8hqygs9gtnmv3"), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Equal(t, "[]", body)

	res, body = Request(t, port, "GET", fmt.Sprintf("stake/delegators/'%s'/validators/'%s'/txs", "cosmosaccaddr1jawd35d9aq4u76sr3fjalmcqc8hqygs9gtnmv3"), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Equal(t, "[]", body)

	//////////////////////
	// testing unbonding

	// create unbond TX
	resultTx = doBeginUnbonding(t, port, seed, name, password, addr, validator1Owner)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query validator
	bond = getDelegation(t, port, addr, validator1Owner)
	require.Equal(t, "30/1", bond.Shares.String())

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.CheckTx.Code)
	require.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// should the sender should have not received any coins as the unbonding has only just begun
	// query sender
	acc = getAccount(t, port, addr)
	coins = acc.GetCoins()
	require.Equal(t, int64(40), coins.AmountOf("steak").Int64())

	// TODO add redelegation, need more complex capabilities such to mock context and
}

func getDelegation(t *testing.T, port string, delegatorAddr, validatorAddr sdk.AccAddress) stake.Delegation {

	// get the account to get the sequence
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/%s/delegation/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var bond stake.Delegation
	err := cdc.UnmarshalJSON([]byte(body), &bond)
	require.Nil(t, err)
	return bond
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
	res, body := Request(t, port, "POST", "/stake/delegations", jsonStr)
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
	res, body := Request(t, port, "POST", "/stake/delegations", jsonStr)
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
	res, body := Request(t, port, "POST", "/stake/delegations", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func getValidatorsREST(t *testing.T, port string) []stake.BechValidator {
	// get the account to get the sequence
	res, body := Request(t, port, "GET", "/stake/validators", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var validators []stake.BechValidator
	err := cdc.UnmarshalJSON([]byte(body), &validators)
	require.Nil(t, err)
	return validators
}
