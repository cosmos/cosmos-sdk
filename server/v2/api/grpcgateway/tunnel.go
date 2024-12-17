package grpcgateway

import (
	"net/http"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager"
)

var _ http.Handler = &Tunnel[transaction.Tx]{}

type Tunnel[T transaction.Tx] struct {
	// gateway is the fallback grpc gateway mux handler.
	gateway *runtime.ServeMux

	// customEndpointMapping is a mapping of custom GET options on proto RPC handlers, to the fully qualified method name.
	//
	// example: /cosmos/bank/v1beta1/denoms_metadata -> cosmos.bank.v1beta1.Query.DenomsMetadata
	customEndpointMapping map[string]string

	appManager appmanager.AppManager[T]
}

func NewTunnel[T transaction.Tx](gateway *runtime.ServeMux) *Tunnel[T] {
	return &Tunnel[T]{gateway: gateway}
}

// Handle some things:
//
// - see if we can match the request to
func (t *Tunnel[T]) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	uri := request.URL.RequestURI()
	uriMatch := matchURI(uri, t.customEndpointMapping)
	if uriMatch != nil {
		if uriMatch.HasParams() {

		} else {

		}
	} else {
		t.gateway.ServeHTTP(writer, request)
	}
}

func createMessage(method string) (gogoproto.Message, error) {
	return nil, nil
}
