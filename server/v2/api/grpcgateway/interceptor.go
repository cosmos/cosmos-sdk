package grpcgateway

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
)

var _ http.Handler = &gatewayInterceptor[transaction.Tx]{}

// queryMetadata holds information related to handling gateway queries.
type queryMetadata struct {
	// queryInputProtoName is the proto name of the query's input type.
	queryInputProtoName string
	// wildcardKeyNames are the wildcard key names from the query's HTTP annotation.
	// for example /foo/bar/{baz}/{qux} would produce []string{"baz", "qux"}
	// this is used for building the query's parameter map.
	wildcardKeyNames []string
}

// gatewayInterceptor handles routing grpc-gateway queries to the app manager's query router.
type gatewayInterceptor[T transaction.Tx] struct {
	logger log.Logger
	// gateway is the fallback grpc gateway mux handler.
	gateway *runtime.ServeMux

	// regexpToQueryMetadata is a mapping of regular expressions of HTTP annotations to metadata for the query.
	// it is built from parsing the HTTP annotations obtained from the gogoproto global registry.'
	//
	// TODO: it might be interesting to make this a 'most frequently used' data structure, so frequently used regexp's are
	// iterated over first.
	regexpToQueryMetadata map[*regexp.Regexp]queryMetadata

	// appManager is used to route queries to the application.
	appManager appmanager.AppManager[T]
}

// newGatewayInterceptor creates a new gatewayInterceptor.
func newGatewayInterceptor[T transaction.Tx](logger log.Logger, gateway *runtime.ServeMux, am appmanager.AppManager[T]) (*gatewayInterceptor[T], error) {
	getMapping, err := getHTTPGetAnnotationMapping()
	if err != nil {
		return nil, err
	}
	// convert the mapping to regular expressions for URL matching.
	regexQueryMD := createRegexMapping(logger, getMapping)
	if err != nil {
		return nil, err
	}
	return &gatewayInterceptor[T]{
		logger:                logger,
		gateway:               gateway,
		regexpToQueryMetadata: regexQueryMD,
		appManager:            am,
	}, nil
}

// ServeHTTP implements the http.Handler interface. This method will attempt to match request URIs to its internal mapping
// of gateway HTTP annotations. If no match can be made, it falls back to the runtime gateway server mux.
func (g *gatewayInterceptor[T]) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	g.logger.Debug("received grpc-gateway request", "request_uri", request.RequestURI)
	match := matchURL(request.URL, g.regexpToQueryMetadata)
	if match == nil {
		// no match cases fall back to gateway mux.
		g.gateway.ServeHTTP(writer, request)
		return
	}

	g.logger.Debug("matched request", "query_input", match.QueryInputName)

	in, out := runtime.MarshalerForRequest(g.gateway, request)

	// extract the proto message type.
	msgType := gogoproto.MessageType(match.QueryInputName)
	msg, ok := reflect.New(msgType.Elem()).Interface().(gogoproto.Message)
	if !ok {
		runtime.DefaultHTTPProtoErrorHandler(request.Context(), g.gateway, out, writer, request, status.Errorf(codes.Internal, "unable to to create gogoproto message from query input name %s", match.QueryInputName))
		return
	}

	// msg population based on http method.
	var inputMsg gogoproto.Message
	var err error
	switch request.Method {
	case http.MethodGet:
		inputMsg, err = g.createMessageFromGetRequest(request.Context(), in, request, msg, match.Params)
	case http.MethodPost:
		inputMsg, err = g.createMessageFromPostRequest(request.Context(), in, request, msg)
	default:
		runtime.DefaultHTTPProtoErrorHandler(request.Context(), g.gateway, out, writer, request, status.Error(codes.InvalidArgument, "HTTP method was not POST or GET"))
	}

	// get the height from the header.
	var height uint64
	heightStr := request.Header.Get(GRPCBlockHeightHeader)
	if heightStr != "" {
		if height, err = strconv.ParseUint(heightStr, 10, 64); err != nil {
			runtime.DefaultHTTPProtoErrorHandler(request.Context(), g.gateway, out, writer, request, status.Errorf(codes.InvalidArgument, "invalid height in header: %s", heightStr))
			return
		}
	}

	responseMsg, err := g.appManager.Query(request.Context(), height, inputMsg)
	if err != nil {
		// if we couldn't find a handler for this request, just fall back to the gateway mux.
		if strings.Contains(err.Error(), "no handler") {
			g.gateway.ServeHTTP(writer, request)
		} else {
			// for all other errors, we just return the error.
			runtime.DefaultHTTPProtoErrorHandler(request.Context(), g.gateway, out, writer, request, err)
		}
		return
	}

	// for no errors, we forward the response.
	runtime.ForwardResponseMessage(request.Context(), g.gateway, out, writer, request, responseMsg)
}

func (g *gatewayInterceptor[T]) createMessageFromPostRequest(_ context.Context, marshaler runtime.Marshaler, req *http.Request, input gogoproto.Message) (gogoproto.Message, error) {
	newReader, err := utilities.IOReaderFactory(req.Body)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	if err = marshaler.NewDecoder(newReader()).Decode(input); err != nil && err != io.EOF {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	return input, nil
}

func (g *gatewayInterceptor[T]) createMessageFromGetRequest(_ context.Context, _ runtime.Marshaler, req *http.Request, input gogoproto.Message, wildcardValues map[string]string) (gogoproto.Message, error) {
	// decode the path wildcards into the message.
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           input,
		TagName:          "json",
		WeaklyTypedInput: true,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create message decoder")
	}
	if err := decoder.Decode(wildcardValues); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = req.ParseForm(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	// im not really sure what this filter is for, but defaulting it like so doesn't seem to break anything.
	// pb.gw.go code uses it, but im not sure why.
	filter := &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
	err = runtime.PopulateQueryParameters(input, req.Form, filter)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	return input, err
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

// createRegexMapping converts the annotationMapping (HTTP annotation -> query input type name) to a
// map of regular expressions for that HTTP annotation pattern, to queryMetadata.
func createRegexMapping(logger log.Logger, annotationMapping map[string]string) map[*regexp.Regexp]queryMetadata {
	regexQueryMD := make(map[*regexp.Regexp]queryMetadata)
	seenPatterns := make(map[string]string)
	for annotation, queryInputName := range annotationMapping {
		pattern, wildcardNames := patternToRegex(annotation)
		reg := regexp.MustCompile(pattern)
		if otherAnnotation, ok := seenPatterns[pattern]; !ok {
			seenPatterns[pattern] = annotation
		} else {
			// TODO: eventually we want this to error, but there is currently a duplicate in the protobuf.
			// see: https://github.com/cosmos/cosmos-sdk/issues/23281
			logger.Warn("duplicate HTTP annotation found %q and %q. query will resolve to %q", annotation, otherAnnotation, queryInputName)
		}
		regexQueryMD[reg] = queryMetadata{
			queryInputProtoName: queryInputName,
			wildcardKeyNames:    wildcardNames,
		}
	}
	return regexQueryMD
}
