package lcd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	cryptoKeys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/tendermint/p2p"
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

func TestNodeInfo(t *testing.T) {
	prepareApp(t)
	_, db, err := initKeybase(t)
	require.Nil(t, err, "Couldn't init Keybase")
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	req, err := http.NewRequest("GET", "/node_info", nil)
	require.Nil(t, err)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var m p2p.NodeInfo
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)
	require.Nil(t, err, "Couldn't parse node info")

	db.Close()
}

//__________________________________________________________
// helpers

func defaultLogger() log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
}

func prepareApp(t *testing.T) {
	logger := defaultLogger()
	db := dbm.NewMemDB()
	name := t.Name()
	app := baseapp.NewBaseApp(name, logger, db)

	header := abci.Header{Height: 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	app.Commit()
}

func initKeybase(t *testing.T) (cryptoKeys.Keybase, *dbm.GoLevelDB, error) {
	os.RemoveAll("./testKeybase")
	db, err := dbm.NewGoLevelDB("keys", "./testKeybase")
	require.Nil(t, err)
	kb := client.GetKeyBase(db)
	keys.SetKeyBase(kb)
	return kb, db, nil
}
