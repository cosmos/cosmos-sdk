package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/common"

	sdk "github.com/cosmos/cosmos-sdk"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// ServiceTxs exposes a REST API service for sendings txs.
// It wraps a Tendermint RPC client.
type ServiceTxs struct {
	node rpcclient.Client
}

func NewServiceTxs(c rpcclient.Client) *ServiceTxs {
	return &ServiceTxs{
		node: c,
	}
}

func (s *ServiceTxs) PostTx(w http.ResponseWriter, r *http.Request) {
	tx := new(sdk.Tx)
	if err := common.ParseRequestAndValidateJSON(r, tx); err != nil {
		common.WriteError(w, err)
		return
	}

	packet := wire.BinaryBytes(*tx)
	commit, err := s.node.BroadcastTxCommit(packet)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	common.WriteSuccess(w, commit)
}

// mux.Router registrars

// RegisterPostTx is a mux.Router handler that exposes POST
// method access to post a transaction to the blockchain.
func (s *ServiceTxs) RegisterPostTx(r *mux.Router) error {
	r.HandleFunc("/tx", s.PostTx).Methods("POST")
	return nil
}

// End of mux.Router registrars
