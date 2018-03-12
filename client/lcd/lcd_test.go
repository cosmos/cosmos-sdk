package lcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	cryptoKeys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/tendermint/p2p"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

func TestKeys(t *testing.T) {
	cmd := junkInit(t)
	defer cmd.Process.Kill()

	// empty keys
	res, body := request(t, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	assert.Equal(t, body, "[]", "Expected an empty array")

	// get seed
	res, body = request(t, "GET", "/keys/seed", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	seed := body
	reg, err := regexp.Compile(`([a-z]+ ){12}`)
	require.Nil(t, err)
	match := reg.MatchString(seed)
	assert.True(t, match, "Returned seed has wrong foramt", seed)

	// add key
	var jsonStr = []byte(`{"name":"test_fail", "password":"1234567890"}`)
	res, body = request(t, "POST", "/keys", jsonStr)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Account creation should require a seed")

	jsonStr = []byte(fmt.Sprintf(`{"name":"test", "password":"1234567890", "seed": "%s"}`, seed))
	res, body = request(t, "POST", "/keys", jsonStr)

	assert.Equal(t, http.StatusOK, res.StatusCode, body)
	addr := body
	assert.Len(t, addr, 40, "Returned address has wrong format", addr)

	// existing keys
	res, body = request(t, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m [1]keys.KeyOutput
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)
	require.NoError(t, err)

	assert.Equal(t, m[0].Name, "test", "Did not serve keys name correctly")
	assert.Equal(t, m[0].Address, addr, "Did not serve keys Address correctly")

	// select key
	res, body = request(t, "GET", "/keys/test", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m2 keys.KeyOutput
	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&m2)

	assert.Equal(t, m2.Name, "test", "Did not serve keys name correctly")
	assert.Equal(t, m2.Address, addr, "Did not serve keys Address correctly")

	// update key
	jsonStr = []byte(`{"old_password":"1234567890", "new_password":"12345678901"}`)
	res, body = request(t, "PUT", "/keys/test", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// here it should say unauthorized as we changed the password before
	res, body = request(t, "PUT", "/keys/test", jsonStr)
	require.Equal(t, http.StatusUnauthorized, res.StatusCode, body)

	// delete key
	jsonStr = []byte(`{"password":"12345678901"}`)
	res, body = request(t, "DELETE", "/keys/test", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
}

//XXX
func junkInit(t *testing.T) *exec.Cmd {
	tests.TestInitBasecoin(t)
	return tests.StartServerForTest(t)
}

func TestVersion(t *testing.T) {
	cmd := junkInit(t)
	defer cmd.Process.Kill()

	// node info
	res, body := request(t, "GET", "/version", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	reg, err := regexp.Compile(`\d+\.\d+\.\d+(-dev)?`)
	require.Nil(t, err)
	match := reg.MatchString(body)
	assert.True(t, match, body)
}

func TestNodeStatus(t *testing.T) {
	cmd := junkInit(t)
	defer cmd.Process.Kill()

	// node info
	res, body := request(t, "GET", "/node_info", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var m p2p.NodeInfo
	decoder := json.NewDecoder(res.Body)
	err := decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse node info")

	assert.NotEqual(t, p2p.NodeInfo{}, m, "res: %v", res)

	// syncing
	res, body = request(t, "GET", "/syncing", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	assert.Equal(t, "true", body)
}

func TestBlock(t *testing.T) {
	cmd := junkInit(t)
	defer cmd.Process.Kill()

	// res, body := request(t, "GET", "/blocks/latest", nil)
	// require.Equal(t, http.StatusOK, res.StatusCode, body)

	// var m ctypes.ResultBlock
	// decoder := json.NewDecoder(res.Body)
	// err := decoder.Decode(&m)
	// require.Nil(t, err, "Couldn't parse block")

	// assert.NotEqual(t, ctypes.ResultBlock{}, m)

	// --

	res, body := request(t, "GET", "/blocks/1", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var m ctypes.ResultBlock
	decoder := json.NewDecoder(res.Body)
	err := decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse block")

	assert.NotEqual(t, ctypes.ResultBlock{}, m)

	// --

	res, body = request(t, "GET", "/blocks/2", nil)
	require.Equal(t, http.StatusNotFound, res.StatusCode, body)
}

func TestValidators(t *testing.T) {
	cmd := junkInit(t)
	defer cmd.Process.Kill()

	// res, body := request(t, "GET", "/validatorsets/latest", nil)
	// require.Equal(t, http.StatusOK, res.StatusCode, body)

	// var m ctypes.ResultValidators
	// decoder := json.NewDecoder(res.Body)
	// err := decoder.Decode(&m)
	// require.Nil(t, err, "Couldn't parse validatorset")

	// assert.NotEqual(t, ctypes.ResultValidators{}, m)

	// --

	res, body := request(t, "GET", "/validatorsets/1", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var m ctypes.ResultValidators
	decoder := json.NewDecoder(res.Body)
	err := decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse validatorset")

	assert.NotEqual(t, ctypes.ResultValidators{}, m)

	// --

	res, body = request(t, "GET", "/validatorsets/2", nil)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestCoinSend(t *testing.T) {
	cmd := junkInit(t)
	defer cmd.Process.Kill()

	// TODO make that account has coins
	kb := client.MockKeyBase()
	info, seed, err := kb.Create("account_with_coins", "1234567890", cryptoKeys.CryptoAlgo("ed25519"))
	require.NoError(t, err)
	addr := string(info.Address())

	// query empty
	res, body := request(t, "GET", "/accounts/1234567890123456789012345678901234567890", nil)
	require.Equal(t, http.StatusNoContent, res.StatusCode, body)

	// query
	res, body = request(t, "GET", "/accounts/"+addr, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	assert.Equal(t, `{
		"coins": [
			{
				"denom": "mycoin",
				"amount": 9007199254740992
			}
		]
	}`, body)

	// create account to send in keybase
	var jsonStr = []byte(fmt.Sprintf(`{"name":"test", "password":"1234567890", "seed": "%s"}`, seed))
	res, body = request(t, "POST", "/keys", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// create receive address
	receiveInfo, _, err := kb.Create("receive_address", "1234567890", cryptoKeys.CryptoAlgo("ed25519"))
	require.NoError(t, err)
	receiveAddr := string(receiveInfo.Address())

	// send
	jsonStr = []byte(`{
		"name":"test", 
		"password":"1234567890", 
		"amount":[{
			"denom": "mycoin",
			"amount": 1
		}]
	}`)
	res, body = request(t, "POST", "/accounts/"+receiveAddr+"/send", jsonStr)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// check if received
	res, body = request(t, "GET", "/accounts/"+receiveAddr, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	assert.Equal(t, `{
		"coins": [
			{
				"denom": "mycoin",
				"amount": 1
			}
		]
	}`, body)
}

//__________________________________________________________
// helpers

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}

func prepareClient(t *testing.T) {
	db := dbm.NewMemDB()
	app := baseapp.NewBaseApp(t.Name(), defaultLogger(), db)
	viper.Set(client.FlagNode, "localhost:46657")
	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()
}

func request(t *testing.T, method string, path string, payload []byte) (*http.Response, string) {
	var res *http.Response
	var err error
	if method == "GET" {
		res, err = http.Get(path)
	}
	if method == "POST" {
		res, err = http.Post(path, "application/json", bytes.NewBuffer(payload))
	}
	require.NoError(t, err)

	output, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	return res, string(output)
}

func initKeybase(t *testing.T) (cryptoKeys.Keybase, *dbm.GoLevelDB, error) {
	os.RemoveAll("./testKeybase")
	db, err := dbm.NewGoLevelDB("keys", "./testKeybase")
	require.Nil(t, err)
	kb := client.GetKeyBase(db)
	keys.SetKeyBase(kb)
	return kb, db, nil
}
