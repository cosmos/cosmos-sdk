package rest

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/cryptostore"
	"github.com/tendermint/go-crypto/keys/storage/memstorage"
)

func getKeyManager() keys.Manager {
	return cryptostore.New(
		cryptostore.SecretBox,
		memstorage.New(),
		keys.MustLoadCodec("english"),
	)
}

func equalKeys(t *testing.T, k1, k2 keys.Info, name bool) {
	assert.Equal(t, k1.Address, k2.Address)
	assert.Equal(t, k1.PubKey, k2.PubKey)
}

func TestKeysCreateGetRecover(t *testing.T) {
	keyMan := getKeyManager()
	serviceKeys := NewServiceKeys(keyMan)

	r := mux.NewRouter()
	err := serviceKeys.RegisterCRUD(r)
	assert.Nil(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	var (
		keyInfo keys.Info

		passPhrase string = "abcdefghijklmno"
		seedPhrase string
	)
	keyInfo.Name = "mykey"

	// create the key
	{
		reqCreate := RequestCreate{
			Name:       keyInfo.Name,
			Passphrase: passPhrase,
			Algo:       defaultAlgo,
		}
		b, err := json.Marshal(reqCreate)
		assert.Nil(t, err)

		resp, err := http.Post(ts.URL+"/keys", "json", bytes.NewBuffer(b))
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, 200)

		var resCreate ResponseCreate
		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		err = json.Unmarshal(body, &resCreate)
		assert.Nil(t, err)

		assert.Equal(t, keyInfo.Name, resCreate.Key.Name)
		keyInfo = resCreate.Key
		seedPhrase = resCreate.Seed
	}

	// get the key and confirm it matches
	{
		resp, err := http.Get(ts.URL + "/keys/" + keyInfo.Name)
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, 200)

		var resKeyInfo keys.Info
		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		err = json.Unmarshal(body, &resKeyInfo)
		assert.Nil(t, err)

		equalKeys(t, keyInfo, resKeyInfo, true)
	}

	// recover the key and confirm it matches
	{

		reqRecover := RequestRecover{
			Name:       "newName",
			Passphrase: passPhrase,
			Seed:       seedPhrase,
			Algo:       defaultAlgo,
		}
		b, err := json.Marshal(reqRecover)
		assert.Nil(t, err)

		resp, err := http.Post(ts.URL+"/keys/recover", "json", bytes.NewBuffer(b))
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, 200)

		var resRecover ResponseRecover
		body, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		err = json.Unmarshal(body, &resRecover)
		assert.Nil(t, err)

		equalKeys(t, keyInfo, resRecover.Key, false)
	}
}
