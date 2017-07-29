package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/client/commands"
	"github.com/tendermint/basecoin/client/commands/proofs"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/fee"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/stack"
	keysutils "github.com/tendermint/go-crypto/cmd"
	keys "github.com/tendermint/go-crypto/keys"
	lightclient "github.com/tendermint/light-client"
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
	if err := parseRequestJSON(r, ckReq); err != nil {
		writeError(w, err)
		return
	}

	key, seed, err := k.manager.Create(ckReq.Name, ckReq.Passphrase, ckReq.Algo)
	if err != nil {
		writeError(w, err)
		return
	}

	res := &CreateKeyResponse{Key: key, Seed: seed}
	writeSuccess(w, res)
}

func (k *Keys) GetKey(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)
	name := query["name"]
	key, err := k.manager.Get(name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, &key)
}

func (k *Keys) ListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := k.manager.List()
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, keys)
}

var (
	errNonMatchingPathAndJSONKeyNames = errors.New("path and json key names don't match")
)

func (k *Keys) UpdateKey(w http.ResponseWriter, r *http.Request) {
	uReq := new(UpdateKeyRequest)
	if err := parseRequestJSON(r, uReq); err != nil {
		writeError(w, err)
		return
	}

	query := mux.Vars(r)
	name := query["name"]
	if name != uReq.Name {
		writeError(w, errNonMatchingPathAndJSONKeyNames)
		return
	}

	if err := k.manager.Update(uReq.Name, uReq.OldPass, uReq.NewPass); err != nil {
		writeError(w, err)
		return
	}

	key, err := k.manager.Get(uReq.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, &key)
}

func (k *Keys) DeleteKey(w http.ResponseWriter, r *http.Request) {
	dReq := new(DeleteKeyRequest)
	if err := parseRequestJSON(r, dReq); err != nil {
		writeError(w, err)
		return
	}

	query := mux.Vars(r)
	name := query["name"]
	if name != dReq.Name {
		writeError(w, errNonMatchingPathAndJSONKeyNames)
		return
	}

	if err := k.manager.Delete(dReq.Name, dReq.Passphrase); err != nil {
		writeError(w, err)
		return
	}

	resp := &ErrorResponse{Success: true}
	writeSuccess(w, resp)
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
	r.HandleFunc("/build/send", doSend).Methods("POST")
	r.HandleFunc("/sign", doSign).Methods("POST")
	r.HandleFunc("/tx", doPostTx).Methods("POST")
	r.HandleFunc("/query/account/{signature}", doAccountQuery).Methods("GET")

	return nil
}

func extractAddress(signature string) (address string, err *ErrorResponse) {
	// Expecting the signature of the form:
	//  sig:<ADDRESS>
	splits := strings.Split(signature, ":")
	if len(splits) < 2 {
		return "", &ErrorResponse{
			Error: `expecting the signature of the form "sig:<ADDRESS>"`,
			Code:  406,
		}
	}
	if splits[0] != "sigs" {
		return "", &ErrorResponse{
			Error: `expecting the signature of the form "sig:<ADDRESS>"`,
			Code:  406,
		}
	}
	return splits[1], nil
}

func doAccountQuery(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)
	signature := query["signature"]
	address, errResp := extractAddress(signature)
	if errResp != nil {
		writeCode(w, errResp, errResp.Code)
		return
	}
	actor, err := commands.ParseActor(address)
	if err != nil {
		writeError(w, err)
		return
	}
	actor = coin.ChainAddr(actor)
	key := stack.PrefixedKey(coin.NameCoin, actor.Bytes())
	account := new(coin.Account)
	proof, err := proofs.GetAndParseAppProof(key, account)
	if lightclient.IsNoDataErr(err) {
		err := fmt.Errorf("account bytes are empty for address: %q", address)
		writeError(w, err)
		return
	} else if err != nil {
		writeError(w, err)
		return
	}

	if err := proofs.OutputProof(account, proof.BlockHeight()); err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, account)
}

func doPostTx(w http.ResponseWriter, r *http.Request) {
	tx := new(basecoin.Tx)
	if err := parseRequestJSON(r, tx); err != nil {
		writeError(w, err)
		return
	}
	commit, err := PostTx(*tx)
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, commit)
}

func doSign(w http.ResponseWriter, r *http.Request) {
	sr := new(SignRequest)
	if err := parseRequestJSON(r, sr); err != nil {
		writeError(w, err)
		return
	}

	tx := sr.Tx
	if err := SignTx(sr.Name, sr.Password, tx); err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, tx)
}

func doSend(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	si := new(SendInput)
	if err := parseRequestJSON(r, si); err != nil {
		writeError(w, err)
		return
	}

	var errsList []string
	if si.From == nil {
		errsList = append(errsList, `"from" cannot be nil`)
	}
	if si.Sequence <= 0 {
		errsList = append(errsList, `"sequence" must be > 0`)
	}
	if si.To == nil {
		errsList = append(errsList, `"to" cannot be nil`)
	}
	if si.Fees == nil {
		errsList = append(errsList, `"fees" cannot be nil`)
	}
	if len(errsList) > 0 {
		err := &ErrorResponse{
			Error: strings.Join(errsList, ", "),
			Code:  406,
		}
		writeCode(w, err, 406)
		return
	}

	coins := []coin.Coin{*si.Fees}
	in := []coin.TxInput{
		coin.NewTxInput(*si.From, coins),
	}
	out := []coin.TxOutput{
		coin.NewTxOutput(*si.To, coins),
	}

	tx := coin.NewSendTx(in, out)
	tx = fee.NewFee(tx, *si.Fees, *si.From)

	signers := []basecoin.Actor{
		*si.From,
		*si.To,
	}
	tx = nonce.NewTx(si.Sequence, signers, tx)
	tx = base.NewChainTx(commands.GetChainID(), 0, tx)

	if si.Multi {
		tx = auth.NewMulti(tx).Wrap()
	} else {
		tx = auth.NewSig(tx).Wrap()
	}
}
