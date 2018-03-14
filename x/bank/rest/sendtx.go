package rest

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/builder"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank/commands"
)

type SendBody struct {
	// fees is not used currently
	// Fees             sdk.Coin  `json="fees"`
	Amount           sdk.Coins `json:"amount"`
	LocalAccountName string    `json:"name"`
	Password         string    `json:"password"`
	ChainID          string    `json:"chain_id"`
	Sequence         string    `json:"sequence"`
}

func SendRequestHandler(cdc *wire.Codec) func(http.ResponseWriter, *http.Request) {
	c := commands.Commander{cdc}
	return func(w http.ResponseWriter, r *http.Request) {
		// collect data
		vars := mux.Vars(r)
		address := vars["address"]

		var m SendBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = json.Unmarshal(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		kb, err := keys.GetKeyBase()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		info, err := kb.Get(m.LocalAccountName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		bz, err := hex.DecodeString(address)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		to := sdk.Address(bz)

		// build message
		msg := commands.BuildMsg(info.PubKey.Address(), to, m.Amount)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		signMsg := sdk.StdSignMsg{
			ChainID:   m.ChainID,
			Sequences: []int64{m.Sequence},
			Msg:       msg,
		}

		// sign
		txBytes, err := builder.SignAndBuild(m.LocalAccountName, m.Password, signMsg, c.Cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// send
		res, err := client.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		output, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
