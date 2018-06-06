package lcd

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	cryptoKeys "github.com/tendermint/go-crypto/keys"
	tmcfg "github.com/tendermint/tendermint/config"
	nm "github.com/tendermint/tendermint/node"
	p2p "github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/proxy"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmrpc "github.com/tendermint/tendermint/rpc/lib/server"
	tmtypes "github.com/tendermint/tendermint/types"
	pvm "github.com/tendermint/tendermint/types/priv_validator"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	client "github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	rpc "github.com/cosmos/cosmos-sdk/client/rpc"
	gapp "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/server"
	tests "github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"
	stakerest "github.com/cosmos/cosmos-sdk/x/stake/client/rest"
)

var (
	coinDenom  = "steak"
	coinAmount = int64(10000000)

	validatorAddr1Hx = ""
	validatorAddr2Hx = ""
	validatorAddr1   = ""
	validatorAddr2   = ""

	// XXX bad globals
	name     = "test"
	password = "0123456789"
	port     string
	seed     string
	sendAddr string
)

func TestKeys(t *testing.T) {

	// empty keys
	// XXX: the test comes with a key setup
	/*
		res, body := request(t, port, "GET", "/keys", nil)
		require.Equal(t, http.StatusOK, res.StatusCode, body)
		assert.Equal(t, "[]", body, "Expected an empty array")
	*/

	// get seed
	res, body := request(t, port, "GET", "/keys/seed", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	newSeed := body
	reg, err := regexp.Compile(`([a-z]+ ){12}`)
	require.Nil(t, err)
	match := reg.MatchString(seed)
	assert.True(t, match, "Returned seed has wrong foramt", seed)

	newName := "test_newname"
	newPassword := "0987654321"

	// add key
	var jsonStr = []byte(fmt.Sprintf(`{"name":"test_fail", "password":"%s"}`, password))
	res, body = request(t, port, "POST", "/keys", jsonStr)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Account creation should require a seed")

	jsonStr = []byte(fmt.Sprintf(`{"name":"%s", "password":"%s", "seed": "%s"}`, newName, newPassword, newSeed))
	res, body = request(t, port, "POST", "/keys", jsonStr)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	addr := body
	assert.Len(t, addr, 40, "Returned address has wrong format", addr)

	// existing keys
	res, body = request(t, port, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m [2]keys.KeyOutput
	err = cdc.UnmarshalJSON([]byte(body), &m)
	require.Nil(t, err)

	addrAcc, _ := sdk.GetAccAddressHex(addr)
	addrBech32, _ := sdk.Bech32ifyAcc(addrAcc)

	assert.Equal(t, name, m[0].Name, "Did not serve keys name correctly")
	assert.Equal(t, sendAddr, m[0].Address, "Did not serve keys Address correctly")
	assert.Equal(t, newName, m[1].Name, "Did not serve keys name correctly")
	assert.Equal(t, addrBech32, m[1].Address, "Did not serve keys Address correctly")

	// select key
	keyEndpoint := fmt.Sprintf("/keys/%s", newName)
	res, body = request(t, port, "GET", keyEndpoint, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m2 keys.KeyOutput
	err = cdc.UnmarshalJSON([]byte(body), &m2)
	require.Nil(t, err)

	assert.Equal(t, newName, m2.Name, "Did not serve keys name correctly")
	assert.Equal(t, addrBech32, m2.Address, "Did not serve keys Address correctly")

	// update key
	jsonStr = []byte(fmt.Sprintf(`{"old_password":"%s", "new_password":"12345678901"}`, newPassword))
	res, body = request(t, port, "PUT", keyEndpoint, jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// here it should say unauthorized as we changed the password before
	res, body = request(t, port, "PUT", keyEndpoint, jsonStr)
	require.Equal(t, http.StatusUnauthorized, res.StatusCode, body)

	// delete key
	jsonStr = []byte(`{"password":"12345678901"}`)
	res, body = request(t, port, "DELETE", keyEndpoint, jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
}

func TestVersion(t *testing.T) {

	// node info
	res, body := request(t, port, "GET", "/version", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	reg, err := regexp.Compile(`\d+\.\d+\.\d+(-dev)?`)
	require.Nil(t, err)
	match := reg.MatchString(body)
	assert.True(t, match, body)
}

func TestNodeStatus(t *testing.T) {

	// node info
	res, body := request(t, port, "GET", "/node_info", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var nodeInfo p2p.NodeInfo
	err := cdc.UnmarshalJSON([]byte(body), &nodeInfo)
	require.Nil(t, err, "Couldn't parse node info")

	assert.NotEqual(t, p2p.NodeInfo{}, nodeInfo, "res: %v", res)

	// syncing
	res, body = request(t, port, "GET", "/syncing", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// we expect that there is no other node running so the syncing state is "false"
	// we c
	assert.Equal(t, "false", body)
}

func TestBlock(t *testing.T) {

	tests.WaitForHeight(2, port)

	var resultBlock ctypes.ResultBlock

	res, body := request(t, port, "GET", "/blocks/latest", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultBlock)
	require.Nil(t, err, "Couldn't parse block")

	assert.NotEqual(t, ctypes.ResultBlock{}, resultBlock)

	// --

	res, body = request(t, port, "GET", "/blocks/1", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = json.Unmarshal([]byte(body), &resultBlock)
	require.Nil(t, err, "Couldn't parse block")

	assert.NotEqual(t, ctypes.ResultBlock{}, resultBlock)

	// --

	res, body = request(t, port, "GET", "/blocks/1000000000", nil)
	require.Equal(t, http.StatusNotFound, res.StatusCode, body)
}

func TestValidators(t *testing.T) {

	var resultVals rpc.ResultValidatorsOutput

	res, body := request(t, port, "GET", "/validatorsets/latest", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultVals)
	require.Nil(t, err, "Couldn't parse validatorset")

	assert.NotEqual(t, rpc.ResultValidatorsOutput{}, resultVals)

	assert.Contains(t, resultVals.Validators[0].Address, "cosmosvaladdr")
	assert.Contains(t, resultVals.Validators[0].PubKey, "cosmosvalpub")

	// --

	res, body = request(t, port, "GET", "/validatorsets/1", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultVals)
	require.Nil(t, err, "Couldn't parse validatorset")

	assert.NotEqual(t, rpc.ResultValidatorsOutput{}, resultVals)

	// --

	res, body = request(t, port, "GET", "/validatorsets/1000000000", nil)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestCoinSend(t *testing.T) {
	bz, _ := hex.DecodeString("8FA6AB57AD6870F6B5B2E57735F38F2F30E73CB6")
	someFakeAddr, _ := sdk.Bech32ifyAcc(bz)

	// query empty
	res, body := request(t, port, "GET", "/accounts/"+someFakeAddr, nil)
	require.Equal(t, http.StatusNoContent, res.StatusCode, body)

	acc := getAccount(t, sendAddr)
	initialBalance := acc.GetCoins()

	// create TX
	receiveAddr, resultTx := doSend(t, port, seed)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was commited
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc = getAccount(t, sendAddr)
	coins := acc.GetCoins()
	mycoins := coins[0]
	assert.Equal(t, coinDenom, mycoins.Denom)
	assert.Equal(t, initialBalance[0].Amount-1, mycoins.Amount)

	// query receiver
	acc = getAccount(t, receiveAddr)
	coins = acc.GetCoins()
	mycoins = coins[0]
	assert.Equal(t, coinDenom, mycoins.Denom)
	assert.Equal(t, int64(1), mycoins.Amount)
}

func TestIBCTransfer(t *testing.T) {

	acc := getAccount(t, sendAddr)
	initialBalance := acc.GetCoins()

	// create TX
	resultTx := doIBCTransfer(t, port, seed)

	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was commited
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc = getAccount(t, sendAddr)
	coins := acc.GetCoins()
	mycoins := coins[0]
	assert.Equal(t, coinDenom, mycoins.Denom)
	assert.Equal(t, initialBalance[0].Amount-1, mycoins.Amount)

	// TODO: query ibc egress packet state
}

func TestTxs(t *testing.T) {

	// TODO: re-enable once we can get txs by tag

	// query wrong
	// res, body := request(t, port, "GET", "/txs", nil)
	// require.Equal(t, http.StatusBadRequest, res.StatusCode, body)

	// query empty
	// res, body = request(t, port, "GET", fmt.Sprintf("/txs?tag=coin.sender='%s'", "8FA6AB57AD6870F6B5B2E57735F38F2F30E73CB6"), nil)
	// require.Equal(t, http.StatusOK, res.StatusCode, body)

	// assert.Equal(t, "[]", body)

	// create TX
	_, resultTx := doSend(t, port, seed)

	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx is findable
	res, body := request(t, port, "GET", fmt.Sprintf("/txs/%s", resultTx.Hash), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// // query sender
	// res, body = request(t, port, "GET", fmt.Sprintf("/txs?tag=coin.sender='%s'", addr), nil)
	// require.Equal(t, http.StatusOK, res.StatusCode, body)

	// assert.NotEqual(t, "[]", body)

	// // query receiver
	// res, body = request(t, port, "GET", fmt.Sprintf("/txs?tag=coin.receiver='%s'", receiveAddr), nil)
	// require.Equal(t, http.StatusOK, res.StatusCode, body)

	// assert.NotEqual(t, "[]", body)
}

func TestValidatorsQuery(t *testing.T) {
	validators := getValidators(t)
	assert.Equal(t, len(validators), 2)

	// make sure all the validators were found (order unknown because sorted by owner addr)
	foundVal1, foundVal2 := false, false
	if validators[0].Owner == validatorAddr1 || validators[1].Owner == validatorAddr1 {
		foundVal1 = true
	}
	if validators[0].Owner == validatorAddr2 || validators[1].Owner == validatorAddr2 {
		foundVal2 = true
	}
	assert.True(t, foundVal1, "validatorAddr1 %v, owner1 %v, owner2 %v", validatorAddr1, validators[0].Owner, validators[1].Owner)
	assert.True(t, foundVal2, "validatorAddr2 %v, owner1 %v, owner2 %v", validatorAddr2, validators[0].Owner, validators[1].Owner)
}

func TestBond(t *testing.T) {

	// create bond TX
	resultTx := doBond(t, port, seed)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was commited
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc := getAccount(t, sendAddr)
	coins := acc.GetCoins()
	assert.Equal(t, int64(87), coins.AmountOf(coinDenom))

	// query candidate
	bond := getDelegation(t, sendAddr, validatorAddr1)
	assert.Equal(t, "10/1", bond.Shares.String())
}

func TestUnbond(t *testing.T) {

	// create unbond TX
	resultTx := doUnbond(t, port, seed)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was commited
	assert.Equal(t, uint32(0), resultTx.CheckTx.Code)
	assert.Equal(t, uint32(0), resultTx.DeliverTx.Code)

	// query sender
	acc := getAccount(t, sendAddr)
	coins := acc.GetCoins()
	assert.Equal(t, int64(98), coins.AmountOf(coinDenom))

	// query candidate
	bond := getDelegation(t, sendAddr, validatorAddr1)
	assert.Equal(t, "9/1", bond.Shares.String())
}

//__________________________________________________________
// helpers

// strt TM and the LCD in process, listening on their respective sockets
func startTMAndLCD() (*nm.Node, net.Listener, error) {

	dir, err := ioutil.TempDir("", "lcd_test")
	if err != nil {
		return nil, nil, err
	}
	viper.Set(cli.HomeFlag, dir)
	kb, err := keys.GetKeyBase() // dbm.NewMemDB()) // :(
	if err != nil {
		return nil, nil, err
	}

	config := GetConfig()
	config.Consensus.TimeoutCommit = 1000
	config.Consensus.SkipTimeoutCommit = false

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger = log.NewFilter(logger, log.AllowError())
	privValidatorFile := config.PrivValidatorFile()
	privVal := pvm.LoadOrGenFilePV(privValidatorFile)
	db := dbm.NewMemDB()
	app := gapp.NewGaiaApp(logger, db)
	cdc = gapp.MakeCodec() // XXX

	genesisFile := config.GenesisFile()
	genDoc, err := tmtypes.GenesisDocFromFile(genesisFile)
	if err != nil {
		return nil, nil, err
	}

	genDoc.Validators = append(genDoc.Validators,
		tmtypes.GenesisValidator{
			PubKey: crypto.GenPrivKeyEd25519().PubKey(),
			Power:  1,
			Name:   "val",
		},
	)

	pk1 := genDoc.Validators[0].PubKey
	pk2 := genDoc.Validators[1].PubKey
	validatorAddr1Hx = hex.EncodeToString(pk1.Address())
	validatorAddr2Hx = hex.EncodeToString(pk2.Address())
	validatorAddr1, _ = sdk.Bech32ifyVal(pk1.Address())
	validatorAddr2, _ = sdk.Bech32ifyVal(pk2.Address())

	// NOTE it's bad practice to reuse pk address for the owner address but doing in the
	// test for simplicity
	var appGenTxs [2]json.RawMessage
	appGenTxs[0], _, _, err = gapp.GaiaAppGenTxNF(cdc, pk1, pk1.Address(), "test_val1", true)
	if err != nil {
		return nil, nil, err
	}
	appGenTxs[1], _, _, err = gapp.GaiaAppGenTxNF(cdc, pk2, pk2.Address(), "test_val2", true)
	if err != nil {
		return nil, nil, err
	}

	genesisState, err := gapp.GaiaAppGenState(cdc, appGenTxs[:])
	if err != nil {
		return nil, nil, err
	}

	// add the sendAddr to genesis
	var info cryptoKeys.Info
	info, seed, err = kb.Create(name, password, cryptoKeys.AlgoEd25519) // XXX global seed
	if err != nil {
		return nil, nil, err
	}
	sendAddrHex, _ := sdk.GetAccAddressHex(info.PubKey.Address().String())
	sendAddr, _ = sdk.Bech32ifyAcc(sendAddrHex) // XXX global
	accAuth := auth.NewBaseAccountWithAddress(info.PubKey.Address())
	accAuth.Coins = sdk.Coins{{"steak", 100}}
	acc := gapp.NewGenesisAccount(&accAuth)
	genesisState.Accounts = append(genesisState.Accounts, acc)

	appState, err := wire.MarshalJSONIndent(cdc, genesisState)
	if err != nil {
		return nil, nil, err
	}
	genDoc.AppStateJSON = appState

	// LCD listen address
	var listenAddr string
	listenAddr, port, err = server.FreeTCPAddr()
	if err != nil {
		return nil, nil, err
	}

	// XXX: need to set this so LCD knows the tendermint node address!
	viper.Set(client.FlagNode, config.RPC.ListenAddress)
	viper.Set(client.FlagChainID, genDoc.ChainID)

	node, err := startTM(config, logger, genDoc, privVal, app)
	if err != nil {
		return nil, nil, err
	}
	lcd, err := startLCD(logger, listenAddr, cdc)
	if err != nil {
		return nil, nil, err
	}

	tests.WaitForStart(port)

	return node, lcd, nil
}

// Create & start in-process tendermint node with memdb
// and in-process abci application.
// TODO: need to clean up the WAL dir or enable it to be not persistent
func startTM(cfg *tmcfg.Config, logger log.Logger, genDoc *tmtypes.GenesisDoc, privVal tmtypes.PrivValidator, app abci.Application) (*nm.Node, error) {
	genDocProvider := func() (*tmtypes.GenesisDoc, error) { return genDoc, nil }
	dbProvider := func(*nm.DBContext) (dbm.DB, error) { return dbm.NewMemDB(), nil }
	n, err := nm.NewNode(cfg,
		privVal,
		proxy.NewLocalClientCreator(app),
		genDocProvider,
		dbProvider,
		logger.With("module", "node"))
	if err != nil {
		return nil, err
	}

	err = n.Start()
	if err != nil {
		return nil, err
	}

	// wait for rpc
	tests.WaitForRPC(GetConfig().RPC.ListenAddress)

	logger.Info("Tendermint running!")
	return n, err
}

// start the LCD. note this blocks!
func startLCD(logger log.Logger, listenAddr string, cdc *wire.Codec) (net.Listener, error) {
	handler := createHandler(cdc)
	return tmrpc.StartHTTPServer(listenAddr, handler, logger)
}

func request(t *testing.T, port, method, path string, payload []byte) (*http.Response, string) {
	var res *http.Response
	var err error
	url := fmt.Sprintf("http://localhost:%v%v", port, path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	require.Nil(t, err)
	res, err = http.DefaultClient.Do(req)
	//	res, err = http.Post(url, "application/json", bytes.NewBuffer(payload))
	require.Nil(t, err)

	output, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	require.Nil(t, err)

	return res, string(output)
}

func getAccount(t *testing.T, sendAddr string) auth.Account {
	// get the account to get the sequence
	res, body := request(t, port, "GET", "/accounts/"+sendAddr, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var acc auth.Account
	err := cdc.UnmarshalJSON([]byte(body), &acc)
	require.Nil(t, err)
	return acc
}

func doSend(t *testing.T, port, seed string) (receiveAddr string, resultTx ctypes.ResultBroadcastTxCommit) {

	// create receive address
	kb := client.MockKeyBase()
	receiveInfo, _, err := kb.Create("receive_address", "1234567890", cryptoKeys.CryptoAlgo("ed25519"))
	require.Nil(t, err)
	receiveAddr, _ = sdk.Bech32ifyAcc(receiveInfo.PubKey.Address())

	acc := getAccount(t, sendAddr)
	sequence := acc.GetSequence()

	// send
	jsonStr := []byte(fmt.Sprintf(`{ "name":"%s", "password":"%s", "sequence":%d, "amount":[{ "denom": "%s", "amount": 1 }] }`, name, password, sequence, coinDenom))
	res, body := request(t, port, "POST", "/accounts/"+receiveAddr+"/send", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	return receiveAddr, resultTx
}

func doIBCTransfer(t *testing.T, port, seed string) (resultTx ctypes.ResultBroadcastTxCommit) {
	// create receive address
	kb := client.MockKeyBase()
	receiveInfo, _, err := kb.Create("receive_address", "1234567890", cryptoKeys.CryptoAlgo("ed25519"))
	require.Nil(t, err)
	receiveAddr, _ := sdk.Bech32ifyAcc(receiveInfo.PubKey.Address())

	// get the account to get the sequence
	acc := getAccount(t, sendAddr)
	sequence := acc.GetSequence()

	// send
	jsonStr := []byte(fmt.Sprintf(`{ "name":"%s", "password":"%s", "sequence":%d, "amount":[{ "denom": "%s", "amount": 1 }] }`, name, password, sequence, coinDenom))
	res, body := request(t, port, "POST", "/ibc/testchain/"+receiveAddr+"/send", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	return resultTx
}

func getDelegation(t *testing.T, delegatorAddr, candidateAddr string) stake.Delegation {
	// get the account to get the sequence
	res, body := request(t, port, "GET", "/stake/"+delegatorAddr+"/bonding_status/"+candidateAddr, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var bond stake.Delegation
	err := cdc.UnmarshalJSON([]byte(body), &bond)
	require.Nil(t, err)
	return bond
}

func doBond(t *testing.T, port, seed string) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, sendAddr)
	sequence := acc.GetSequence()

	// send
	jsonStr := []byte(fmt.Sprintf(`{
		"name": "%s",
		"password": "%s",
		"sequence": %d,
		"delegate": [
			{
				"delegator_addr": "%s",
				"validator_addr": "%s",
				"bond": { "denom": "%s", "amount": 10 }
			}
		],
		"unbond": []
	}`, name, password, sequence, sendAddr, validatorAddr1, coinDenom))
	res, body := request(t, port, "POST", "/stake/delegations", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func doUnbond(t *testing.T, port, seed string) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, sendAddr)
	sequence := acc.GetSequence()

	// send
	jsonStr := []byte(fmt.Sprintf(`{
		"name": "%s",
		"password": "%s",
		"sequence": %d,
		"bond": [],
		"unbond": [
			{
				"delegator_addr": "%s",
				"validator_addr": "%s",
				"shares": "1"
			}
		]
	}`, name, password, sequence, sendAddr, validatorAddr1))
	res, body := request(t, port, "POST", "/stake/delegations", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err := cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

func getValidators(t *testing.T) []stakerest.StakeValidatorOutput {
	// get the account to get the sequence
	res, body := request(t, port, "GET", "/stake/validators", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var validators []stakerest.StakeValidatorOutput
	err := cdc.UnmarshalJSON([]byte(body), &validators)
	require.Nil(t, err)
	return validators
}
