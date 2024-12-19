package grpcgateway

import (
	"errors"
	"net/http"
	"strconv"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/stf"
)

var _ http.Handler = &gatewayInterceptor[transaction.Tx]{}

// gatewayInterceptor handles routing grpc-gateway queries to the app manager's query router.
type gatewayInterceptor[T transaction.Tx] struct {
	logger log.Logger
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
func newGatewayInterceptor[T transaction.Tx](logger log.Logger, gateway *runtime.ServeMux, am appmanager.AppManager[T]) (*gatewayInterceptor[T], error) {
	getMapping, err := getHTTPGetAnnotationMapping()
	if err != nil {
		return nil, err
	}
	return &gatewayInterceptor[T]{
		logger:                logger,
		gateway:               gateway,
		customEndpointMapping: getMapping,
		appManager:            am,
	}, nil
}

// ServeHTTP implements the http.Handler interface. This function will attempt to match http requests to the
// interceptors internal mapping of http annotations to query request type names.
// If no match can be made, it falls back to the runtime gateway server mux.
func (g *gatewayInterceptor[T]) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	g.logger.Debug("received grpc-gateway request", "request_uri", request.RequestURI)
	match := matchURL(request.URL, g.customEndpointMapping)
	if match == nil {
		// no match cases fall back to gateway mux.
		g.gateway.ServeHTTP(writer, request)
		return
	}
	g.logger.Debug("matched request", "query_input", match.QueryInputName)
	_, out := runtime.MarshalerForRequest(g.gateway, request)
	var msg gogoproto.Message
	var err error

	switch request.Method {
	case http.MethodPost:
		msg, err = createMessageFromJSON(match, request)
	case http.MethodGet:
		msg, err = createMessage(match)
	default:
		runtime.DefaultHTTPProtoErrorHandler(request.Context(), g.gateway, out, writer, request, status.Error(codes.Unimplemented, "HTTP method must be POST or GET"))
		return
	}
	if err != nil {
		runtime.DefaultHTTPProtoErrorHandler(request.Context(), g.gateway, out, writer, request, err)
		return
	}

	// extract block height header
	var height uint64
	heightStr := request.Header.Get(GRPCBlockHeightHeader)
	if heightStr != "" {
		height, err = strconv.ParseUint(heightStr, 10, 64)
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, "invalid height: %s", heightStr)
			runtime.DefaultHTTPProtoErrorHandler(request.Context(), g.gateway, out, writer, request, err)
			return
		}
	}

	query, err := g.appManager.Query(request.Context(), height, msg)
	if err != nil {
		// if we couldn't find a handler for this request, just fall back to the gateway mux.
		if errors.Is(err, stf.ErrNoHandler) {
			g.gateway.ServeHTTP(writer, request)
		} else {
			// for all other errors, we just return the error.
			runtime.DefaultHTTPProtoErrorHandler(request.Context(), g.gateway, out, writer, request, err)
		}
		return
	}
	// for no errors, we forward the response.
	runtime.ForwardResponseMessage(request.Context(), g.gateway, out, writer, request, query)
}

// getHTTPGetAnnotationMapping returns a mapping of RPC Method HTTP GET annotation to the RPC Handler's Request Input type full name.
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
			serviceDesc := fd.Services().Get(i)
			for j := 0; j < serviceDesc.Methods().Len(); j++ {
				methodDesc := serviceDesc.Methods().Get(j)

				httpAnnotation := proto.GetExtension(methodDesc.Options(), annotations.E_Http)
				if httpAnnotation == nil {
					continue
				}

				httpRule, ok := httpAnnotation.(*annotations.HttpRule)
				if !ok || httpRule == nil {
					continue
				}
				if httpRule.GetGet() == "" {
					continue
				}

				httpGets[httpRule.GetGet()] = string(methodDesc.Input().FullName())
			}
		}
		return true
	})

	return httpGets, nil
}
