package proxy_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	keys "github.com/tendermint/go-keys"
	"github.com/tendermint/go-keys/cryptostore"
	"github.com/tendermint/go-keys/proxy"
	"github.com/tendermint/go-keys/proxy/types"
	"github.com/tendermint/go-keys/storage/memstorage"
)

func TestKeyServer(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	r := setupServer()

	// let's abstract this out a bit....
	keys, code, err := listKeys(r)
	require.Nil(err)
	require.Equal(http.StatusOK, code)
	assert.Equal(0, len(keys))

	algo := "ed25519"
	n1, n2 := "personal", "business"
	p0, p1, p2 := "1234", "over10chars...", "really-secure!@#$"

	// this fails for validation
	_, code, err = createKey(r, n1, p0, algo)
	require.Nil(err, "%+v", err)
	require.NotEqual(http.StatusOK, code)

	// new password better
	key, code, err := createKey(r, n1, p1, algo)
	require.Nil(err, "%+v", err)
	require.Equal(http.StatusOK, code)
	require.Equal(key.Name, n1)

	// the other one works
	key2, code, err := createKey(r, n2, p2, algo)
	require.Nil(err, "%+v", err)
	require.Equal(http.StatusOK, code)
	require.Equal(key2.Name, n2)

	// let's abstract this out a bit....
	keys, code, err = listKeys(r)
	require.Nil(err)
	require.Equal(http.StatusOK, code)
	if assert.Equal(2, len(keys)) {
		// in alphabetical order
		assert.Equal(keys[0].Name, n2)
		assert.Equal(keys[1].Name, n1)
	}

	// get works
	k, code, err := getKey(r, n1)
	require.Nil(err, "%+v", err)
	require.Equal(http.StatusOK, code)
	assert.Equal(k.Name, n1)
	assert.NotNil(k.Address)
	assert.Equal(k.Address, key.Address)

	// delete with proper key
	_, code, err = deleteKey(r, n1, p1)
	require.Nil(err, "%+v", err)
	require.Equal(http.StatusOK, code)

	// after delete, get and list different
	_, code, err = getKey(r, n1)
	require.Nil(err, "%+v", err)
	require.NotEqual(http.StatusOK, code)
	keys, code, err = listKeys(r)
	require.Nil(err, "%+v", err)
	require.Equal(http.StatusOK, code)
	if assert.Equal(1, len(keys)) {
		assert.Equal(keys[0].Name, n2)
	}

}

func setupServer() http.Handler {
	// make the storage with reasonable defaults
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		memstorage.New(),
	)

	// build your http server
	ks := proxy.NewKeyServer(cstore, "ed25519")
	r := mux.NewRouter()
	sk := r.PathPrefix("/keys").Subrouter()
	ks.Register(sk)
	return r
}

// return data, status code, and error
func listKeys(h http.Handler) (keys.Infos, int, error) {
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/keys/", nil)
	if err != nil {
		return nil, 0, err
	}

	h.ServeHTTP(rr, req)
	if http.StatusOK != rr.Code {
		return nil, rr.Code, nil
	}

	data := keys.Infos{}
	err = json.Unmarshal(rr.Body.Bytes(), &data)
	return data, rr.Code, err
}

func getKey(h http.Handler, name string) (*keys.Info, int, error) {
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/keys/"+name, nil)
	if err != nil {
		return nil, 0, err
	}

	h.ServeHTTP(rr, req)
	if http.StatusOK != rr.Code {
		return nil, rr.Code, nil
	}

	data := keys.Info{}
	err = json.Unmarshal(rr.Body.Bytes(), &data)
	return &data, rr.Code, err
}

func createKey(h http.Handler, name, passphrase, algo string) (*keys.Info, int, error) {
	rr := httptest.NewRecorder()
	post := types.CreateKeyRequest{
		Name:       name,
		Passphrase: passphrase,
		Algo:       algo,
	}
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(&post)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", "/keys/", &b)
	if err != nil {
		return nil, 0, err
	}

	h.ServeHTTP(rr, req)
	if http.StatusOK != rr.Code {
		return nil, rr.Code, nil
	}

	data := keys.Info{}
	err = json.Unmarshal(rr.Body.Bytes(), &data)
	return &data, rr.Code, err
}

func deleteKey(h http.Handler, name, passphrase string) (*types.ErrorResponse, int, error) {
	rr := httptest.NewRecorder()
	post := types.DeleteKeyRequest{
		Name:       name,
		Passphrase: passphrase,
	}
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(&post)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("DELETE", "/keys/"+name, &b)
	if err != nil {
		return nil, 0, err
	}

	h.ServeHTTP(rr, req)
	if http.StatusOK != rr.Code {
		return nil, rr.Code, nil
	}

	data := types.ErrorResponse{}
	err = json.Unmarshal(rr.Body.Bytes(), &data)
	return &data, rr.Code, err
}
