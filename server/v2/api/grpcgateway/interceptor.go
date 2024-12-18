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

var _ http.Handler = &GatewayInterceptor[transaction.Tx]{}

type GatewayInterceptor[T transaction.Tx] struct {
	// gateway is the fallback grpc gateway mux handler.
	gateway *runtime.ServeMux

	// customEndpointMapping is a mapping of custom GET options on proto RPC handlers, to the fully qualified method name.
	//
	// example: /cosmos/bank/v1beta1/denoms_metadata -> cosmos.bank.v1beta1.Query.DenomsMetadata
	customEndpointMapping map[string]string

	// appManager is used to route queries through the SDK router.
	appManager appmanager.AppManager[T]
}

func NewGatewayInterceptor[T transaction.Tx](gateway *runtime.ServeMux, am appmanager.AppManager[T]) *GatewayInterceptor[T] {
	return &GatewayInterceptor[T]{
		gateway:               gateway,
		customEndpointMapping: GetProtoHTTPGetRuleMapping(),
		appManager:            am,
	}
}

func (g *GatewayInterceptor[T]) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("printing mapping")
	fmt.Println(g.customEndpointMapping)
	fmt.Println("got request for: ", request.URL.Path)
	uri := request.URL.RequestURI()
	fmt.Println("checking URI: ", uri)
	uriMatch := matchURI(uri, g.customEndpointMapping)
	if uriMatch != nil {
		fmt.Println("got match: ", uriMatch.QueryInputName)
		var msg gogoproto.Message
		var err error

		switch request.Method {
		case http.MethodPost:
			msg, err = createMessageFromJSON(uriMatch, request)
		case http.MethodGet:
			msg, err = createMessage(uriMatch)
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
		fmt.Println("no custom endpoint mapping found, falling back to gateway router")
		g.gateway.ServeHTTP(writer, request)
	}
}

// GetProtoHTTPGetRuleMapping returns a mapping of proto method full name to it's HTTP GET annotation.
func GetProtoHTTPGetRuleMapping() map[string]string {
	protoFiles, err := gogoproto.MergedRegistry()
	if err != nil {
		panic(err)
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
				fmt.Printf("input name: %q \t get option: %q\n", md.Input().FullName(), httpRule.GetGet())
			}
		}
		return true
	})

	return httpGets
}
