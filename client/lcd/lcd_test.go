package lcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	"github.com/cosmos/cosmos-sdk/mock"
	"github.com/cosmos/cosmos-sdk/server"
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
	kb, db, err := initKeybase(t)
	require.Nil(t, err, "Couldn't init Keybase")

	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// empty keys
	req, err := http.NewRequest("GET", "/keys", nil)
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
	body := res.Body.String()
	require.Equal(t, body, "[]", "Expected an empty array")

	info, _, err := kb.Create("test", "1234567890", cryptoKeys.CryptoAlgo("ed25519"))
	require.Nil(t, err, "Couldn't add key")

	// existing keys
	req, err = http.NewRequest("GET", "/keys", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var m [1]keys.KeyOutput
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)

	assert.Equal(t, m[0].Name, "test", "Did not serve keys name correctly")
	assert.Equal(t, m[0].Address, info.PubKey.Address().String(), "Did not serve keys Address correctly")

	// select key
	req, _ = http.NewRequest("GET", "/keys/test", nil)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
	var m2 keys.KeyOutput
	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&m2)

	assert.Equal(t, m2.Name, "test", "Did not serve keys name correctly")
	assert.Equal(t, m2.Address, info.PubKey.Address().String(), "Did not serve keys Address correctly")

	// update key
	var jsonStr = []byte(`{"old_password":"1234567890", "new_password":"12345678901"}`)
	req, err = http.NewRequest("PUT", "/keys/test", bytes.NewBuffer(jsonStr))
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

	// here it should say unauthorized as we changed the password before
	req, err = http.NewRequest("PUT", "/keys/test", bytes.NewBuffer(jsonStr))
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code, res.Body.String())

	// delete key
	jsonStr = []byte(`{"password":"12345678901"}`)
	req, err = http.NewRequest("DELETE", "/keys/test", bytes.NewBuffer(jsonStr))
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code, res.Body.String())

	db.Close()
}

func TestNodeStatus(t *testing.T) {
	startServer(t)
	// TODO need to kill server after
	prepareClient(t)
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// node info
	req, err := http.NewRequest("GET", "/node_info", nil)
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m p2p.NodeInfo
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse node info")

	assert.NotEqual(t, p2p.NodeInfo{}, m)

	// syncing
	req, err = http.NewRequest("GET", "/syncing", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.Equal(t, "true", res.Body.String())
}

func TestBlock(t *testing.T) {
	startServer(t)
	// TODO need to kill server after
	prepareClient(t)
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	req, err := http.NewRequest("GET", "/blocks/latest", nil)
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m ctypes.ResultBlock
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse block")

	assert.NotEqual(t, ctypes.ResultBlock{}, m)

	req, err = http.NewRequest("GET", "/blocks/1", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.NotEqual(t, ctypes.ResultBlock{}, m)

	req, err = http.NewRequest("GET", "/blocks/2", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestValidators(t *testing.T) {
	startServer(t)
	// TODO need to kill server after
	prepareClient(t)
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	req, err := http.NewRequest("GET", "/validatorsets/latest", nil)
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m ctypes.ResultValidators
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse block")

	assert.NotEqual(t, ctypes.ResultValidators{}, m)

	req, err = http.NewRequest("GET", "/validatorsets/1", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	assert.NotEqual(t, ctypes.ResultValidators{}, m)

	req, err = http.NewRequest("GET", "/validatorsets/2", nil)
	require.Nil(t, err)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
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

func initKeybase(t *testing.T) (cryptoKeys.Keybase, *dbm.GoLevelDB, error) {
	os.RemoveAll("./testKeybase")
	db, err := dbm.NewGoLevelDB("keys", "./testKeybase")
	require.Nil(t, err)
	kb := client.GetKeyBase(db)
	keys.SetKeyBase(kb)
	return kb, db, nil
}
