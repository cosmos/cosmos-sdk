package server

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	keys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/server/types"
)

type Keys struct {
	manager keys.Manager
	algo    string
}

func New(manager keys.Manager, algo string) Keys {
	return Keys{
		manager: manager,
		algo:    algo,
	}
}

func (k Keys) GenerateKey(w http.ResponseWriter, r *http.Request) {
	req := types.CreateKeyRequest{
		Algo: k.algo, // default key type from cli
	}
	err := readRequest(r, &req)
	if err != nil {
		writeError(w, err)
		return
	}

	key, seed, err := k.manager.Create(req.Name, req.Passphrase, req.Algo)
	if err != nil {
		writeError(w, err)
		return
	}

	res := types.CreateKeyResponse{key, seed}
	writeSuccess(w, &res)
}

func (k Keys) GetKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	key, err := k.manager.Get(name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, &key)
}

func (k Keys) ListKeys(w http.ResponseWriter, r *http.Request) {

	keys, err := k.manager.List()
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, keys)
}

func (k Keys) UpdateKey(w http.ResponseWriter, r *http.Request) {
	req := types.UpdateKeyRequest{}
	err := readRequest(r, &req)
	if err != nil {
		writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]
	if name != req.Name {
		writeError(w, errors.New("path and json key names don't match"))
		return
	}

	err = k.manager.Update(req.Name, req.OldPass, req.NewPass)
	if err != nil {
		writeError(w, err)
		return
	}

	key, err := k.manager.Get(req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, &key)
}

func (k Keys) DeleteKey(w http.ResponseWriter, r *http.Request) {
	req := types.DeleteKeyRequest{}
	err := readRequest(r, &req)
	if err != nil {
		writeError(w, err)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]
	if name != req.Name {
		writeError(w, errors.New("path and json key names don't match"))
		return
	}

	err = k.manager.Delete(req.Name, req.Passphrase)
	if err != nil {
		writeError(w, err)
		return
	}

	// not really an error, but something generic
	resp := types.ErrorResponse{
		Success: true,
	}
	writeSuccess(w, &resp)
}

func (k Keys) Register(r *mux.Router) {
	r.HandleFunc("/", k.GenerateKey).Methods("POST")
	r.HandleFunc("/", k.ListKeys).Methods("GET")
	r.HandleFunc("/{name}", k.GetKey).Methods("GET")
	r.HandleFunc("/{name}", k.UpdateKey).Methods("POST", "PUT")
	r.HandleFunc("/{name}", k.DeleteKey).Methods("DELETE")
}
