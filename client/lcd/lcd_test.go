package lcd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	cryptoKeys "github.com/tendermint/go-crypto/keys"
	dbm "github.com/tendermint/tmlibs/db"
)

func TestKeys(t *testing.T) {
	kb, err := initKeybase()
	if err != nil {
		t.Errorf("Couldn't init Keybase. Error $s", err.Error())
	}
	cdc := app.MakeCodec()
	r := initRouter(cdc)

	// empty keys
	req, _ := http.NewRequest("GET", "/keys", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	checkResponseCode(t, http.StatusOK, res.Code)
	if body := res.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}

	info, _, err := kb.Create("test", "1234567890", cryptoKeys.CryptoAlgo("ed25519"))
	if err != nil {
		t.Errorf("Couldn't add key. Error $s", err.Error())
	}

	// existing keys
	req, _ = http.NewRequest("GET", "/keys", nil)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	checkResponseCode(t, http.StatusOK, res.Code)
	var m [1]keys.KeyOutput
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&m)

	if m[0].Name != "test" {
		t.Errorf("Did not serve keys name correctly. Got %s", m[0].Name)
	}
	if m[0].Address != info.PubKey.Address().String() {
		t.Errorf("Did not serve keys Address correctly. Got %s, Expected %s", m[0].Address, info.PubKey.Address().String())
	}

	// select key
	req, _ = http.NewRequest("GET", "/keys/test", nil)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	checkResponseCode(t, http.StatusOK, res.Code)
	var m2 keys.KeyOutput
	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&m2)

	if m2.Name != "test" {
		t.Errorf("Did not serve keys name correctly. Got %s", m2.Name)
	}
	if m2.Address != info.PubKey.Address().String() {
		t.Errorf("Did not serve keys Address correctly. Got %s, Expected %s", m2.Address, info.PubKey.Address().String())
	}

	// update key
	var jsonStr = []byte(`{"old_password":"1234567890", "new_password":"12345678901"}`)
	req, _ = http.NewRequest("PUT", "/keys/test", bytes.NewBuffer(jsonStr))
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	checkResponseCode(t, http.StatusOK, res.Code)

	// here it should say unauthorized as we changed the password before
	req, _ = http.NewRequest("PUT", "/keys/test", bytes.NewBuffer(jsonStr))
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	checkResponseCode(t, http.StatusUnauthorized, res.Code)

	// delete key
	jsonStr = []byte(`{"password":"12345678901"}`)
	req, _ = http.NewRequest("DELETE", "/keys/test", bytes.NewBuffer(jsonStr))
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	checkResponseCode(t, http.StatusOK, res.Code)
}

func initKeybase() (cryptoKeys.Keybase, error) {
	os.RemoveAll("./testKeybase")
	db, err := dbm.NewGoLevelDB("keys", "./testKeybase")
	if err != nil {
		return nil, err
	}
	kb := client.GetKeyBase(db)
	keys.SetKeyBase(kb)
	return kb, nil
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
