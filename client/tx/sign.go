package tx

import (
	"encoding/json"
	"net/http"

	keybase "github.com/cosmos/cosmos-sdk/client/keys"
	keys "github.com/tendermint/go-crypto/keys"
)

// REST request body
// TODO does this need to be exposed?
type SignTxBody struct {
	Name     string `json="name"`
	Password string `json="password"`
	TxBytes  string `json="tx"`
}

// sign transaction REST Handler
func SignTxRequstHandler(w http.ResponseWriter, r *http.Request) {
	var kb keys.Keybase
	var m SignTxBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&m)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	kb, err = keybase.GetKeyBase()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	//TODO check if account exists
	sig, _, err := kb.Sign(m.Name, m.Password, []byte(m.TxBytes))
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(sig.Bytes())
}
