package proxy

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	keys "github.com/tendermint/go-keys"
	"github.com/tendermint/go-keys/proxy/types"
)

type KeyServer struct {
	manager keys.Manager
}

func NewKeyServer(manager keys.Manager) KeyServer {
	return KeyServer{
		manager: manager,
	}
}

func (k KeyServer) GenerateKey(w http.ResponseWriter, r *http.Request) {
	req := types.CreateKeyRequest{}
	err := readRequest(r, &req)
	if err != nil {
		writeError(w, err)
		return
	}

	key, err := k.manager.Create(req.Name, req.Passphrase, req.Algo)
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, &key)
}

func (k KeyServer) GetKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	key, err := k.manager.Get(name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, &key)
}

func (k KeyServer) ListKeys(w http.ResponseWriter, r *http.Request) {

	keys, err := k.manager.List()
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, keys)
}

func (k KeyServer) UpdateKey(w http.ResponseWriter, r *http.Request) {
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

func (k KeyServer) DeleteKey(w http.ResponseWriter, r *http.Request) {
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

func (k KeyServer) Register(r *mux.Router) {
	r.HandleFunc("/", k.GenerateKey).Methods("POST")
	r.HandleFunc("/", k.ListKeys).Methods("GET")
	r.HandleFunc("/{name}", k.GetKey).Methods("GET")
	r.HandleFunc("/{name}", k.UpdateKey).Methods("POST", "PUT")
	r.HandleFunc("/{name}", k.DeleteKey).Methods("DELETE")
}
