package grpcgateway

import (
	"encoding/json"
	"fmt"
	"net/http"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/reflect/protoreflect"

	"google.golang.org/protobuf/proto"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager"
)

var _ http.Handler = &gatewayInterceptor[transaction.Tx]{}

// gatewayInterceptor handles routing grpc-gateway queries to the app manager's query router.
type gatewayInterceptor[T transaction.Tx] struct {
	// gateway is the fallback grpc gateway mux handler.
	gateway *runtime.ServeMux

	// customEndpointMapping is a mapping of custom GET options on proto RPC handlers, to the fully qualified method name.
	//
	// example: /cosmos/bank/v1beta1/denoms_metadata -> cosmos.bank.v1beta1.Query.DenomsMetadata
	customEndpointMapping map[string]string

	// appManager is used to route queries to the application.
	appManager appmanager.AppManager[T]
}

// newGatewayInterceptor creates a new gatewayInterceptor.
func newGatewayInterceptor[T transaction.Tx](gateway *runtime.ServeMux, am appmanager.AppManager[T]) (*gatewayInterceptor[T], error) {
	getMapping, err := getHTTPGetAnnotationMapping()
	if err != nil {
		return nil, err
	}
	return &gatewayInterceptor[T]{
		gateway:               gateway,
		customEndpointMapping: getMapping,
		appManager:            am,
	}, nil
}

// ServeHTTP implements the http.Handler interface. This function will attempt to match http requests to the
// interceptors internal mapping of http annotations to query request type names.
// If no match can be made, it falls back to the runtime gateway server mux.
func (g *gatewayInterceptor[T]) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	match := matchURL(request.URL, g.customEndpointMapping)
	if match != nil {
		var msg gogoproto.Message
		var err error

		switch request.Method {
		case http.MethodPost:
			msg, err = createMessageFromJSON(match, request)
		case http.MethodGet:
			msg, err = createMessage(match)
		default:
			http.Error(writer, "unsupported http method", http.StatusMethodNotAllowed)
			return
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		query, err := g.appManager.Query(request.Context(), 0, msg)
		if err != nil {
			http.Error(writer, "Error querying", http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(query); err != nil {
			http.Error(writer, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		}
	} else {
		g.gateway.ServeHTTP(writer, request)
	}
}

// getHTTPGetAnnotationMapping returns a mapping of proto query input type full name to its RPC method's HTTP GET annotation.
//
// example: "/cosmos/auth/v1beta1/account_info/{address}":"cosmos.auth.v1beta1.Query.AccountInfo"
func getHTTPGetAnnotationMapping() (map[string]string, error) {
	protoFiles, err := gogoproto.MergedRegistry()
	if err != nil {
		return nil, err
	}

	httpGets := make(map[string]string)
	protoFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Services().Len(); i++ {
			// Get the service descriptor
			sd := fd.Services().Get(i)

			for j := 0; j < sd.Methods().Len(); j++ {
				// Get the method descriptor
				md := sd.Methods().Get(j)

				httpOption := proto.GetExtension(md.Options(), annotations.E_Http)
				if httpOption == nil {
					continue
				}

				httpRule, ok := httpOption.(*annotations.HttpRule)
				if !ok || httpRule == nil {
					continue
				}
				if httpRule.GetGet() == "" {
					continue
				}

				httpGets[httpRule.GetGet()] = string(md.Input().FullName())
			}
		}
		return true
	})

	return httpGets, nil
}
