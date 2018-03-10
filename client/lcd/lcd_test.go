package lcd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	cryptoKeys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/tendermint/p2p"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	"github.com/cosmos/cosmos-sdk/server"
)

func TestKeys(t *testing.T) {
	_, db, err := initKeybase(t)
	require.Nil(t, err, "Couldn't init Keybase")

	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// empty keys
	res := request(t, r, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	body := res.Body.String()
	assert.Equal(t, body, "[]", "Expected an empty array")

	// add key
	addr := createKey(t, r)
	assert.Len(t, addr, 40, "Returned address has wrong format", res.Body.String())

	// existing keys
	res = request(t, r, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var m [1]keys.KeyOutput
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)
	require.NoError(t, err)

	assert.Equal(t, m[0].Name, "test", "Did not serve keys name correctly")
	assert.Equal(t, m[0].Address, addr, "Did not serve keys Address correctly")

	// select key
	res = request(t, r, "GET", "/keys/test", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var m2 keys.KeyOutput
	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&m2)

	assert.Equal(t, m2.Name, "test", "Did not serve keys name correctly")
	assert.Equal(t, m2.Address, addr, "Did not serve keys Address correctly")

	// update key
	var jsonStr = []byte(`{"old_password":"1234567890", "new_password":"12345678901"}`)
	res = request(t, r, "PUT", "/keys/test", jsonStr)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	// here it should say unauthorized as we changed the password before
	res = request(t, r, "PUT", "/keys/test", jsonStr)
	require.Equal(t, http.StatusUnauthorized, res.Code, res.Body.String())

	// delete key
	jsonStr = []byte(`{"password":"12345678901"}`)
	res = request(t, r, "DELETE", "/keys/test", jsonStr)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	db.Close()
}

func TestVersion(t *testing.T) {
	prepareClient(t)
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// node info
	res := request(t, r, "GET", "/version", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	// TODO fix regexp
	// reg, err := regexp.Compile(`v\d+\.\d+\.\d+(-dev)?`)
	// require.Nil(t, err)
	// match := reg.MatchString(res.Body.String())
	// assert.True(t, match, res.Body.String())
	assert.Equal(t, "0.11.1-dev", res.Body.String())
}

func TestNodeStatus(t *testing.T) {
	ch := server.StartServer(t)
	defer close(ch)
	prepareClient(t)

	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// node info
	res := request(t, r, "GET", "/node_info", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m p2p.NodeInfo
	decoder := json.NewDecoder(res.Body)
	err := decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse node info")

	assert.NotEqual(t, p2p.NodeInfo{}, m, "res: %v", res)

	// syncing
	res = request(t, r, "GET", "/syncing", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.Equal(t, "true", res.Body.String())
}

func TestBlock(t *testing.T) {
	ch := server.StartServer(t)
	defer close(ch)
	prepareClient(t)

	cdc := app.MakeCodec()
	r := initRouter(cdc)

	res := request(t, r, "GET", "/blocks/latest", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m ctypes.ResultBlock
	decoder := json.NewDecoder(res.Body)
	err := decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse block")

	assert.NotEqual(t, ctypes.ResultBlock{}, m)

	// --

	res = request(t, r, "GET", "/blocks/1", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.NotEqual(t, ctypes.ResultBlock{}, m)

	// --

	res = request(t, r, "GET", "/blocks/2", nil)
	require.Equal(t, http.StatusNotFound, res.Code, res.Body.String())
}

func TestValidators(t *testing.T) {
	ch := server.StartServer(t)
	defer close(ch)

	prepareClient(t)
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	res := request(t, r, "GET", "/validatorsets/latest", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m ctypes.ResultValidators
	decoder := json.NewDecoder(res.Body)
	err := decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse validatorset")

	assert.NotEqual(t, ctypes.ResultValidators{}, m)

	// --

	res = request(t, r, "GET", "/validatorsets/1", nil)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.NotEqual(t, ctypes.ResultValidators{}, m)

	// --

	res = request(t, r, "GET", "/validatorsets/2", nil)
	require.Equal(t, http.StatusNotFound, res.Code)
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

// setupViper creates a homedir to run inside,
// and returns a cleanup function to defer
func setupViper() func() {
	rootDir, err := ioutil.TempDir("", "mock-sdk-cmd")
	if err != nil {
		panic(err) // fuck it!
	}
	viper.Set("home", rootDir)
	return func() {
		os.RemoveAll(rootDir)
	}
}

func startServer(t *testing.T) {
	defer setupViper()()
	// init server
	initCmd := server.InitCmd(mock.GenInitOptions, log.NewNopLogger())
	err := initCmd.RunE(nil, nil)
	require.NoError(t, err)

	// start server
	viper.Set("with-tendermint", true)
	startCmd := server.StartCmd(mock.NewApp, log.NewNopLogger())
	timeout := time.Duration(3) * time.Second

	err = runOrTimeout(startCmd, timeout)
	require.NoError(t, err)
}

// copied from server/start_test.go
func runOrTimeout(cmd *cobra.Command, timeout time.Duration) error {
	done := make(chan error)
	go func(out chan<- error) {
		// this should NOT exit
		err := cmd.RunE(nil, nil)
		if err != nil {
			out <- err
		}
		out <- fmt.Errorf("start died for unknown reasons")
	}(done)
	timer := time.NewTimer(timeout)

	select {
	case err := <-done:
		return err
	case <-timer.C:
		return nil
	}
}

func createKey(t *testing.T, r http.Handler) string {
	var jsonStr = []byte(`{"name":"test", "password":"1234567890"}`)
	res := request(t, r, "POST", "/keys", jsonStr)

	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

	addr := res.Body.String()
	return addr
}

func request(t *testing.T, r http.Handler, method string, path string, payload []byte) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, bytes.NewBuffer(payload))
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	return res
}

func initKeybase(t *testing.T) (cryptoKeys.Keybase, *dbm.GoLevelDB, error) {
	os.RemoveAll("./testKeybase")
	db, err := dbm.NewGoLevelDB("keys", "./testKeybase")
	require.Nil(t, err)
	kb := client.GetKeyBase(db)
	keys.SetKeyBase(kb)
	return kb, db, nil
}
