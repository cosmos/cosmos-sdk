package grpcgateway

import (
	"errors"
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

const MaxBodySize = 1 << 20 // 1 MB

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

	matcher uriMatcher

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
	wildcardMatchers, simpleMatchers := createRegexMapping(logger, getMapping)
	matcher := uriMatcher{
		wildcardURIMatchers: wildcardMatchers,
		simpleMatchers:      simpleMatchers,
	}
	return &gatewayInterceptor[T]{
		logger:     logger,
		gateway:    gateway,
		matcher:    matcher,
		appManager: am,
	}, nil
}

// ServeHTTP implements the http.Handler interface. This method will attempt to match request URIs to its internal mapping
// of gateway HTTP annotations. If no match can be made, it falls back to the runtime gateway server mux.
func (g *gatewayInterceptor[T]) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	g.logger.Debug("received grpc-gateway request", "request_uri", request.RequestURI)
	match := g.matcher.matchURL(request.URL)
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
		inputMsg, err = g.createMessageFromGetRequest(request, msg, match.Params)
	case http.MethodPost:
		inputMsg, err = g.createMessageFromPostRequest(in, request, msg)
	default:
		runtime.DefaultHTTPProtoErrorHandler(request.Context(), g.gateway, out, writer, request, status.Error(codes.InvalidArgument, "HTTP method was not POST or GET"))
		return
	}
	if err != nil {
		// the errors returned from the message creation methods return status errors. no need to make one here.
		runtime.DefaultHTTPProtoErrorHandler(request.Context(), g.gateway, out, writer, request, err)
		return
	}

	// get the height from the header.
	var height uint64
	heightStr := request.Header.Get(GRPCBlockHeightHeader)
	heightStr = strings.Trim(heightStr, `\"`)
	if heightStr != "" && heightStr != "latest" {
		height, err = strconv.ParseUint(heightStr, 10, 64)
		if err != nil {
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

func (g *gatewayInterceptor[T]) createMessageFromPostRequest(marshaler runtime.Marshaler, req *http.Request, input gogoproto.Message) (gogoproto.Message, error) {
	if req.ContentLength > MaxBodySize {
		return nil, status.Errorf(codes.InvalidArgument, "request body too large: %d bytes, max=%d", req.ContentLength, MaxBodySize)
	}
	newReader, err := utilities.IOReaderFactory(req.Body)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	if err = marshaler.NewDecoder(newReader()).Decode(input); err != nil && !errors.Is(err, io.EOF) {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	return input, nil
}

func (g *gatewayInterceptor[T]) createMessageFromGetRequest(req *http.Request, input gogoproto.Message, wildcardValues map[string]string) (gogoproto.Message, error) {
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

	filter := filterFromPathParams(wildcardValues)
	err = runtime.PopulateQueryParameters(input, req.Form, filter)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	return input, err
}

func filterFromPathParams(pathParams map[string]string) *utilities.DoubleArray {
	var prefixPaths [][]string

	for k := range pathParams {
		prefixPaths = append(prefixPaths, []string{k})
	}

	// Pass these to NewDoubleArray using the "spread" (...) syntax.
	return utilities.NewDoubleArray(prefixPaths)
}

// getHTTPGetAnnotationMapping returns a mapping of RPC Method HTTP GET annotation to the RPC Handler's Request Input type full name.
//
// example: "/cosmos/auth/v1beta1/account_info/{address}":"cosmos.auth.v1beta1.Query.AccountInfo"
func getHTTPGetAnnotationMapping() (map[string]string, error) {
	protoFiles, err := gogoproto.MergedRegistry()
	if err != nil {
		return nil, err
	}

	annotationToQueryInputName := make(map[string]string)
	protoFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		for i := 0; i < fd.Services().Len(); i++ {
			serviceDesc := fd.Services().Get(i)
			for j := 0; j < serviceDesc.Methods().Len(); j++ {
				methodDesc := serviceDesc.Methods().Get(j)
				httpExtension := proto.GetExtension(methodDesc.Options(), annotations.E_Http)
				if httpExtension == nil {
					continue
				}

				httpRule, ok := httpExtension.(*annotations.HttpRule)
				if !ok || httpRule == nil {
					continue
				}
				queryInputName := string(methodDesc.Input().FullName())
				annotations := append(httpRule.GetAdditionalBindings(), httpRule)
				for _, a := range annotations {
					if httpAnnotation := a.GetGet(); httpAnnotation != "" {
						annotationToQueryInputName[httpAnnotation] = queryInputName
					}
					if httpAnnotation := a.GetPost(); httpAnnotation != "" {
						annotationToQueryInputName[httpAnnotation] = queryInputName
					}
				}
			}
		}
		return true
	})
	return annotationToQueryInputName, nil
}

// createRegexMapping converts the annotationMapping (HTTP annotation -> query input type name) to a
// map of regular expressions for that HTTP annotation pattern, to queryMetadata.
func createRegexMapping(logger log.Logger, annotationMapping map[string]string) (map[*regexp.Regexp]queryMetadata, map[string]queryMetadata) {
	wildcardMatchers := make(map[*regexp.Regexp]queryMetadata)
	// seen patterns is a map of URI patterns to annotations. for simple queries (no wildcards) the annotation is used
	// for the key.
	seenPatterns := make(map[string]string)
	simpleMatchers := make(map[string]queryMetadata)

	for annotation, queryInputName := range annotationMapping {
		pattern, wildcardNames := patternToRegex(annotation)
		if len(wildcardNames) == 0 {
			if otherAnnotation, ok := seenPatterns[annotation]; ok {
				// TODO: eventually we want this to error, but there is currently a duplicate in the protobuf.
				// see: https://github.com/cosmos/cosmos-sdk/issues/23281
				logger.Warn("duplicate HTTP annotation found", "annotation1", annotation, "annotation2", otherAnnotation, "query_input_name", queryInputName)
			}
			simpleMatchers[annotation] = queryMetadata{
				queryInputProtoName: queryInputName,
				wildcardKeyNames:    nil,
			}
			seenPatterns[annotation] = annotation
		} else {
			reg := regexp.MustCompile(pattern)
			if otherAnnotation, ok := seenPatterns[pattern]; ok {
				// TODO: eventually we want this to error, but there is currently a duplicate in the protobuf.
				// see: https://github.com/cosmos/cosmos-sdk/issues/23281
				logger.Warn("duplicate HTTP annotation found", "annotation1", annotation, "annotation2", otherAnnotation, "query_input_name", queryInputName)
			}
			wildcardMatchers[reg] = queryMetadata{
				queryInputProtoName: queryInputName,
				wildcardKeyNames:    wildcardNames,
			}
			seenPatterns[pattern] = annotation

		}
	}
	return wildcardMatchers, simpleMatchers
}
