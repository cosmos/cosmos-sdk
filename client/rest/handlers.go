package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/tendermint/basecoin"
	keysutils "github.com/tendermint/go-crypto/cmd"
	keys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/tmlibs/common"
)

type Keys struct {
	algo    string
	manager keys.Manager
}

func DefaultKeysManager() keys.Manager {
	return keysutils.GetKeyManager()
}

func New(manager keys.Manager, algo string) *Keys {
	return &Keys{
		algo:    algo,
		manager: manager,
	}
}

func (k *Keys) GenerateKey(w http.ResponseWriter, r *http.Request) {
	ckReq := &CreateKeyRequest{
		Algo: k.algo,
	}
	if err := common.ParseRequestAndValidateJSON(r, ckReq); err != nil {
		common.WriteError(w, err)
		return
	}

	key, seed, err := k.manager.Create(ckReq.Name, ckReq.Passphrase, ckReq.Algo)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	res := &CreateKeyResponse{Key: key, Seed: seed}
	common.WriteSuccess(w, res)
}

func (k *Keys) GetKey(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)
	name := query["name"]
	key, err := k.manager.Get(name)
	if err != nil {
		common.WriteError(w, err)
		return
	}
	common.WriteSuccess(w, &key)
}

func (k *Keys) ListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := k.manager.List()
	if err != nil {
		common.WriteError(w, err)
		return
	}
	common.WriteSuccess(w, keys)
}

var (
	errNonMatchingPathAndJSONKeyNames = errors.New("path and json key names don't match")
)

func (k *Keys) UpdateKey(w http.ResponseWriter, r *http.Request) {
	uReq := new(UpdateKeyRequest)
	if err := common.ParseRequestAndValidateJSON(r, uReq); err != nil {
		common.WriteError(w, err)
		return
	}

	query := mux.Vars(r)
	name := query["name"]
	if name != uReq.Name {
		common.WriteError(w, errNonMatchingPathAndJSONKeyNames)
		return
	}

	if err := k.manager.Update(uReq.Name, uReq.OldPass, uReq.NewPass); err != nil {
		common.WriteError(w, err)
		return
	}

	key, err := k.manager.Get(uReq.Name)
	if err != nil {
		common.WriteError(w, err)
		return
	}
	common.WriteSuccess(w, &key)
}

func (k *Keys) DeleteKey(w http.ResponseWriter, r *http.Request) {
	dReq := new(DeleteKeyRequest)
	if err := common.ParseRequestAndValidateJSON(r, dReq); err != nil {
		common.WriteError(w, err)
		return
	}

	query := mux.Vars(r)
	name := query["name"]
	if name != dReq.Name {
		common.WriteError(w, errNonMatchingPathAndJSONKeyNames)
		return
	}

	if err := k.manager.Delete(dReq.Name, dReq.Passphrase); err != nil {
		common.WriteError(w, err)
		return
	}

	resp := &common.ErrorResponse{Success: true}
	common.WriteSuccess(w, resp)
}

func (k *Keys) Register(r *mux.Router) {
	r.HandleFunc("/keys", k.GenerateKey).Methods("POST")
	r.HandleFunc("/keys", k.ListKeys).Methods("GET")
	r.HandleFunc("/keys/{name}", k.GetKey).Methods("GET")
	r.HandleFunc("/keys/{name}", k.UpdateKey).Methods("POST", "PUT")
	r.HandleFunc("/keys/{name}", k.DeleteKey).Methods("DELETE")
}

type Context struct {
	Keys *Keys
}

func (ctx *Context) RegisterHandlers(r *mux.Router) error {
	ctx.Keys.Register(r)
	r.HandleFunc("/sign", doSign).Methods("POST")
	r.HandleFunc("/tx", doPostTx).Methods("POST")

	return nil
}

func doPostTx(w http.ResponseWriter, r *http.Request) {
	tx := new(basecoin.Tx)
	if err := common.ParseRequestAndValidateJSON(r, tx); err != nil {
		common.WriteError(w, err)
		return
	}
	commit, err := PostTx(*tx)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	common.WriteSuccess(w, commit)
}

func doSign(w http.ResponseWriter, r *http.Request) {
	sr := new(SignRequest)
	if err := common.ParseRequestAndValidateJSON(r, sr); err != nil {
		common.WriteError(w, err)
		return
	}

	tx := sr.Tx
	if err := SignTx(sr.Name, sr.Password, tx); err != nil {
		common.WriteError(w, err)
		return
	}
	common.WriteSuccess(w, tx)
}
