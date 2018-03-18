package tx

import (
	"encoding/json"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/builder"
)

type BroadcastTxBody struct {
	TxBytes string `json="tx"`
}

func BroadcastTxRequestHandler(w http.ResponseWriter, r *http.Request) {
	var m BroadcastTxBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&m)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	res, err := builder.BroadcastTx([]byte(m.TxBytes))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte(string(res.Height)))
}
