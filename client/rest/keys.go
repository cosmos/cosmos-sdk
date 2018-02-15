package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	keys "github.com/tendermint/go-crypto/keys"

	sdk "github.com/cosmos/cosmos-sdk"
)

const (
	defaultAlgo = "ed25519" // TODO: allow this to come in via requests
)

var (
	errNonMatchingPathAndJSONKeyNames = errors.New("path and json key names don't match")
)

// ServiceKeys exposes a REST API service for
// managing keys and signing transactions
type ServiceKeys struct {
	manager keys.Manager
}

// New returns a new instance of the keys service
func NewServiceKeys(manager keys.Manager) *ServiceKeys {
	return &ServiceKeys{
		manager: manager, // XXX keycmd.GetKeyManager()
	}
}

func (s *ServiceKeys) Create(w http.ResponseWriter, r *http.Request) {
	req := &RequestCreate{
		Algo: defaultAlgo,
	}
	if err := sdk.ParseRequestAndValidateJSON(r, req); err != nil {
		sdk.WriteError(w, err)
		return
	}

	key, seed, err := s.manager.Create(req.Name, req.Passphrase, req.Algo)
	if err != nil {
		sdk.WriteError(w, err)
		return
	}

	res := &ResponseCreate{Key: key, Seed: seed}
	sdk.WriteSuccess(w, res)
}

func (s *ServiceKeys) Get(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)
	name := query["name"]
	key, err := s.manager.Get(name)
	if err != nil {
		sdk.WriteError(w, err)
		return
	}
	sdk.WriteSuccess(w, &key)
}

func (s *ServiceKeys) List(w http.ResponseWriter, r *http.Request) {
	keys, err := s.manager.List()
	if err != nil {
		sdk.WriteError(w, err)
		return
	}
	sdk.WriteSuccess(w, keys)
}

func (s *ServiceKeys) Update(w http.ResponseWriter, r *http.Request) {
	req := new(RequestUpdate)
	if err := sdk.ParseRequestAndValidateJSON(r, req); err != nil {
		sdk.WriteError(w, err)
		return
	}

	query := mux.Vars(r)
	name := query["name"]
	if name != req.Name {
		sdk.WriteError(w, errNonMatchingPathAndJSONKeyNames)
		return
	}

	if err := s.manager.Update(req.Name, req.OldPass, req.NewPass); err != nil {
		sdk.WriteError(w, err)
		return
	}

	key, err := s.manager.Get(req.Name)
	if err != nil {
		sdk.WriteError(w, err)
		return
	}
	sdk.WriteSuccess(w, &key)
}

func (s *ServiceKeys) Delete(w http.ResponseWriter, r *http.Request) {
	req := new(RequestDelete)
	if err := sdk.ParseRequestAndValidateJSON(r, req); err != nil {
		sdk.WriteError(w, err)
		return
	}

	query := mux.Vars(r)
	name := query["name"]
	if name != req.Name {
		sdk.WriteError(w, errNonMatchingPathAndJSONKeyNames)
		return
	}

	if err := s.manager.Delete(req.Name, req.Passphrase); err != nil {
		sdk.WriteError(w, err)
		return
	}

	resp := &sdk.ErrorResponse{Success: true}
	sdk.WriteSuccess(w, resp)
}

func (s *ServiceKeys) Recover(w http.ResponseWriter, r *http.Request) {
	req := &RequestRecover{
		Algo: defaultAlgo,
	}
	if err := sdk.ParseRequestAndValidateJSON(r, req); err != nil {
		sdk.WriteError(w, err)
		return
	}

	key, err := s.manager.Recover(req.Name, req.Passphrase, req.Seed)
	if err != nil {
		sdk.WriteError(w, err)
		return
	}

	res := &ResponseRecover{Key: key}
	sdk.WriteSuccess(w, res)
}

func (s *ServiceKeys) SignTx(w http.ResponseWriter, r *http.Request) {
	req := new(RequestSign)
	if err := sdk.ParseRequestAndValidateJSON(r, req); err != nil {
		sdk.WriteError(w, err)
		return
	}

	tx := req.Tx

	var err error
	if sign, ok := tx.Unwrap().(keys.Signable); ok {
		err = s.manager.Sign(req.Name, req.Password, sign)
	}
	if err != nil {
		sdk.WriteError(w, err)
		return
	}
	sdk.WriteSuccess(w, tx)
}

// mux.Router registrars

// RegisterCRUD is a convenience method to register all
// CRUD for keys to allow access by methods and routes:
// POST:      /keys
// POST:      /keys/recover
// GET:	      /keys
// GET:	      /keys/{name}
// POST, PUT: /keys/{name}
// DELETE:    /keys/{name}
func (s *ServiceKeys) RegisterCRUD(r *mux.Router) error {
	r.HandleFunc("/keys", s.Create).Methods("POST")
	r.HandleFunc("/keys/recover", s.Recover).Methods("POST")
	r.HandleFunc("/keys", s.List).Methods("GET")
	r.HandleFunc("/keys/{name}", s.Get).Methods("GET")
	r.HandleFunc("/keys/{name}", s.Update).Methods("POST", "PUT")
	r.HandleFunc("/keys/{name}", s.Delete).Methods("DELETE")

	return nil
}

// RegisterSignTx is a mux.Router handler that
// exposes POST method access to sign a transaction.
func (s *ServiceKeys) RegisterSignTx(r *mux.Router) error {
	r.HandleFunc("/sign", s.SignTx).Methods("POST")
	return nil
}

// End of mux.Router registrars
